from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from typing import List, Optional
from datetime import datetime, date
from app.database import get_db
from app.models import CrewType
from app.schemas.models import DutyRecordCreate, DutyRecordUpdate, DutyRecordResponse, MonthlyWorkHoursResponse
from app.services.crew_service import DutyRulesService
from app.services.optimization_service import WorkHoursService

router = APIRouter()


@router.post("/duty-records", response_model=DutyRecordResponse, status_code=status.HTTP_201_CREATED)
def create_duty_record(record: DutyRecordCreate, db: Session = Depends(get_db)):
    db_record = __import__("app.models", fromlist=["DutyRecord"]).DutyRecord(**record.model_dump())
    db.add(db_record)
    db.commit()
    db.refresh(db_record)
    return db_record


@router.get("/duty-records", response_model=List[DutyRecordResponse])
def list_duty_records(crew_member_id: Optional[int] = None, is_completed: Optional[bool] = None, 
                      db: Session = Depends(get_db)):
    from app.models import DutyRecord
    query = db.query(DutyRecord)
    if crew_member_id:
        query = query.filter(DutyRecord.crew_member_id == crew_member_id)
    if is_completed is not None:
        query = query.filter(DutyRecord.is_completed == is_completed)
    return query.order_by(DutyRecord.start_time.desc()).all()


@router.post("/duty-records/{record_id}/complete", response_model=DutyRecordResponse)
def complete_duty_record(record_id: int, end_time: Optional[datetime] = None, db: Session = Depends(get_db)):
    try:
        return WorkHoursService.complete_duty_record(db, record_id, end_time)
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))


@router.get("/monthly-hours/{member_id}/{year}/{month}", response_model=MonthlyWorkHoursResponse)
def get_monthly_hours(member_id: int, year: int, month: int, db: Session = Depends(get_db)):
    return DutyRulesService.get_monthly_work_hours(db, member_id, year, month)


@router.get("/monthly-report/{year}/{month}")
def get_monthly_report(year: int, month: int, crew_type: Optional[CrewType] = None, db: Session = Depends(get_db)):
    return WorkHoursService.get_monthly_report(db, year, month, crew_type)


@router.post("/recalculate/{member_id}/{year}/{month}")
def recalculate_monthly_hours(member_id: int, year: int, month: int, db: Session = Depends(get_db)):
    from app.services.delay_service import DelayService
    DelayService.recalculate_monthly_hours(db, member_id)
    return {"status": "success", "message": "月度工时已重新核算"}
