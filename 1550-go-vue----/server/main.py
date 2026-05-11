from datetime import datetime, timedelta
from typing import List, Optional
from fastapi import FastAPI, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from sqlalchemy import func, and_
from .database import engine, Base, get_db
from .models import (
    Station, StationType, TideMeasurement, WaveMeasurement, WaterTempMeasurement,
    SalinityMeasurement, DissolvedOxygenMeasurement, ChlorophyllMeasurement,
    RedtideEvent, RedtideStatus, WaveAlertLevel, Farmer, Notification
)
from .schemas import (
    Station as StationSchema, StationCreate,
    TideMeasurement as TideMeasurementSchema, TideDailyStats,
    WaveMeasurement as WaveMeasurementSchema,
    RedtideEvent as RedtideEventSchema, RedtideEventUpdate,
    Farmer as FarmerSchema, FarmerCreate,
    Notification as NotificationSchema,
    StationDataInput
)
from .services import (
    calculate_wave_alert_level, is_salinity_sensor_ok,
    check_redtide_suspected, check_redtide_resolved,
    create_redtide_event, update_redtide_event_status
)
from .config import settings

Base.metadata.create_all(bind=engine)

app = FastAPI(title="海洋环境监测系统", version="1.0.0")


@app.get("/")
def read_root():
    return {"name": "海洋环境监测系统", "version": "1.0.0", "status": "running"}


@app.post("/stations/", response_model=StationSchema)
def create_station(station: StationCreate, db: Session = Depends(get_db)):
    db_station = db.query(Station).filter(Station.id == station.id).first()
    if db_station:
        raise HTTPException(status_code=400, detail="监测站ID已存在")
    
    db_station = Station(
        id=station.id,
        name=station.name,
        station_type=station.station_type,
        location=station.location,
        sea_area=station.sea_area
    )
    db.add(db_station)
    db.commit()
    db.refresh(db_station)
    return db_station


@app.get("/stations/", response_model=List[StationSchema])
def list_stations(db: Session = Depends(get_db)):
    return db.query(Station).all()


@app.get("/stations/{station_id}", response_model=StationSchema)
def get_station(station_id: str, db: Session = Depends(get_db)):
    station = db.query(Station).filter(Station.id == station_id).first()
    if not station:
        raise HTTPException(status_code=404, detail="监测站不存在")
    return station


@app.post("/stations/{station_id}/data")
def post_station_data(station_id: str, data: StationDataInput, db: Session = Depends(get_db)):
    station = db.query(Station).filter(Station.id == station_id).first()
    if not station:
        raise HTTPException(status_code=404, detail="监测站不存在")
    
    if data.tide is not None:
        db.add(TideMeasurement(station_id=station_id, measured_at=data.measured_at, tide_level=data.tide))
    
    if data.wave_height is not None:
        db.add(WaveMeasurement(
            station_id=station_id,
            measured_at=data.measured_at,
            significant_wave_height=data.wave_height,
            wave_period=data.wave_period,
            wave_direction=data.wave_direction
        ))
    
    if data.water_temp is not None:
        db.add(WaterTempMeasurement(station_id=station_id, measured_at=data.measured_at, temperature=data.water_temp))
    
    if data.salinity is not None:
        db.add(SalinityMeasurement(station_id=station_id, measured_at=data.measured_at, salinity=data.salinity))
    
    if data.dissolved_oxygen is not None:
        db.add(DissolvedOxygenMeasurement(station_id=station_id, measured_at=data.measured_at, dissolved_oxygen=data.dissolved_oxygen))
    
    if data.chlorophyll is not None:
        db.add(ChlorophyllMeasurement(station_id=station_id, measured_at=data.measured_at, chlorophyll=data.chlorophyll))
        
        active_event = db.query(RedtideEvent).filter(
            RedtideEvent.station_id == station_id,
            RedtideEvent.status.notin_([RedtideStatus.RESOLVED])
        ).first()
        
        if active_event:
            if active_event.status in [RedtideStatus.PUBLISHED, RedtideStatus.MONITORING]:
                if check_redtide_resolved(db, station_id, data.measured_at):
                    update_redtide_event_status(db, active_event.id, RedtideStatus.RESOLVED)
                else:
                    if active_event.status == RedtideStatus.PUBLISHED:
                        update_redtide_event_status(db, active_event.id, RedtideStatus.MONITORING)
        else:
            if check_redtide_suspected(db, station_id, data.measured_at):
                create_redtide_event(db, station_id)
    
    db.commit()
    return {"status": "success", "message": "数据已入库"}


@app.get("/stations/{station_id}/tide")
def get_station_tide(
    station_id: str,
    start_time: Optional[datetime] = Query(None, description="开始时间"),
    end_time: Optional[datetime] = Query(None, description="结束时间"),
    view: str = Query("raw", description="视图模式: raw(原始数据)或 daily(日统计)"),
    db: Session = Depends(get_db)
):
    station = db.query(Station).filter(Station.id == station_id).first()
    if not station:
        raise HTTPException(status_code=404, detail="监测站不存在")
    
    query = db.query(TideMeasurement).filter(TideMeasurement.station_id == station_id)
    
    if start_time:
        query = query.filter(TideMeasurement.measured_at >= start_time)
    if end_time:
        query = query.filter(TideMeasurement.measured_at <= end_time)
    
    if view == "daily":
        measurements = query.order_by(TideMeasurement.measured_at).all()
        
        daily_data = {}
        for m in measurements:
            date_key = m.measured_at.date().isoformat()
            if date_key not in daily_data:
                daily_data[date_key] = []
            daily_data[date_key].append(m.tide_level)
        
        result = []
        for date, levels in sorted(daily_data.items()):
            result.append(TideDailyStats(
                date=date,
                max_tide=max(levels),
                min_tide=min(levels),
                avg_tide=sum(levels) / len(levels)
            ))
        return {"station_id": station_id, "view": "daily", "data": result}
    else:
        measurements = query.order_by(TideMeasurement.measured_at.desc()).all()
        hourly_data = []
        last_hour = None
        
        for m in sorted(measurements, key=lambda x: x.measured_at, reverse=True):
            hour_key = m.measured_at.replace(minute=0, second=0, microsecond=0)
            if hour_key != last_hour:
                hourly_data.append(m)
                last_hour = hour_key
        
        return {
            "station_id": station_id,
            "view": "hourly",
            "data": hourly_data
        }


