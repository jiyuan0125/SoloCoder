from datetime import date, datetime
from typing import Optional, List
from fastapi import FastAPI, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from app.database import engine, Base, get_db
from app.models import RopewayStatus, Operator, GondolaCapacity, WeatherData, StatusRecord
from app.schemas import (
    WeatherDataCreate, WeatherDataResponse,
    StatusRecordResponse,
    OperatorCreate, OperatorResponse,
    InspectionCreate, InspectionUpdate, InspectionResponse,
    GondolaCapacityCreate, GondolaCapacityResponse,
    PassengerRecordCreate, PassengerRecordResponse,
    QueueDataCreate, QueueDataResponse,
    DailyReportResponse,
    CurrentStatusResponse, RecoveryConfirm, StatusTransitionResponse
)
from app.services.weather_service import WeatherMonitor
from app.services.inspection_service import InspectionService
from app.services.passenger_service import PassengerService
from app.services.report_service import ReportService
from app.config import LAG_PROTECTION_MINUTES, MIN_OPERATING_HOURS

Base.metadata.create_all(bind=engine)

app = FastAPI(title="索道运行管理系统", version="1.0.0")


@app.get("/", tags=["系统"])
def root():
    return {"message": "索道运行管理系统 API", "status": "running"}


@app.get("/api/status", response_model=CurrentStatusResponse, tags=["状态监控"])
def get_current_status(db: Session = Depends(get_db)):
    monitor = WeatherMonitor(db)
    current_status = monitor.current_status
    
    latest_weather = db.query(WeatherData).order_by(
        WeatherData.timestamp.desc()
    ).first()
    
    latest_status = db.query(StatusRecord).order_by(
        StatusRecord.timestamp.desc()
    ).first()
    
    from app.services.weather_service import _lag_counters
    
    return CurrentStatusResponse(
        current_status=current_status,
        last_update=latest_status.timestamp if latest_status else None,
        current_wind_speed=latest_weather.wind_speed if latest_weather else None,
        has_lightning=latest_weather.has_lightning if latest_weather else False,
        pending_confirmation=monitor.pending_confirmation,
        lag_counter=_lag_counters.get('global', 0),
        lag_threshold=LAG_PROTECTION_MINUTES
    )


@app.post("/api/weather", response_model=WeatherDataResponse, tags=["天气监控"])
def update_weather(data: WeatherDataCreate, db: Session = Depends(get_db)):
    monitor = WeatherMonitor(db)
    weather, status = monitor.update_weather(
        wind_speed=data.wind_speed,
        has_lightning=data.has_lightning,
        temperature=data.temperature,
        humidity=data.humidity
    )
    return weather


@app.get("/api/weather", response_model=List[WeatherDataResponse], tags=["天气监控"])
def get_recent_weather(minutes: int = 30, db: Session = Depends(get_db)):
    monitor = WeatherMonitor(db)
    return monitor.get_recent_weather(minutes=minutes)


@app.post("/api/status/recover", response_model=StatusTransitionResponse, tags=["状态监控"])
def confirm_recovery(data: RecoveryConfirm, db: Session = Depends(get_db)):
    monitor = WeatherMonitor(db)
    operator = db.query(Operator).filter(Operator.id == data.operator_id).first()
    if not operator:
        raise HTTPException(status_code=404, detail="操作员不存在")
    
    record = monitor.confirm_recovery(
        operator_id=data.operator_id,
        notes=data.notes
    )
    
    if not record:
        return StatusTransitionResponse(
            success=False,
            new_status=monitor.current_status,
            reason="天气条件不满足恢复要求，或当前已是正常状态"
        )
    
    return StatusTransitionResponse(
        success=True,
        new_status=RopewayStatus.NORMAL,
        reason=record.reason
    )


@app.get("/api/status/history", response_model=List[StatusRecordResponse], tags=["状态监控"])
def get_status_history(limit: int = 20, db: Session = Depends(get_db)):
    monitor = WeatherMonitor(db)
    return monitor.get_status_history(limit=limit)


@app.post("/api/operators", response_model=OperatorResponse, tags=["操作员管理"])
def create_operator(data: OperatorCreate, db: Session = Depends(get_db)):
    existing = db.query(Operator).filter(Operator.username == data.username).first()
    if existing:
        raise HTTPException(status_code=400, detail="用户名已存在")
    
    operator = Operator(username=data.username, full_name=data.full_name)
    db.add(operator)
    db.commit()
    db.refresh(operator)
    return operator


@app.get("/api/operators", response_model=List[OperatorResponse], tags=["操作员管理"])
def list_operators(db: Session = Depends(get_db)):
    return db.query(Operator).all()


