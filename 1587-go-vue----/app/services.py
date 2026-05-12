from sqlalchemy.orm import Session
from app.models import (
    Vehicle, Charger, ChargingSession, Alert,
    VehicleStatus, ChargerStatus, AssignmentStatus,
    AlertType, NotificationType
)
from app.crud import (
    CRUDVehicle, CRUDCharger, CRUDChargingSession,
    CRUDAlert, CRUDDispatchLog, CRUDAssignment
)
from typing import Optional, Tuple
from datetime import datetime, timedelta
import math

LOW_BATTERY_THRESHOLD = 15.0
AUTO_CHARGE_THRESHOLD = 30.0
TARGET_CHARGE = 90.0
BATTERY_CONSUMPTION_PER_KM = 2.0

class ChargingService:
    @staticmethod
    def calculate_charging_time(
        start_battery: float,
        target_battery: float,
        battery_capacity_kwh: float,
        charger_power_kw: float
    ) -> timedelta:
        energy_needed = (target_battery - start_battery) / 100.0 * battery_capacity_kwh
        hours_needed = energy_needed / charger_power_kw
        return timedelta(hours=hours_needed)

    @staticmethod
    def find_available_charger(db: Session, station_id: int) -> Optional[Charger]:
        available_chargers = CRUDCharger.get_idle_by_station(db, station_id)
        if not available_chargers:
            return None
        return available_chargers[0]

    @staticmethod
    def try_allocate_charger(
        db: Session,
        vehicle: Vehicle
    ) -> Tuple[Optional[Charger], Optional[str]]:
        if vehicle.current_station_id is None:
            return None, "车辆未在站点，无法分配充电桩"

        if vehicle.status not in [VehicleStatus.IDLE, VehicleStatus.RUNNING]:
            return None, f"车辆状态为 {vehicle.status}，无法分配充电"

        selected_charger = None
        attempt_count = 0

        while attempt_count < 5:
            candidate_charger = ChargingService.find_available_charger(db, vehicle.current_station_id)
            if not candidate_charger:
                return None, "该站点无可用充电桩"

            if not CRUDVehicle.lock(db, vehicle):
                return None, "车辆锁定失败，可能已被占用"

            if CRUDCharger.lock(db, candidate_charger, vehicle.id):
                db.refresh(candidate_charger)
                if candidate_charger.status == ChargerStatus.LOCKED and candidate_charger.locked_by_vehicle_id == vehicle.id:
                    selected_charger = candidate_charger
                    break
                else:
                    CRUDVehicle.release(db, vehicle)
            else:
                CRUDVehicle.release(db, vehicle)

            attempt_count += 1

        if selected_charger is None:
            return None, "多次尝试分配充电桩失败，请稍后重试"

        return selected_charger, None

    @staticmethod
    def start_charging(
        db: Session,
        vehicle: Vehicle,
        charger: Charger
    ) -> ChargingSession:
        charging_time = ChargingService.calculate_charging_time(
            vehicle.current_battery,
            TARGET_CHARGE,
            vehicle.max_battery_capacity_kwh,
            charger.power_kw
        )
        estimated_end = datetime.utcnow() + charging_time

        CRUDCharger.set_charging(db, charger)
        CRUDVehicle.set_charging(db, vehicle)
        CRUDCharger.increment_charging_count(db, charger)

        session = CRUDChargingSession.create(
            db,
            vehicle.id,
            charger.id,
            vehicle.current_battery,
            TARGET_CHARGE,
            estimated_end
        )

        CRUDDispatchLog.create(
            db,
            f"车辆 {vehicle.plate_number} 开始在充电桩 {charger.name} 充电",
            "INFO",
            vehicle.id,
            charger.id
        )

        return session

    @staticmethod
    def process_vehicle_arrival(db: Session, vehicle: Vehicle) -> dict:
        result = {
            "vehicle_id": vehicle.id,
            "plate_number": vehicle.plate_number,
            "battery_level": vehicle.current_battery,
            "action_taken": None,
            "charging_session": None,
            "message": None
        }

        if vehicle.current_battery < AUTO_CHARGE_THRESHOLD:
            result["message"] = f"电量 {vehicle.current_battery:.1f}% 低于 {AUTO_CHARGE_THRESHOLD}%，尝试分配充电"

            charger, error = ChargingService.try_allocate_charger(db, vehicle)
            if charger:
                session = ChargingService.start_charging(db, vehicle, charger)
                result["action_taken"] = "charging_allocated"
                result["charging_session"] = {
                    "session_id": session.id,
                    "charger_name": charger.name,
                    "estimated_end": session.estimated_end_time
                }
            else:
                result["action_taken"] = "allocation_failed"
                result["message"] = error
        else:
            result["message"] = f"电量 {vehicle.current_battery:.1f}% 充足，无需充电"

        return result

    @staticmethod
    def update_charging_progress(db: Session, session: ChargingSession) -> dict:
        now = datetime.utcnow()
        start_time = session.start_time
        estimated_end = session.estimated_end_time

        if estimated_end is None:
            return {"progress": 0, "current_battery": session.current_battery}

        total_duration = (estimated_end - start_time).total_seconds()
        if total_duration <= 0:
            return {"progress": 100, "current_battery": session.target_battery}

        elapsed = (now - start_time).total_seconds()
        progress_ratio = min(elapsed / total_duration, 1.0)

        battery_increase = (session.target_battery - session.start_battery) * progress_ratio
        current_battery = round(session.start_battery + battery_increase, 1)

        CRUDChargingSession.update_progress(db, session, current_battery)

        vehicle = CRUDVehicle.get_by_id(db, session.vehicle_id)
        if vehicle:
            CRUDVehicle.update_battery(db, vehicle, current_battery)

        result = {
            "session_id": session.id,
            "progress": round(progress_ratio * 100, 1),
            "current_battery": current_battery,
            "target_battery": session.target_battery,
            "is_complete": progress_ratio >= 1.0
        }

        if progress_ratio >= 1.0:
            ChargingService.complete_charging(db, session)
            result["action"] = "charging_completed"

        return result

    @staticmethod
    def complete_charging(db: Session, session: ChargingSession):
        CRUDChargingSession.complete(db, session, session.target_battery)

        vehicle = CRUDVehicle.get_by_id(db, session.vehicle_id)
        charger = CRUDCharger.get_by_id(db, session.charger_id)

        if vehicle:
            vehicle.current_battery = session.target_battery
            vehicle.status = VehicleStatus.IDLE
            db.commit()

        if charger:
            CRUDCharger.release(db, charger)

        CRUDDispatchLog.create(
            db,
            f"车辆 {vehicle.plate_number if vehicle else '未知'} 充电完成",
            "INFO",
            session.vehicle_id,
            session.charger_id
        )

    @staticmethod
    def handle_charger_fault(db: Session, charger: Charger):
        CRUDCharger.set_faulty(db, charger)

        active_session = CRUDChargingSession.get_active_by_charger(db, charger.id)
        if active_session:
            CRUDChargingSession.cancel(db, active_session)
            vehicle = CRUDVehicle.get_by_id(db, active_session.vehicle_id)
            if vehicle:
                vehicle.status = VehicleStatus.IDLE
                db.commit()

        if charger.station_id is not None:
            notification_type = NotificationType.TERMINAL
            message = f"充电桩 {charger.name} 发生故障，请调度员安排车辆到其他站点充电"
        else:
            notification_type = NotificationType.LOG_ONLY
            message = f"充电桩 {charger.name} 发生故障（未配置站点信息），请手动处理"

        CRUDAlert.create(
            db,
            AlertType.CHARGER_FAULT,
            message,
            charger_id=charger.id,
            notification_type=notification_type
        )

        CRUDDispatchLog.create(
            db,
            message,
            "ERROR",
            charger_id=charger.id
        )