@app.get("/stations/{station_id}/wave")
def get_station_wave(
    station_id: str,
    start_time: Optional[datetime] = Query(None),
    end_time: Optional[datetime] = Query(None),
    db: Session = Depends(get_db)
):
    station = db.query(Station).filter(Station.id == station_id).first()
    if not station:
        raise HTTPException(status_code=404, detail="监测站不存在")
    
    query = db.query(WaveMeasurement).filter(WaveMeasurement.station_id == station_id)
    
    if start_time:
        query = query.filter(WaveMeasurement.measured_at >= start_time)
    if end_time:
        query = query.filter(WaveMeasurement.measured_at <= end_time)
    
    measurements = query.order_by(WaveMeasurement.measured_at.desc()).all()
    
    result = []
    for m in measurements:
        result.append({
            "id": m.id,
            "station_id": m.station_id,
            "measured_at": m.measured_at,
            "significant_wave_height": m.significant_wave_height,
            "wave_period": m.wave_period,
            "wave_direction": m.wave_direction,
            "alert_level": calculate_wave_alert_level(m.significant_wave_height)
        })
    
    max_alert = WaveAlertLevel.NONE
    if result:
        max_height = max(m["significant_wave_height"] for m in result)
        max_alert = calculate_wave_alert_level(max_height)
    
    return {
        "station_id": station_id,
        "current_alert_level": max_alert,
        "data": result
    }


@app.get("/stations/{station_id}/redtide")
def get_station_redtide(
    station_id: str,
    db: Session = Depends(get_db)
):
    station = db.query(Station).filter(Station.id == station_id).first()
    if not station:
        raise HTTPException(status_code=404, detail="监测站不存在")
    
    events = db.query(RedtideEvent).filter(
        RedtideEvent.station_id == station_id
    ).order_by(RedtideEvent.suspected_at.desc()).all()
    
    active_event = None
    for event in events:
        if event.status != RedtideStatus.RESOLVED:
            active_event = event
            break
    
    recent_chlorophyll = db.query(ChlorophyllMeasurement).filter(
        ChlorophyllMeasurement.station_id == station_id
    ).order_by(ChlorophyllMeasurement.measured_at.desc()).limit(10).all()
    
    return {
        "station_id": station_id,
        "active_event": active_event,
        "history": events,
        "recent_chlorophyll": recent_chlorophyll
    }


@app.get("/redtide-events/", response_model=List[RedtideEventSchema])
def list_redtide_events(
    status: Optional[RedtideStatus] = Query(None),
    db: Session = Depends(get_db)
):
    query = db.query(RedtideEvent)
    if status:
        query = query.filter(RedtideEvent.status == status)
    return query.order_by(RedtideEvent.suspected_at.desc()).all()


@app.patch("/redtide-events/{event_id}", response_model=RedtideEventSchema)
def update_redtide_event(
    event_id: int,
    update: RedtideEventUpdate,
    db: Session = Depends(get_db)
):
    event = db.query(RedtideEvent).filter(RedtideEvent.id == event_id).first()
    if not event:
        raise HTTPException(status_code=404, detail="赤潮事件不存在")
    
    if update.status:
        event = update_redtide_event_status(db, event_id, update.status)
        if not event:
            raise HTTPException(status_code=400, detail="状态更新失败")
    
    if update.description:
        event.description = update.description
        db.commit()
        db.refresh(event)
    
    return event


@app.post("/farmers/", response_model=FarmerSchema)
def create_farmer(farmer: FarmerCreate, db: Session = Depends(get_db)):
    db_farmer = Farmer(
        name=farmer.name,
        phone=farmer.phone,
        email=farmer.email,
        sea_area=farmer.sea_area,
        farm_name=farmer.farm_name
    )
    db.add(db_farmer)
    db.commit()
    db.refresh(db_farmer)
    return db_farmer


@app.get("/farmers/", response_model=List[FarmerSchema])
def list_farmers(sea_area: Optional[str] = Query(None), db: Session = Depends(get_db)):
    query = db.query(Farmer)
    if sea_area:
        query = query.filter(Farmer.sea_area == sea_area)
    return query.all()


@app.get("/notifications/", response_model=List[NotificationSchema])
def list_notifications(
    farmer_id: Optional[int] = Query(None),
    is_read: Optional[bool] = Query(None),
    db: Session = Depends(get_db)
):
    query = db.query(Notification)
    if farmer_id:
        query = query.filter(Notification.farmer_id == farmer_id)
    if is_read is not None:
        query = query.filter(Notification.is_read == is_read)
    return query.order_by(Notification.sent_at.desc()).all()
