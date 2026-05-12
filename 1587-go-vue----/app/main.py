from fastapi import FastAPI, Depends, HTTPException, BackgroundTasks
from sqlalchemy.orm import Session
from typing import List
from datetime import datetime

from app.database import engine, Base, get_db
from app.models import (
    ChargerStatus, VehicleStatus, AlertType, AlertStatus, NotificationType
)
from app.schemas import (
    StationCreate, StationUpdate, StationResponse,
    ChargerCreate, ChargerUpdate, ChargerResponse,
    VehicleCreate, VehicleUpdate, VehicleStatusReport, VehicleResponse,
    RouteCreate, RouteUpdate, RouteResponse,
    ScheduleCreate, ScheduleResponse,
    ChargingSessionResponse, AlertResponse,
    DispatchLogResponse, ChargingProgress
)
from app.crud import (
    CRUDStation, CRUDCharger, CRUDVehicle, CRUDRoute, CRUDSchedule,
    CRUDChargingSession, CRUDAlert, CRUDDispatchLog
)
from app.services import (
    ChargingService, VehicleReportService, DispatchService
)

Base.metadata.create_all(bind=engine)

app = FastAPI(title="有轨电车运营管理系统", version="1.0.0")

@app.get("/")
def root():
    return {"message": "有轨电车运营管理系统 API", "version": "1.0.0"}

@app.post("/stations/", response_model=StationResponse)
def create_station(data: StationCreate, db: Session = Depends(get_db)):
    return CRUDStation.create(db, data)

@app.get("/stations/", response_model=List[StationResponse])
def list_stations(db: Session = Depends(get_db)):
    return CRUDStation.get_all(db)

@app.get("/stations/{station_id}", response_model=StationResponse)
def get_station(station_id: int, db: Session = Depends(get_db)):
    station = CRUDStation.get_by_id(db, station_id)
    if not station:
        raise HTTPException(status_code=404, detail="站点不存在")
    return station

@app.put("/stations/{station_id}", response_model=StationResponse)
def update_station(station_id: int, data: StationUpdate, db: Session = Depends(get_db)):
    station = CRUDStation.get_by_id(db, station_id)
    if not station:
        raise HTTPException(status_code=404, detail="站点不存在")
    return CRUDStation.update(db, station, data)

@app.post("/chargers/", response_model=ChargerResponse)
def create_charger(data: ChargerCreate, db: Session = Depends(get_db)):
    return CRUDCharger.create(db, data)

@app.get("/chargers/", response_model=List[ChargerResponse])
def list_chargers(db: Session = Depends(get_db)):
    return CRUDCharger.get_all(db)

@app.get("/chargers/{charger_id}", response_model=ChargerResponse)
def get_charger(charger_id: int, db: Session = Depends(get_db)):
    charger = CRUDCharger.get_by_id(db, charger_id)
    if not charger:
        raise HTTPException(status_code=404, detail="充电桩不存在")
    return charger

@app.put("/chargers/{charger_id}", response_model=ChargerResponse)
def update_charger(charger_id: int, data: ChargerUpdate, db: Session = Depends(get_db)):
    charger = CRUDCharger.get_by_id(db, charger_id)
    if not charger:
        raise HTTPException(status_code=404, detail="充电桩不存在")
    return CRUDCharger.update(db, charger, data)

@app.post("/chargers/{charger_id}/fault")
def report_charger_fault(charger_id: int, db: Session = Depends(get_db)):
    charger = CRUDCharger.get_by_id(db, charger_id)
    if not charger:
        raise HTTPException(status_code=404, detail="充电桩不存在")
    ChargingService.handle_charger_fault(db, charger)
    return {"message": "充电桩故障已处理", "charger_id": charger_id}

@app.post("/vehicles/", response_model=VehicleResponse)
def create_vehicle(data: VehicleCreate, db: Session = Depends(get_db)):
    return CRUDVehicle.create(db, data)

@app.get("/vehicles/", response_model=List[VehicleResponse])
def list_vehicles(db: Session = Depends(get_db)):
    return CRUDVehicle.get_all(db)

@app.get("/vehicles/{vehicle_id}", response_model=VehicleResponse)
def get_vehicle(vehicle_id: int, db: Session = Depends(get_db)):
    vehicle = CRUDVehicle.get_by_id(db, vehicle_id)
    if not vehicle:
        raise HTTPException(status_code=404, detail="车辆不存在")
    return vehicle

