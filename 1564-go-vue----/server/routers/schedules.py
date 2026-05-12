from typing import List, Optional
from datetime import datetime
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session

from ..database import get_db
from ..models import Pilot as PilotModel, FlightSchedule as FlightScheduleModel
from ..schemas import (
    FlightSchedule as FlightScheduleSchema, FlightScheduleCreate, FlightScheduleUpdate,
    ValidationResult, PilotRecommendation
)
from ..validation_service import validate_flight_assignment
from ..schedule_service import (
    confirm_schedule, start_flight, complete_flight,
    recommend_pilots_for_flight, calculate_flight_duration,
    finish_rest
)
import json

router = APIRouter(prefix="/api/schedules", tags=["schedules"])


@router.get("", response_model=List[FlightScheduleSchema])
def list_schedules(
    pilot_id: Optional[int] = None,
    status: Optional[str] = None,
    start_date: Optional[datetime] = None,
    end_date: Optional[datetime] = None,
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db)
):
    query = db.query(FlightScheduleModel)
    if pilot_id:
        query = query.filter(FlightScheduleModel.pilot_id == pilot_id)
    if status:
        query = query.filter(FlightScheduleModel.status == status)
    if start_date:
        query = query.filter(FlightScheduleModel.departure_time >= start_date)
    if end_date:
        query = query.filter(FlightScheduleModel.departure_time <= end_date)
    return query.order_by(FlightScheduleModel.departure_time.desc()).offset(skip).limit(limit).all()


@router.post("/validate", response_model=ValidationResult)
def validate_schedule(
    schedule: FlightScheduleCreate,
    db: Session = Depends(get_db)
):
    pilot = db.query(PilotModel).filter(PilotModel.id == schedule.pilot_id).first()
    if not pilot:
        raise HTTPException(status_code=404, detail="飞行员不存在")
    
    flight_duration = calculate_flight_duration(schedule.departure_time, schedule.arrival_time)
    if flight_duration <= 0:
        raise HTTPException(status_code=400, detail="到达时间必须晚于起飞时间")
    
    return validate_flight_assignment(
        db, pilot, schedule.aircraft_type, schedule.is_international,
        schedule.departure_time, schedule.arrival_time, flight_duration
    )


@router.post("", response_model=FlightScheduleSchema)
def create_schedule(
    schedule: FlightScheduleCreate,
    db: Session = Depends(get_db)
):
    pilot = db.query(PilotModel).filter(PilotModel.id == schedule.pilot_id).first()
    if not pilot:
        raise HTTPException(status_code=404, detail="飞行员不存在")
    
    flight_duration = calculate_flight_duration(schedule.departure_time, schedule.arrival_time)
    if flight_duration <= 0:
        raise HTTPException(status_code=400, detail="到达时间必须晚于起飞时间")
    
    validation = validate_flight_assignment(
        db, pilot, schedule.aircraft_type, schedule.is_international,
        schedule.departure_time, schedule.arrival_time, flight_duration
    )
    
    schedule_data = schedule.model_dump()
    schedule_data["flight_duration"] = flight_duration
    schedule_data["validation_result"] = json.dumps({
        "messages": validation.messages,
        "warnings": validation.warnings
    }, ensure_ascii=False)
    schedule_data["validation_passed"] = validation.valid
    
    db_schedule = FlightScheduleModel(**schedule_data)
    db.add(db_schedule)
    db.commit()
    db.refresh(db_schedule)
    return db_schedule


@router.get("/{schedule_id}", response_model=FlightScheduleSchema)
def get_schedule(schedule_id: int, db: Session = Depends(get_db)):
    schedule = db.query(FlightScheduleModel).filter(FlightScheduleModel.id == schedule_id).first()
    if not schedule:
        raise HTTPException(status_code=404, detail="排班记录不存在")
    return schedule