@app.post("/api/inspections", response_model=InspectionResponse, tags=["检修管理"])
def create_inspection(data: InspectionCreate, db: Session = Depends(get_db)):
    service = InspectionService(db)
    return service.create_inspection(
        inspection_type=data.inspection_type,
        scheduled_date=data.scheduled_date,
        notes=data.notes
    )


@app.patch("/api/inspections/{inspection_id}", response_model=InspectionResponse, tags=["检修管理"])
def update_inspection(inspection_id: int, data: InspectionUpdate, db: Session = Depends(get_db)):
    service = InspectionService(db)
    inspection = service.update_inspection(
        inspection_id,
        completed_date=data.completed_date,
        status=data.status,
        is_urgent=data.is_urgent,
        resolved=data.resolved,
        issues_found=data.issues_found,
        notes=data.notes,
        operator_id=data.operator_id
    )
    if not inspection:
        raise HTTPException(status_code=404, detail="检修记录不存在")
    return inspection


@app.get("/api/inspections", response_model=List[InspectionResponse], tags=["检修管理"])
def list_inspections(pending_only: bool = True, db: Session = Depends(get_db)):
    service = InspectionService(db)
    if pending_only:
        return service.get_pending_inspections()
    return db.query(Operator).all()


@app.get("/api/inspections/tomorrow-check", tags=["检修管理"])
def check_tomorrow_operation(db: Session = Depends(get_db)):
    service = InspectionService(db)
    can_operate = service.can_operate_tomorrow()
    urgent = service.get_urgent_inspections()
    
    return {
        "can_operate": can_operate,
        "urgent_inspections_count": len(urgent),
        "message": "允许运营" if can_operate else "存在未处理的紧急问题，次日不允许运营"
    }


@app.post("/api/capacity", response_model=GondolaCapacityResponse, tags=["载客管理"])
def set_gondola_capacity(data: GondolaCapacityCreate, db: Session = Depends(get_db)):
    service = PassengerService(db)
    return service.set_capacity(data.total_capacity)


@app.get("/api/capacity", tags=["载客管理"])
def get_current_capacity(db: Session = Depends(get_db)):
    service = PassengerService(db)
    capacity = db.query(GondolaCapacity).filter(
        GondolaCapacity.is_active == True
    ).order_by(GondolaCapacity.effective_date.desc()).first()
    
    if not capacity:
        return {"total_capacity": 0, "is_active": False}
    
    return {
        "id": capacity.id,
        "total_capacity": capacity.total_capacity,
        "is_active": capacity.is_active,
        "effective_date": capacity.effective_date
    }


@app.post("/api/passengers", response_model=PassengerRecordResponse, tags=["载客管理"])
def record_passengers(data: PassengerRecordCreate, db: Session = Depends(get_db)):
    service = PassengerService(db)
    return service.record_passengers(
        gondola_id=data.gondola_id,
        adult_count=data.adult_count,
        child_count=data.child_count
    )


@app.post("/api/queue", response_model=QueueDataResponse, tags=["排队管理"])
def update_queue(data: QueueDataCreate, db: Session = Depends(get_db)):
    service = PassengerService(db)
    queue_data, is_limited = service.update_queue(data.queue_count)
    return queue_data


@app.get("/api/queue/status", tags=["排队管理"])
def get_queue_status(db: Session = Depends(get_db)):
    service = PassengerService(db)
    return service.get_current_queue_status()


@app.post("/api/reports/generate", response_model=DailyReportResponse, tags=["运营报告"])
def generate_report(
    report_date: Optional[date] = Query(None, description="报告日期，默认为今天"),
    db: Session = Depends(get_db)
):
    service = ReportService(db)
    return service.generate_daily_report(report_date)


@app.get("/api/reports", response_model=List[DailyReportResponse], tags=["运营报告"])
def list_reports(
    start_date: Optional[date] = Query(None),
    end_date: Optional[date] = Query(None),
    db: Session = Depends(get_db)
):
    service = ReportService(db)
    if start_date and end_date:
        return service.get_reports(start_date, end_date)
    
    today = date.today()
    return service.get_reports(today, today)


@app.get("/api/reports/{report_date}", response_model=DailyReportResponse, tags=["运营报告"])
def get_report(report_date: date, db: Session = Depends(get_db)):
    service = ReportService(db)
    report = service.get_report_by_date(report_date)
    if not report:
        raise HTTPException(status_code=404, detail="该日期的报告不存在，请先生成")
    return report


@app.get("/api/config", tags=["系统配置"])
def get_system_config():
    return {
        "wind_speed_decelerate": 8.0,
        "wind_speed_pause": 14.0,
        "wind_speed_stop": 20.0,
        "lag_protection_minutes": LAG_PROTECTION_MINUTES,
        "min_operating_hours": MIN_OPERATING_HOURS,
        "queue_capacity_multiplier": 3,
        "child_height_limit": 1.2
    }