@app.get("/vehicles/{vehicle_id}/status")
def get_vehicle_detailed_status(vehicle_id: int, db: Session = Depends(get_db)):
    vehicle = CRUDVehicle.get_by_id(db, vehicle_id)
    if not vehicle:
        raise HTTPException(status_code=404, detail="车辆不存在")

    active_session = CRUDChargingSession.get_active_by_vehicle(db, vehicle_id)
    charging_status = None

    if active_session:
        charger = CRUDCharger.get_by_id(db, active_session.charger_id)
        charging_status = {
            "session_id": active_session.id,
            "charger_name": charger.name if charger else None,
            "start_time": active_session.start_time,
            "estimated_end": active_session.estimated_end_time,
            "current_battery": active_session.current_battery,
            "target_battery": active_session.target_battery
        }

    return {
        "vehicle_id": vehicle.id,
        "plate_number": vehicle.plate_number,
        "status": vehicle.status.value,
        "current_battery": vehicle.current_battery,
        "current_station_id": vehicle.current_station_id,
        "current_route_id": vehicle.current_route_id,
        "last_report_time": vehicle.last_report_time,
        "charging_status": charging_status
    }

@app.put("/vehicles/{vehicle_id}", response_model=VehicleResponse)
def update_vehicle(vehicle_id: int, data: VehicleUpdate, db: Session = Depends(get_db)):
    vehicle = CRUDVehicle.get_by_id(db, vehicle_id)
    if not vehicle:
        raise HTTPException(status_code=404, detail="车辆不存在")
    return CRUDVehicle.update(db, vehicle, data)

@app.post("/vehicles/{vehicle_id}/report")
def report_vehicle_status(
    vehicle_id: int,
    data: VehicleStatusReport,
    db: Session = Depends(get_db)
):
    vehicle = CRUDVehicle.get_by_id(db, vehicle_id)
    if not vehicle:
        raise HTTPException(status_code=404, detail="车辆不存在")

    if vehicle.status == VehicleStatus.CHARGING:
        raise HTTPException(status_code=400, detail="车辆正在充电，无法上报状态")

    result = VehicleReportService.report_vehicle_status(
        db, vehicle, data.latitude, data.longitude, data.current_battery
    )
    return result

@app.post("/vehicles/{vehicle_id}/arrive")
def vehicle_arrive_at_station(
    vehicle_id: int,
    station_id: int,
    db: Session = Depends(get_db)
):
    vehicle = CRUDVehicle.get_by_id(db, vehicle_id)
    if not vehicle:
        raise HTTPException(status_code=404, detail="车辆不存在")

    station = CRUDStation.get_by_id(db, station_id)
    if not station:
        raise HTTPException(status_code=404, detail="站点不存在")

    vehicle.current_station_id = station_id
    db.commit()
    db.refresh(vehicle)

    result = ChargingService.process_vehicle_arrival(db, vehicle)
    return result

@app.post("/vehicles/{vehicle_id}/depart")
def vehicle_depart_from_station(vehicle_id: int, db: Session = Depends(get_db)):
    vehicle = CRUDVehicle.get_by_id(db, vehicle_id)
    if not vehicle:
        raise HTTPException(status_code=404, detail="车辆不存在")

    if vehicle.status == VehicleStatus.CHARGING:
        raise HTTPException(status_code=400, detail="车辆正在充电，无法出发")

    vehicle.current_station_id = None
    CRUDVehicle.set_running(db, vehicle)
    return {"message": "车辆已出发", "vehicle_id": vehicle_id}

@app.post("/routes/", response_model=RouteResponse)
def create_route(data: RouteCreate, db: Session = Depends(get_db)):
    return CRUDRoute.create(db, data)

@app.get("/routes/", response_model=List[RouteResponse])
def list_routes(db: Session = Depends(get_db)):
    return CRUDRoute.get_all(db)

@app.get("/routes/{route_id}", response_model=RouteResponse)
def get_route(route_id: int, db: Session = Depends(get_db)):
    route = CRUDRoute.get_by_id(db, route_id)
    if not route:
        raise HTTPException(status_code=404, detail="线路不存在")
    return route

@app.put("/routes/{route_id}", response_model=RouteResponse)
def update_route(route_id: int, data: RouteUpdate, db: Session = Depends(get_db)):
    route = CRUDRoute.get_by_id(db, route_id)
    if not route:
        raise HTTPException(status_code=404, detail="线路不存在")
    return CRUDRoute.update(db, route, data)

