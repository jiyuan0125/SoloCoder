from sqlalchemy.orm import Session
from app.models import (
    Station, Charger, Vehicle, Route, Schedule, ChargingSession,
    Assignment, Alert, DispatchLog,
    ChargerStatus, VehicleStatus, AssignmentStatus,
    AlertType, AlertStatus, NotificationType
)
from app.schemas import (
    StationCreate, StationUpdate,
    ChargerCreate, ChargerUpdate,
    VehicleCreate, VehicleUpdate,
    RouteCreate, RouteUpdate,
    ScheduleCreate
)
from typing import Optional, List
from datetime import datetime

def round_battery(value: Optional[float]) -> Optional[float]:
    if value is None:
        return 100.0
    return round(value, 1)

class CRUDStation:
    @staticmethod
    def get_all(db: Session) -> List[Station]:
        return db.query(Station).all()

    @staticmethod
    def get_by_id(db: Session, station_id: int) -> Optional[Station]:
        return db.query(Station).filter(Station.id == station_id).first()

    @staticmethod
    def create(db: Session, data: StationCreate) -> Station:
        station = Station(**data.model_dump())
        db.add(station)
        db.commit()
        db.refresh(station)
        return station

    @staticmethod
    def update(db: Session, station: Station, data: StationUpdate) -> Station:
        for key, value in data.model_dump(exclude_unset=True).items():
            setattr(station, key, value)
        db.commit()
        db.refresh(station)
        return station

class CRUDCharger:
    @staticmethod
    def get_all(db: Session) -> List[Charger]:
        return db.query(Charger).all()

    @staticmethod
    def get_by_id(db: Session, charger_id: int) -> Optional[Charger]:
        return db.query(Charger).filter(Charger.id == charger_id).first()

    @staticmethod
    def get_by_station(db: Session, station_id: int) -> List[Charger]:
        return db.query(Charger).filter(Charger.station_id == station_id).all()

    @staticmethod
    def get_idle_by_station(db: Session, station_id: int) -> List[Charger]:
        return db.query(Charger).filter(
            Charger.station_id == station_id,
            Charger.status == ChargerStatus.IDLE
        ).order_by(Charger.total_charging_count.asc()).all()

    @staticmethod
    def create(db: Session, data: ChargerCreate) -> Charger:
        charger = Charger(**data.model_dump())
        db.add(charger)
        db.commit()
        db.refresh(charger)
        return charger

    @staticmethod
    def update(db: Session, charger: Charger, data: ChargerUpdate) -> Charger:
        for key, value in data.model_dump(exclude_unset=True).items():
            setattr(charger, key, value)
        db.commit()
        db.refresh(charger)
        return charger

    @staticmethod
    def lock(db: Session, charger: Charger, vehicle_id: int) -> bool:
        if charger.status != ChargerStatus.IDLE:
            return False
        charger.status = ChargerStatus.LOCKED
        charger.locked_by_vehicle_id = vehicle_id
        db.commit()
        return True

    @staticmethod
    def release(db: Session, charger: Charger) -> None:
        charger.status = ChargerStatus.IDLE
        charger.locked_by_vehicle_id = None
        db.commit()

    @staticmethod
    def set_charging(db: Session, charger: Charger) -> None:
        charger.status = ChargerStatus.CHARGING
        db.commit()

    @staticmethod
    def increment_charging_count(db: Session, charger: Charger) -> None:
        charger.total_charging_count += 1
        db.commit()

    @staticmethod
    def set_faulty(db: Session, charger: Charger) -> None:
        charger.status = ChargerStatus.FAULTY
        db.commit()

