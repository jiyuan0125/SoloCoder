from datetime import datetime
from typing import List, Optional
from fastapi import FastAPI, Depends, HTTPException, Query
from fastapi.responses import JSONResponse
from sqlalchemy.orm import Session
from .database import get_db, init_db
from .config import settings
from .models import (
    Alert, AlertStatus, AlertPriority,
    RescueForce, ForceStatus,
    Event, EventStatus, EventLevel,
    EventForceAssignment, TimelineEntry, RescueResult
)
from .schemas import (
    AlertCreate, AlertVerify, AlertResponse,
    ForceCreate, ForceResponse,
    EventResponse, EventDetailResponse,
    DispatchForceRequest, ForceArrivalRequest,
    CompleteRescueRequest, MonthlyStatsResponse,
    EventLevelUpdate
)
from .services import (
    check_urgent_alert, create_event_for_alert,
    get_available_forces, dispatch_force,
    mark_force_arrived, complete_event,
    get_monthly_stats, reject_alert
)

app = FastAPI(
    title="搜救协调管理系统",
    description="搜求协调管理系统 - 报警、事件、力量调度全流程管理",
    version="1.0.0"
)


@app.on_event("startup")
async def startup_event():
    init_db()


@app.get("/")
def root():
    return {"message": "搜救协调管理系统 API", "version": "1.0.0"}


@app.post("/alerts", response_model=AlertResponse, summary="录入报警信息")
def create_alert(
    alert_data: AlertCreate,
    db: Session = Depends(get_db)
):
    alert = Alert(
        location=alert_data.location,
        location_lat=alert_data.location_lat,
        location_lon=alert_data.location_lon,
        description=alert_data.description,
        reporter_name=alert_data.reporter_name,
        reporter_contact=alert_data.reporter_contact,
        alert_time=alert_data.alert_time or datetime.utcnow()
    )
    
    db.add(alert)
    db.flush()
    
    if check_urgent_alert(db, alert):
        alert.priority = AlertPriority.URGENT
    
    db.commit()
    db.refresh(alert)
    
    return alert


@app.get("/alerts", response_model=List[AlertResponse], summary="获取报警列表")
def list_alerts(
    status: Optional[AlertStatus] = None,
    db: Session = Depends(get_db)
):
    query = db.query(Alert)
    if status:
        query = query.filter(Alert.status == status)
    return query.order_by(Alert.alert_time.desc()).all()


@app.get("/alerts/{alert_id}", response_model=AlertResponse, summary="获取单个报警")
def get_alert(alert_id: int, db: Session = Depends(get_db)):
    alert = db.query(Alert).filter(Alert.id == alert_id).first()
    if not alert:
        raise HTTPException(status_code=404, detail="报警信息不存在")
    return alert


@app.post("/alerts/{alert_id}/verify", response_model=AlertResponse, summary="核实报警信息")
def verify_alert(
    alert_id: int,
    verify_data: AlertVerify,
    db: Session = Depends(get_db)
):
    alert = db.query(Alert).filter(Alert.id == alert_id).first()
    if not alert:
        raise HTTPException(status_code=404, detail="报警信息不存在")
    
    if alert.status != AlertStatus.RECEIVED:
        raise HTTPException(status_code=400, detail="报警信息已核实或已拒绝")
    
    time_elapsed = (datetime.utcnow() - alert.alert_time).total_seconds() / 60
    if time_elapsed > settings.VERIFICATION_TIME_LIMIT_MINUTES:
        raise HTTPException(
            status_code=400,
            detail=f"报警超过 {settings.VERIFICATION_TIME_LIMIT_MINUTES} 分钟未核实"
        )
    
    if verify_data.accept:
        if alert.event_id:
            event = db.query(Event).filter(Event.id == alert.event_id).first()
            if event and event.status == EventStatus.RECEIVED:
                event.status = EventStatus.VERIFIED
                event.verified_at = datetime.utcnow()
            
            alert.status = AlertStatus.VERIFIED
            alert.verified_at = datetime.utcnow()
            alert.verified_by = verify_data.verified_by
            alert.verification_notes = verify_data.verification_notes
            
            db.commit()
            db.refresh(alert)
            return alert
        
        create_event_for_