@app.post("/schedules/", response_model=ScheduleResponse)
def create_schedule(data: ScheduleCreate, db: Session = Depends(get_db)):
    if data.is_last_run:
        from app.crud import CRUDRoute
        route = CRUDRoute.get_by_id(db, data.route_id)
        if route:
            required_return = data.departure_time
            if data.return_time < required_return:
                raise HTTPException(
                    status_code=400,
                    detail="末班车返回时间必须确保能跑完全程并回车场"
                )
    return CRUDSchedule.create(db, data)

@app.get("/schedules/", response_model=List[ScheduleResponse])
def list_schedules(db: Session = Depends(get_db)):
    return CRUDSchedule.get_all(db)

@app.get("/charging-sessions/", response_model=List[ChargingSessionResponse])
def list_charging_sessions(db: Session = Depends(get_db)):
    return CRUDChargingSession.get_all(db)

@app.get("/charging-sessions/active", response_model=List[ChargingSessionResponse])
def list_active_sessions(db: Session = Depends(get_db)):
    from app.models import ChargingSession, AssignmentStatus
    return db.query(ChargingSession).filter(
        ChargingSession.status == AssignmentStatus.IN_PROGRESS
    ).all()

@app.get("/charging-sessions/{session_id}", response_model=ChargingSessionResponse)
def get_charging_session(session_id: int, db: Session = Depends(get_db)):
    session = CRUDChargingSession.get_by_id(db, session_id)
    if not session:
        raise HTTPException(status_code=404, detail="充电会话不存在")
    return session

@app.get("/charging-sessions/{session_id}/progress", response_model=ChargingProgress)
def get_charging_progress(session_id: int, db: Session = Depends(get_db)):
    session = CRUDChargingSession.get_by_id(db, session_id)
    if not session:
        raise HTTPException(status_code=404, detail="充电会话不存在")

    progress = ChargingService.update_charging_progress(db, session)
    return ChargingProgress(
        session_id=session.id,
        vehicle_id=session.vehicle_id,
        charger_id=session.charger_id,
        current_battery=progress["current_battery"],
        target_battery=session.target_battery,
        progress_percent=progress["progress"],
        estimated_end_time=session.estimated_end_time
    )

@app.get("/alerts/", response_model=List[AlertResponse])
def list_alerts(db: Session = Depends(get_db)):
    return CRUDAlert.get_all(db)

@app.get("/alerts/active", response_model=List[AlertResponse])
def list_active_alerts(db: Session = Depends(get_db)):
    return CRUDAlert.get_active(db)

@app.put("/alerts/{alert_id}/acknowledge")
def acknowledge_alert(alert_id: int, db: Session = Depends(get_db)):
    from app.models import Alert
    alert = db.query(Alert).filter(Alert.id == alert_id).first()
    if not alert:
        raise HTTPException(status_code=404, detail="告警不存在")
    CRUDAlert.acknowledge(db, alert)
    return {"message": "告警已确认", "alert_id": alert_id}

@app.put("/alerts/{alert_id}/resolve")
def resolve_alert(alert_id: int, db: Session = Depends(get_db)):
    from app.models import Alert
    alert = db.query(Alert).filter(Alert.id == alert_id).first()
    if not alert:
        raise HTTPException(status_code=404, detail="告警不存在")
    CRUDAlert.resolve(db, alert)
    return {"message": "告警已解决", "alert_id": alert_id}

@app.get("/dispatch-logs/", response_model=List[DispatchLogResponse])
def list_dispatch_logs(db: Session = Depends(get_db)):
    return CRUDDispatchLog.get_all(db)

@app.post("/monitor/run")
def run_periodic_monitor(db: Session = Depends(get_db)):
    result = DispatchService.periodic_monitor(db)
    return result

@app.get("/monitor/vehicle-range/{vehicle_id}")
def get_vehicle_remaining_range(vehicle_id: int, db: Session = Depends(get_db)):
    vehicle = CRUDVehicle.get_by_id(db, vehicle_id)
    if not vehicle:
        raise HTTPException(status_code=404, detail="车辆不存在")

    remaining_range = DispatchService.estimate_remaining_range_km(vehicle)
    can_complete = DispatchService.can_complete_current_assignment(db, vehicle)

    return {
        "vehicle_id": vehicle.id,
        "plate_number": vehicle.plate_number,
        "current_battery": vehicle.current_battery,
        "remaining_range_km": round(remaining_range, 2),
        "can_complete_current_assignment": can_complete,
        "current_route_id": vehicle.current_route_id
    }