class CRUDVehicle:
    @staticmethod
    def get_all(db: Session) -> List[Vehicle]:
        return db.query(Vehicle).all()

    @staticmethod
    def get_by_id(db: Session, vehicle_id: int) -> Optional[Vehicle]:
        return db.query(Vehicle).filter(Vehicle.id == vehicle_id).first()

    @staticmethod
    def get_by_plate(db: Session, plate: str) -> Optional[Vehicle]:
        return db.query(Vehicle).filter(Vehicle.plate_number == plate).first()

    @staticmethod
    def get_reserve_vehicles(db: Session) -> List[Vehicle]:
        return db.query(Vehicle).filter(Vehicle.status == VehicleStatus.RESERVE).all()

    @staticmethod
    def create(db: Session, data: VehicleCreate) -> Vehicle:
        vehicle = Vehicle(**data.model_dump())
        vehicle.current_battery = round_battery(vehicle.current_battery)
        db.add(vehicle)
        db.commit()
        db.refresh(vehicle)
        return vehicle

    @staticmethod
    def update(db: Session, vehicle: Vehicle, data: VehicleUpdate) -> Vehicle:
        update_data = data.model_dump(exclude_unset=True)
        if 'current_battery' in update_data:
            update_data['current_battery'] = round_battery(update_data['current_battery'])
        for key, value in update_data.items():
            setattr(vehicle, key, value)
        db.commit()
        db.refresh(vehicle)
        return vehicle

    @staticmethod
    def update_battery(db: Session, vehicle: Vehicle, battery: float) -> Vehicle:
        vehicle.current_battery = round_battery(battery)
        db.commit()
        db.refresh(vehicle)
        return vehicle

    @staticmethod
    def update_location(db: Session, vehicle: Vehicle, lat: float, lng: float) -> Vehicle:
        vehicle.latitude = lat
        vehicle.longitude = lng
        vehicle.last_report_time = datetime.utcnow()
        db.commit()
        db.refresh(vehicle)
        return vehicle

    @staticmethod
    def lock(db: Session, vehicle: Vehicle) -> bool:
        if vehicle.status not in [VehicleStatus.IDLE, VehicleStatus.RUNNING]:
            return False
        vehicle.status = VehicleStatus.LOCKED
        db.commit()
        return True

    @staticmethod
    def release(db: Session, vehicle: Vehicle) -> None:
        vehicle.status = VehicleStatus.IDLE
        db.commit()

    @staticmethod
    def set_charging(db: Session, vehicle: Vehicle) -> None:
        vehicle.status = VehicleStatus.CHARGING
        db.commit()

    @staticmethod
    def set_running(db: Session, vehicle: Vehicle) -> None:
        vehicle.status = VehicleStatus.RUNNING
        db.commit()

    @staticmethod
    def set_reserve(db: Session, vehicle: Vehicle) -> None:
        vehicle.status = VehicleStatus.RESERVE
        db.commit()

class CRUDRoute:
    @staticmethod
    def get_all(db: Session) -> List[Route]:
        return db.query(Route).all()

    @staticmethod
    def get_by_id(db: Session, route_id: int) -> Optional[Route]:
        return db.query(Route).filter(Route.id == route_id).first()

    @staticmethod
    def create(db: Session, data: RouteCreate) -> Route:
        route = Route(**data.model_dump())
        db.add(route)
        db.commit()
        db.refresh(route)
        return route

    @staticmethod
    def update(db: Session, route: Route, data: RouteUpdate) -> Route:
        for key, value in data.model_dump(exclude_unset=True).items():
            setattr(route, key, value)
        db.commit()
        db.refresh(route)
        return route

class CRUDSchedule:
    @staticmethod
    def get_all(db: Session) -> List[Schedule]:
        return db.query(Schedule).all()

    @staticmethod
    def get_by_id(db: Session, schedule_id: int) -> Optional[Schedule]:
        return db.query(Schedule).filter(Schedule.id == schedule_id).first()

    @staticmethod
    def create(db: Session, data: ScheduleCreate) -> Schedule:
        schedule = Schedule(**data.model_dump())
        db.add(schedule)
        db.commit()
        db.refresh(schedule)
        return schedule

    @staticmethod
    def assign_vehicle(db: Session, schedule: Schedule, vehicle_id: int) -> None:
        schedule.vehicle_id = vehicle_id
        db.commit()