class VehicleReportService:
    @staticmethod
    def check_low_battery(db: Session, vehicle: Vehicle, battery_level: float) -> Optional[Alert]:
        if battery_level < LOW_BATTERY_THRESHOLD:
            existing_alert = db.query(Alert).filter(
                Alert.vehicle_id == vehicle.id,
                Alert.type == AlertType.LOW_BATTERY,
                Alert.status.in_(["active", "acknowledged"])
            ).first()

            if not existing_alert:
                alert = CRUDAlert.create(
                    db,
                    AlertType.LOW_BATTERY,
                    f"车辆 {vehicle.plate_number} 电量低于 {LOW_BATTERY_THRESHOLD}%，当前电量：{battery_level:.1f}%",
                    vehicle_id=vehicle.id,
                    notification_type=NotificationType.TERMINAL
                )
                CRUDDispatchLog.create(
                    db,
                    f"低电量告警：车辆 {vehicle.plate_number} 电量 {battery_level:.1f}%",
                    "WARNING",
                    vehicle.id
                )
                return alert
        return None

    @staticmethod
    def report_vehicle_status(
        db: Session,
        vehicle: Vehicle,
        latitude: float,
        longitude: float,
        current_battery: float
    ) -> dict:
        CRUDVehicle.update_location(db, vehicle, latitude, longitude)
        CRUDVehicle.update_battery(db, vehicle, current_battery)

        result = {
            "vehicle_id": vehicle.id,
            "plate_number": vehicle.plate_number,
            "battery_level": current_battery,
            "low_battery_alert": None,
            "actions": []
        }

        alert = VehicleReportService.check_low_battery(db, vehicle, current_battery)
        if alert:
            result["low_battery_alert"] = {
                "alert_id": alert.id,
                "message": alert.message
            }
            result["actions"].append("low_battery_alert_created")

        return result