@router.put("/{schedule_id}", response_model=FlightScheduleSchema)
def update_schedule(
    schedule_id: int,
    schedule_update: FlightScheduleUpdate,
    db: Session = Depends(get_db)
):
    schedule = db.query(FlightScheduleModel).filter(FlightScheduleModel.id == schedule_id).first()
    if not schedule:
        raise HTTPException(status_code=404, detail="排班记录不存在")
    
    if schedule.status in ["confirmed", "in_flight", "completed"]:
        raise HTTPException(status_code=400, detail="已确认/执行中的排班无法修改")
    
    update_data = schedule_update.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(schedule, key, value)
    
    if "departure_time" in update_data or "arrival_time" in update_data:
        schedule.flight_duration = calculate_flight_duration(schedule.departure_time, schedule.arrival_time)
    
    db.commit()
    db.refresh(schedule)
    return schedule


@router.post("/{schedule_id}/confirm", response_model=FlightScheduleSchema)
def confirm_schedule_endpoint(schedule_id: int, db: Session = Depends(get_db)):
    schedule = db.query(FlightScheduleModel).filter(FlightScheduleModel.id == schedule_id).first()
    if not schedule:
        raise HTTPException(status_code=404, detail="排班记录不存在")
    
    if schedule.status != "pending":
        raise HTTPException(status_code=400, detail="只有待确认的排班可以确认")
    
    if not schedule.validation_passed:
        raise HTTPException(status_code=400, detail="排班校验未通过，无法确认")
    
    return confirm_schedule(db, schedule)


@router.post("/{schedule_id}/start", response_model=FlightScheduleSchema)
def start_schedule_endpoint(schedule_id: int, db: Session = Depends(get_db)):
    schedule = db.query(FlightScheduleModel).filter(FlightScheduleModel.id == schedule_id).first()
    if not schedule:
        raise HTTPException(status_code=404, detail="排班记录不存在")
    
    if schedule.status != "confirmed":
        raise HTTPException(status_code=400, detail="只有已确认的排班可以开始执飞")
    
    try:
        return start_flight(db, schedule)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.post("/{schedule_id}/complete", response_model=FlightScheduleSchema)
def complete_schedule_endpoint(schedule_id: int, db: Session = Depends(get_db)):
    schedule = db.query(FlightScheduleModel).filter(FlightScheduleModel.id == schedule_id).first()
    if not schedule:
        raise HTTPException(status_code=404, detail="排班记录不存在")
    
    if schedule.status != "in_flight":
        raise HTTPException(status_code=400, detail="只有执飞中的排班可以完成")
    
    return complete_flight(db, schedule)


@router.post("/pilots/{pilot_id}/finish-rest")
def finish_pilot_rest(pilot_id: int, db: Session = Depends(get_db)):
    pilot = db.query(PilotModel).filter(PilotModel.id == pilot_id).first()
    if not pilot:
        raise HTTPException(status_code=404, detail="飞行员不存在")
    
    if pilot.status != "rest":
        raise HTTPException(status_code=400, detail="飞行员不处于休息状态")
    
    finish_rest(db, pilot)
    return {"message": "休息结束，已转为待命状态", "pilot_id": pilot_id, "status": pilot.status}


@router.delete("/{schedule_id}")
def delete_schedule(schedule_id: int, db: Session = Depends(get_db)):
    schedule = db.query(FlightScheduleModel).filter(FlightScheduleModel.id == schedule_id).first()
    if not schedule:
        raise HTTPException(status_code=404, detail="排班记录不存在")
    
    if schedule.status in ["confirmed", "in_flight", "completed"]:
        raise HTTPException(status_code=400, detail="已确认/执行中的排班无法删除")
    
    db.delete(schedule)
    db.commit()
    return {"message": "删除成功"}


@router.get("/recommend/pilots", response_model=List[PilotRecommendation])
def recommend_pilots(
    aircraft_type: str = Query(..., description="机型"),
    is_international: bool = Query(False, description="是否国际航线"),
    departure_time: datetime = Query(..., description="起飞时间"),
    arrival_time: datetime = Query(..., description="到达时间"),
    limit: int = Query(10, ge=1, le=50),
    db: Session = Depends(get_db)
):
    flight_duration = calculate_flight_duration(departure_time, arrival_time)
    if flight_duration <= 0:
        raise HTTPException(status_code=400, detail="到达时间必须晚于起飞时间")
    
    return recommend_pilots_for_flight(
        db, aircraft_type, is_international,
        departure_time, arrival_time, limit
    )
