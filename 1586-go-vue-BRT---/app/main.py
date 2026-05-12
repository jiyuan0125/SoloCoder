from fastapi import FastAPI, Depends, HTTPException, Query
from fastapi.responses import JSONResponse
from sqlalchemy.orm import Session
from sqlalchemy.exc import IntegrityError
from typing import Optional, List
from datetime import datetime, timedelta
import os
import json

from app.database import get_db, engine, Base
from app.models import (
    Station, Route, RouteStation, Vehicle, 
    SignalPriority, OperationParameter, Schedule, 
    ArrivalRecord, Direction, ScheduleStatus
)
from app.schemas import (
    StationCreate, StationUpdate, StationResponse,
    RouteCreate, RouteUpdate, RouteResponse,
    RouteStationCreate, RouteStationUpdate, RouteStationResponse,
    VehicleCreate, VehicleUpdate, VehicleResponse,
    SignalPriorityCreate, SignalPriorityUpdate, SignalPriorityResponse,
    OperationParameterCreate, OperationParameterUpdate, OperationParameterResponse,
    ScheduleResponse, ScheduleDetailResponse,
    ArrivalRecordCreate, ArrivalRecordResponse,
    SignalPriorityRequestResponse, SignalPriorityRequest,
    PunctualityReport, ScheduleGenerationRequest, ScheduleStation
)
from app.utils import (
    check_and_request_signal_priority, check_vehicle_crowd,
    is_on_time
)
from app.scheduler import (
    generate_schedule_for_route, update_route_stations_and_reschedule,
    get_schedules_for_route, get_punctuality_report
)

Base.metadata.create_all(bind=engine)

app = FastAPI(title="BRT运营调度管理系统", version="1.0.0")


@app.get("/")
def root():
    return {"message": "BRT运营调度管理系统", "version": "1.0.0"}


@app.post("/stations/", response_model=StationResponse)
def create_station(station: StationCreate, db: Session = Depends(get_db)):
    db_station = Station(**station.model_dump())
    db.add(db_station)
    try:
        db.commit()
        db.refresh(db_station)
        return db_station
    except IntegrityError:
        db.rollback()
        raise HTTPException(status_code=400, detail="站点编码已存在")