class DispatchService:
    @staticmethod
    def estimate_remaining_range_km(vehicle: Vehicle) -> float:
        remaining_kwh = (vehicle.current_battery / 100.0) * vehicle.max_battery_capacity_kwh
        return remaining_kwh / BATTERY_CONSUMPTION_PER_KM

    @staticmethod
    def can_complete_current_assignment(db: Session, vehicle: Vehicle) -> bool:
        if vehicle.current_route_id is None:
            return True

        from app.crud import CRUDRoute
        route = CRUDRoute.get_by_id(db, vehicle.current_route_id)
        if not route:
            return True

        remaining_range = DispatchService.estimate_remaining_range_km(vehicle)
        return remaining_range >= route.total_distance_km

    @staticmethod
    def find_reserve_vehicle(db: Session) -> Optional[Vehicle]:
        reserve_vehicles = CRUDVehicle.get_reserve_vehicles(db)
        for v in reserve_vehicles:
            if v.current_battery >= 50.0:
                return v
        return None

    @staticmethod
    def check_and_dispatch(db: Session, vehicle: Vehicle) -> dict:
        result = {
            "vehicle_id": vehicle.id,
            "plate_number": vehicle.plate_number,
            "battery_level": vehicle.current_battery,
            "needs_intervention": False,
            "intervention_type": None,
            "reserve_vehicle": None,
            "message": None
        }

        if vehicle.status not in [VehicleStatus.RUNNING, VehicleStatus.IDLE]:
            result["message"] = f"车辆状态为 {vehicle.status}，无需调度"
            return result

        can_complete = DispatchService.can_complete_current_assignment(db, vehicle)
        if can_complete:
            result["message"] = "电量充足，可完成当前交路"
            return result

        result["needs_intervention"] = True
        result["intervention_type"] = "charge_and_replace"
        result["message"] = "电量不足，无法完成当前交路，安排充电并调度备用车"

        if vehicle.current_station_id is not None:
            charger, error = ChargingService.try_allocate_charger(db, vehicle)
            if charger:
                ChargingService.start_charging(db, vehicle, charger)
                result["message"] += f"，已分配充电桩 {charger.name}"

        reserve = DispatchService.find_reserve_vehicle(db)
        if reserve:
            reserve.status = VehicleStatus.RUNNING
            if vehicle.current_route_id:
                reserve.current_route_id = vehicle.current_route_id
            db.commit()

            CRUDAssignment.create(
                db,
                reserve.id,
                route_id=vehicle.current_route_id,
                reason=f"接替车辆 {vehicle.plate_number}（电量不足）"
            )

            result["reserve_vehicle"] = {
                "vehicle_id": reserve.id,
                "plate_number": reserve.plate_number,
                "battery_level": reserve.current_battery
            }

            CRUDDispatchLog.create(
                db,
                f"调度备用车 {reserve.plate_number} 接替 {vehicle.plate_number}",
                "INFO",
                vehicle.id
            )
        else:
            result["message"] += "，但无可用备用车"
            CRUDDispatchLog.create(
                db,
                f"车辆 {vehicle.plate_number} 需要充电但无备用车可用",
                "WARNING",
                vehicle.id
            )

        return result

    @staticmethod
    def check_last_run_schedule(db: Session, schedule) -> bool:
        from app.crud import CRUDRoute
        route = CRUDRoute.get_by_id(db, schedule.route_id)
        if not route:
            return False

        now = datetime.utcnow()
        required_return_time = schedule.departure_time + timedelta(minutes=route.estimated_duration_min)

        return schedule.return_time >= required_return_time

    @staticmethod
    def periodic_monitor(db: Session) -> dict:
        vehicles = CRUDVehicle.get_all(db)
        results = []

        for vehicle in vehicles:
            if vehicle.status == VehicleStatus.CHARGING:
                active_session = CRUDChargingSession.get_active_by_vehicle(db, vehicle.id)
                if active_session:
                    progress = ChargingService.update_charging_progress(db, active_session)
                    results.append({
                        "vehicle_id": vehicle.id,
                        "type": "charging_update",
                        "data": progress
                    })
            elif vehicle.status == VehicleStatus.RUNNING:
                dispatch_result = DispatchService.check_and_dispatch(db, vehicle)
                if dispatch_result["needs_intervention"]:
                    results.append({
                        "vehicle_id": vehicle.id,
                        "type": "dispatch_intervention",
                        "data": dispatch_result
                    })

        return {
            "total_vehicles": len(vehicles),
            "processed_events": len(results),
            "events": results
        }