class CRUDChargingSession:
    @staticmethod
    def get_all(db: Session) -> List[ChargingSession]:
        return db.query(ChargingSession).all()

    @staticmethod
    def get_by_id(db: Session, session_id: int) -> Optional[ChargingSession]:
        return db.query(ChargingSession).filter(ChargingSession.id == session_id).first()

    @staticmethod
    def get_active_by_vehicle(db: Session, vehicle_id: int) -> Optional[ChargingSession]:
        return db.query(ChargingSession).filter(
            ChargingSession.vehicle_id == vehicle_id,
            ChargingSession.status == AssignmentStatus.IN_PROGRESS
        ).first()

    @staticmethod
    def get_active_by_charger(db: Session, charger_id: int) -> Optional[ChargingSession]:
        return db.query(ChargingSession).filter(
            ChargingSession.charger_id == charger_id,
            ChargingSession.status == AssignmentStatus.IN_PROGRESS
        ).first()

    @staticmethod
    def create(
        db: Session,
        vehicle_id: int,
        charger_id: int,
        start_battery: float,
        target_battery: float = 90.0,
        estimated_end_time: Optional[datetime] = None
    ) -> ChargingSession:
        session = ChargingSession(
            vehicle_id=vehicle_id,
            charger_id=charger_id,
            start_time=datetime.utcnow(),
            start_battery=round_battery(start_battery),
            target_battery=round_battery(target_battery),
            current_battery=round_battery(start_battery),
            estimated_end_time=estimated_end_time,
            status=AssignmentStatus.IN_PROGRESS
        )
        db.add(session)
        db.commit()
        db.refresh(session)
        return session

    @staticmethod
    def update_progress(db: Session, session: ChargingSession, current_battery: float) -> None:
        session.current_battery = round_battery(current_battery)
        db.commit()

    @staticmethod
    def complete(db: Session, session: ChargingSession, end_battery: float) -> None:
        session.status = AssignmentStatus.COMPLETED
        session.end_time = datetime.utcnow()
        session.current_battery = round_battery(end_battery)
        db.commit()

    @staticmethod
    def cancel(db: Session, session: ChargingSession) -> None:
        session.status = AssignmentStatus.CANCELLED
        session.end_time = datetime.utcnow()
        db.commit()

class CRUDAssignment:
    @staticmethod
    def create(
        db: Session,
        vehicle_id: int,
        route_id: Optional[int] = None,
        schedule_id: Optional[int] = None,
        reason: Optional[str] = None
    ) -> Assignment:
        assignment = Assignment(
            vehicle_id=vehicle_id,
            route_id=route_id,
            schedule_id=schedule_id,
            status=AssignmentStatus.PENDING,
            reason=reason
        )
        db.add(assignment)
        db.commit()
        db.refresh(assignment)
        return assignment

    @staticmethod
    def update_status(db: Session, assignment: Assignment, status: AssignmentStatus) -> None:
        assignment.status = status
        db.commit()

class CRUDAlert:
    @staticmethod
    def get_all(db: Session) -> List[Alert]:
        return db.query(Alert).all()

    @staticmethod
    def get_active(db: Session) -> List[Alert]:
        return db.query(Alert).filter(Alert.status == AlertStatus.ACTIVE).all()

    @staticmethod
    def create(
        db: Session,
        alert_type: AlertType,
        message: str,
        vehicle_id: Optional[int] = None,
        charger_id: Optional[int] = None,
        notification_type: NotificationType = NotificationType.TERMINAL
    ) -> Alert:
        alert = Alert(
            type=alert_type,
            vehicle_id=vehicle_id,
            charger_id=charger_id,
            message=message,
            status=AlertStatus.ACTIVE,
            notification_type=notification_type
        )
        db.add(alert)
        db.commit()
        db.refresh(alert)
        return alert

    @staticmethod
    def acknowledge(db: Session, alert: Alert) -> None:
        alert.status = AlertStatus.ACKNOWLEDGED
        db.commit()

    @staticmethod
    def resolve(db: Session, alert: Alert) -> None:
        alert.status = AlertStatus.RESOLVED
        alert.resolved_at = datetime.utcnow()
        db.commit()

class CRUDDispatchLog:
    @staticmethod
    def get_all(db: Session) -> List[DispatchLog]:
        return db.query(DispatchLog).all()

    @staticmethod
    def create(
        db: Session,
        message: str,
        level: str = "INFO",
        vehicle_id: Optional[int] = None,
        charger_id: Optional[int] = None
    ) -> DispatchLog:
        log = DispatchLog(
            level=level,
            message=message,
            related_vehicle_id=vehicle_id,
            related_charger_id=charger_id
        )
        db.add(log)
        db.commit()
        db.refresh(log)
        return log