@app.get("/stations/", response_model=List[StationResponse])
def get_stations(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return db.query(Station).offset(skip).limit(limit).all()


@app.get("/stations/{station_id}", response_model=StationResponse)
def get_station(station_id: int, db: Session = Depends(get_db)):
    station = db.query(Station).filter(Station.id == station_id).first()
    if not station:
        raise HTTPException(status_code=404, detail="站点不存在")
    return station


@app.put("/stations/{station_id}", response_model=StationResponse)
def update_station(station_id: int, station_update: StationUpdate, db: Session = Depends(get_db)):
    station = db.query(Station).filter(Station.id == station_id).first()
    if not station:
        raise HTTPException(status_code=404, detail="站点不存在")
    
    for key, value in station_update.model_dump(exclude_unset=True).items():
        setattr(station, key, value)
    
    db.commit()
    db.refresh(station)
    return station


@app.delete("/stations/{station_id}")
def delete_station(station_id: int, db: Session = Depends(get_db)):
    station = db.query(Station).filter(Station.id == station_id).first()
    if not station:
        raise HTTPException(status_code=404, detail="站点不存在")
    
    db.delete(station)
    db.commit()
    return {"message": "站点已删除"}


@app.post("/routes/", response_model=RouteResponse)
def create_route(route: RouteCreate, db: Session = Depends(get_db)):
    db_route = Route(**route.model_dump())
    db.add(db_route)
    try:
        db.commit()
        db.refresh(db_route)
        return db_route
    except IntegrityError:
        db.rollback()
        raise HTTPException(status_code=400, detail="线路编码已存在")


@app.get("/routes/", response_model=List[RouteResponse])
def get_routes(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return db.query(Route).offset(skip).limit(limit).all()


@app.get("/routes/{route_id}", response_model=RouteResponse)
def get_route(route_id: int, db: Session = Depends(get_db)):
    route = db.query(Route).filter(Route.id == route_id).first()
    if not route:
        raise HTTPException(status_code=404, detail="线路不存在")
    return route


@app.put("/routes/{route_id}", response_model=RouteResponse)
def update_route(route_id: int, route_update: RouteUpdate, db: Session = Depends(get_db)):
    route = db.query(Route).filter(Route.id == route_id).first()
    if not route:
        raise HTTPException(status_code=404, detail="线路不存在")
    
    for key, value in route_update.model_dump(exclude_unset=True).items():
        setattr(route, key, value)
    
    db.commit()
    db.refresh(route)
    return route


@app.delete("/routes/{route_id}")
def delete_route(route_id: int, db: Session = Depends(get_db)):
    route = db.query(Route).filter(Route.id == route_id).first()
    if not route:
        raise HTTPException(status_code=404, detail="线路不存在")
    
    db.delete(route)
    db.commit()
    return {"message": "线路已删除"}


@app.post("/route-stations/", response_model=RouteStationResponse)
def add_route_station(rs: RouteStationCreate, db: Session = Depends(get_db)):
    existing = db.query(RouteStation).filter(
        RouteStation.route_id == rs.route_id,
        RouteStation.direction == rs.direction,
        RouteStation.sequence == rs.sequence
    ).first()
    
    if existing:
        raise HTTPException(status_code=400, detail="该位置已有站点")
    
    db_rs = RouteStation(**rs.model_dump())
    db.add(db_rs)
    
    try:
        db.commit()
        db.refresh(db_rs)
        
        update_route_stations_and_reschedule(db, rs.route_id, rs.direction)
        
        station = db.query(Station).filter(Station.id == rs.station_id).first()
        return RouteStationResponse(
            id=db_rs.id,
            route_id=db_rs.route_id,
            station_id=db_rs.station_id,
            direction=db_rs.direction,
            sequence=db_rs.sequence,
            travel_time_from_prev=db_rs.travel_time_from_prev,
            stop_time=db_rs.stop_time,
            station_name=station.name if station else None,
            station_code=station.code if station else None
        )
    except IntegrityError:
        db.rollback()
        raise HTTPException(status_code=400, detail="添加站点失败")


@app.get("/routes/{route_id}/stations", response_model=List[RouteStationResponse])
def get_route_stations(route_id: int, direction: Optional[Direction] = None, db: Session = Depends(get_db)):
    query = db.query(RouteStation).filter(RouteStation.route_id == route_id)
    if direction:
        query = query.filter(RouteStation.direction == direction)
    
    route_stations = query.order_by(RouteStation.direction, RouteStation.sequence).all()
    
    result = []
    for rs in route_stations:
        station = db.query(Station).filter(Station.id == rs.station_id).first()
        result.append(RouteStationResponse(
            id=rs.id,
            route_id=rs.route_id,
            station_id=rs.station_id,
            direction=rs.direction,
            sequence=rs.sequence,
            travel_time_from_prev=rs.travel_time_from_prev,
            stop_time=rs.stop_time,
            station_name=station.name if station else None,
            station_code=station.code if station else None
        ))
    
    return result


@app.put("/route-stations/{rs_id}", response_model=RouteStationResponse)
def update_route_station(rs_id: int, rs_update: RouteStationUpdate, db: Session = Depends(get_db)):
    rs = db.query(RouteStation).filter(RouteStation.id == rs_id).first()
    if not rs:
        raise HTTPException(status_code=404, detail="线路站点关系不存在")
    
    for key, value in rs_update.model_dump(exclude_unset=True).items():
        setattr(rs, key, value)
    
    db.commit()
    db.refresh(rs)
    
    update_route_stations_and_reschedule(db, rs.route_id, rs.direction)
    
    station = db.query(Station).filter(Station.id == rs.station_id).first()
    return RouteStationResponse(
        id=rs.id,
        route_id=rs.route_id,
        station_id=rs.station_id,
        direction=rs.direction,
        sequence=rs.sequence,
        travel_time_from_prev=rs.travel_time_from_prev,
        stop_time=rs.stop_time,
        station_name=station.name if station else None,
        station_code=station.code if station else None
    )


@app.delete("/route-stations/{rs_id}")
def delete_route_station(rs_id: int, db: Session = Depends(get_db)):
    rs = db.query(RouteStation).filter(RouteStation.id == rs_id).first()
    if not rs:
        raise HTTPException(status_code=404, detail="线路站点关系不存在")
    
    route_id = rs.route_id
    direction = rs.direction
    
    db.delete(rs)
    db.commit()
    
    update_route_stations_and_reschedule(db, route_id, direction)
    
    return {"message": "站点已从线路中删除，未执行排班已同步调整"}


@app.post("/vehicles/", response_model=VehicleResponse)
def create_vehicle(vehicle: VehicleCreate, db: Session = Depends(get_db)):
    db_vehicle = Vehicle(**vehicle.model_dump())
    db.add(db_vehicle)
    try:
        db.commit()
        db.refresh(db_vehicle)
        return VehicleResponse(
            id=db_vehicle.id,
            plate_number=db_vehicle.plate_number,
            vehicle_code=db_vehicle.vehicle_code,
            capacity=db_vehicle.capacity,
            current_route_id=db_vehicle.current_route_id,
            current_direction=db_vehicle.current_direction,
            status=db_vehicle.status,
            current_load=db_vehicle.current_load,
            is_crowded=check_vehicle_crowd(db_vehicle),
            created_at=db_vehicle.created_at,
            updated_at=db_vehicle.updated_at
        )
    except IntegrityError:
        db.rollback()
        raise HTTPException(status_code=400, detail="车辆信息已存在")


@app.get("/vehicles/", response_model=List[VehicleResponse])
def get_vehicles(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    vehicles = db.query(Vehicle).offset(skip).limit(limit).all()
    result = []
    for v in vehicles:
        result.append(VehicleResponse(
            id=v.id,
            plate_number=v.plate_number,
            vehicle_code=v.vehicle_code,
            capacity=v.capacity,
            current_route_id=v.current_route_id,
            current_direction=v.current_direction,
            status=v.status,
            current_load=v.current_load,
            is_crowded=check_vehicle_crowd(v),
            created_at=v.created_at,
            updated_at=v.updated_at
        ))
    return result


@app.get("/vehicles/{vehicle_id}", response_model=VehicleResponse)
def get_vehicle(vehicle_id: int, db: Session = Depends(get_db)):
    vehicle = db.query(Vehicle).filter(Vehicle.id == vehicle_id).first()
    if not vehicle:
        raise HTTPException(status_code=404, detail="车辆不存在")
    
    return VehicleResponse(
        id=vehicle.id,
        plate_number=vehicle.plate_number,
        vehicle_code=vehicle.vehicle_code,
        capacity=vehicle.capacity,
        current_route_id=vehicle.current_route_id,
        current_direction=vehicle.current_direction,
        status=vehicle.status,
        current_load=vehicle.current_load,
        is_crowded=check_vehicle_crowd(vehicle),
        created_at=vehicle.created_at,
        updated_at=vehicle.updated_at
    )


@app.put("/vehicles/{vehicle_id}", response_model=VehicleResponse)
def update_vehicle(vehicle_id: int, vehicle_update: VehicleUpdate, db: Session = Depends(get_db)):
    vehicle = db.query(Vehicle).filter(Vehicle.id == vehicle_id).first()
    if not vehicle:
        raise HTTPException(status_code=404, detail="车辆不存在")
    
    for key, value in vehicle_update.model_dump(exclude_unset=True).items():
        setattr(vehicle, key, value)
    
    db.commit()
    db.refresh(vehicle)
    
    return VehicleResponse(
        id=vehicle.id,
        plate_number=vehicle.plate_number,
        vehicle_code=vehicle.vehicle_code,
        capacity=vehicle.capacity,
        current_route_id=vehicle.current_route_id,
        current_direction=vehicle.current_direction,
        status=vehicle.status,
        current_load=vehicle.current_load,
        is_crowded=check_vehicle_crowd(vehicle),
        created_at=vehicle.created_at,
        updated_at=vehicle.updated_at
    )


@app.post("/signal-priorities/", response_model=SignalPriorityResponse)
def create_signal_priority(sp: SignalPriorityCreate, db: Session = Depends(get_db)):
    db_sp = SignalPriority(**sp.model_dump())
    db.add(db_sp)
    db.commit()
    db.refresh(db_sp)
    return db_sp


@app.get("/signal-priorities/", response_model=List[SignalPriorityResponse])
def get_signal_priorities(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return db.query(SignalPriority).offset(skip).limit(limit).all()


@app.put("/signal-priorities/{sp_id}", response_model=SignalPriorityResponse)
def update_signal_priority(sp_id: int, sp_update: SignalPriorityUpdate, db: Session = Depends(get_db)):
    sp = db.query(SignalPriority).filter(SignalPriority.id == sp_id).first()
    if not sp:
        raise HTTPException(status_code=404, detail="信号优先配置不存在")
    
    for key, value in sp_update.model_dump(exclude_unset=True).items():
        setattr(sp, key, value)
    
    db.commit()
    db.refresh(sp)
    return sp


@app.post("/signal-priorities/request")
def request_signal_priority(request: SignalPriorityRequest, db: Session = Depends(get_db)):
    result = check_and_request_signal_priority(
        db, request.vehicle_id, request.distance, request.intersection_name
    )
    return result


@app.post("/operation-parameters/", response_model=OperationParameterResponse)
def create_operation_parameter(op: OperationParameterCreate, db: Session = Depends(get_db)):
    existing = db.query(OperationParameter).filter(
        OperationParameter.route_id == op.route_id
    ).first()
    
    if existing:
        raise HTTPException(status_code=400, detail="该线路已有运营参数配置")
    
    db_op = OperationParameter(**op.model_dump())
    db.add(db_op)
    db.commit()
    db.refresh(db_op)
    return db_op


@app.get("/operation-parameters/{route_id}", response_model=OperationParameterResponse)
def get_operation_parameter(route_id: int, db: Session = Depends(get_db)):
    op = db.query(OperationParameter).filter(OperationParameter.route_id == route_id).first()
    if not op:
        raise HTTPException(status_code=404, detail="运营参数配置不存在")
    return op


@app.put("/operation-parameters/{op_id}", response_model=OperationParameterResponse)
def update_operation_parameter(op_id: int, op_update: OperationParameterUpdate, db: Session = Depends(get_db)):
    op = db.query(OperationParameter).filter(OperationParameter.id == op_id).first()
    if not op:
        raise HTTPException(status_code=404, detail="运营参数配置不存在")
    
    for key, value in op_update.model_dump(exclude_unset=True).items():
        setattr(op, key, value)
    
    db.commit()
    db.refresh(op)
    return op


@app.post("/schedules/generate")
def generate_schedules(request: ScheduleGenerationRequest, db: Session = Depends(get_db)):
    try:
        date = datetime.strptime(request.date, "%Y-%m-%d")
        schedules = generate_schedule_for_route(db, request.route_id, date)
        return {
            "message": f"成功生成 {len(schedules)} 个排班",
            "schedules": [s.id for s in schedules]
        }
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/schedules/", response_model=List[ScheduleResponse])
def get_schedules(
    route_id: Optional[int] = None,
    date: Optional[str] = None,
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db)
):
    query = db.query(Schedule)
    
    if route_id:
        query = query.filter(Schedule.route_id == route_id)
    
    if date:
        try:
            date_obj = datetime.strptime(date, "%Y-%m-%d")
            date_start = date_obj.replace(hour=0, minute=0, second=0, microsecond=0)
            date_end = date_start + timedelta(days=1)
            query = query.filter(
                Schedule.start_time >= date_start,
                Schedule.start_time < date_end
            )
        except ValueError:
            raise HTTPException(status_code=400, detail="日期格式错误，应为 YYYY-MM-DD")
    
    return query.order_by(Schedule.start_time).offset(skip).limit(limit).all()


@app.get("/schedules/{schedule_id}", response_model=ScheduleDetailResponse)
def get_schedule_detail(schedule_id: int, db: Session = Depends(get_db)):
    schedule = db.query(Schedule).filter(Schedule.id == schedule_id).first()
    if not schedule:
        raise HTTPException(status_code=404, detail="排班不存在")
    
    schedule_data = json.loads(schedule.schedule_data)
    stations = []
    
    for station_data in schedule_data["stations"]:
        stations.append(ScheduleStation(
            station_id=station_data["station_id"],
            station_name=station_data["station_name"],
            sequence=station_data["sequence"],
            planned_arrival_time=datetime.fromisoformat(station_data["planned_arrival_time"])
        ))
    
    return ScheduleDetailResponse(
        schedule=schedule,
        stations=stations
    )


@app.post("/arrival-records/")
def record_arrival(record: ArrivalRecordCreate, db: Session = Depends(get_db)):
    arrival_record = db.query(ArrivalRecord).filter(
        ArrivalRecord.schedule_id == record.schedule_id,
        ArrivalRecord.station_id == record.station_id
    ).first()
    
    if not arrival_record:
        raise HTTPException(status_code=404, detail="到站记录不存在")
    
    arrival_record.actual_arrival_time = record.actual_arrival_time
    arrival_record.is_on_time = is_on_time(
        arrival_record.planned_arrival_time,
        record.actual_arrival_time
    )
    
    db.commit()
    db.refresh(arrival_record)
    
    return {
        "message": "到站记录已更新",
        "record": ArrivalRecordResponse(
            id=arrival_record.id,
            schedule_id=arrival_record.schedule_id,
            station_id=arrival_record.station_id,
            planned_arrival_time=arrival_record.planned_arrival_time,
            actual_arrival_time=arrival_record.actual_arrival_time,
            is_on_time=arrival_record.is_on_time,
            created_at=arrival_record.created_at
        )
    }


@app.get("/punctuality/", response_model=List[PunctualityReport])
def get_punctuality(
    route_id: Optional[int] = None,
    start_date: Optional[str] = None,
    end_date: Optional[str] = None,
    db: Session = Depends(get_db)
):
    start_dt = datetime.strptime(start_date, "%Y-%m-%d") if start_date else None
    end_dt = datetime.strptime(end_date, "%Y-%m-%d") if end_date else None
    
    reports = get_punctuality_report(db, route_id, start_dt, end_dt)
    
    return [
        PunctualityReport(
            route_id=r["route_id"],
            route_name=r["route_name"],
            total_schedules=r["total_schedules"],
            on_time_count=r["on_time_count"],
            punctuality_rate=r["punctuality_rate"]
        )
        for r in reports
    ]


@app.get("/signal-priorities/vehicle/{vehicle_id}/today", response_model=List[SignalPriorityRequestResponse])
def get_vehicle_signal_requests_today(vehicle_id: int, db: Session = Depends(get_db)):
    today = datetime.now().date()
    requests = db.query(SignalPriorityRequest).filter(
        SignalPriorityRequest.vehicle_id == vehicle_id,
        SignalPriorityRequest.request_time >= datetime.combine(today, datetime.min.time())
    ).order_by(SignalPriorityRequest.request_time.desc()).all()
    
    return requests


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", "8000"))
    uvicorn.run("app.main:app", host="0.0.0.0", port=port, reload=True)
