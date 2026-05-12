from datetime import date
from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, status, Query
from sqlalchemy.orm import Session

from ..config import get_db
from ..models import (
    Crew, CrewStatus, AttendanceRecord, AttendanceType,
    SalaryRecord, Assignment
)
from ..schemas import (
    AttendanceRecordCreate, AttendanceRecordResponse,
    MonthlyAttendanceSummary, SalaryRecordResponse,
    SalaryRecordCreate, SalarySettlementRequest
)
from ..services import (
    get_monthly_attendance_summary, calculate_monthly_salary,
    off_board_crew
)

router = APIRouter(prefix="/api", tags=["attendance"])


@router.post("/attendance/", response_model=AttendanceRecordResponse, status_code=status.HTTP_201_CREATED)
def create_attendance_record(record_data: AttendanceRecordCreate, db: Session = Depends(get_db)):
    crew = db.query(Crew).filter(Crew.id == record_data.crew_id).first()
    if not crew:
        raise HTTPException(status_code=404, detail="船员不存在")
    
    valid_types = [e.value for e in AttendanceType]
    if record_data.attendance_type not in valid_types:
        raise HTTPException(status_code=400, detail=f"无效的考勤类型，有效类型: {valid_types}")
    
    existing = db.query(AttendanceRecord).filter(
        AttendanceRecord.crew_id == record_data.crew_id,
        AttendanceRecord.attendance_date == record_data.attendance_date
    ).first()
    
    if existing:
        existing.attendance_type = record_data.attendance_type
        existing.is_half_day = record_data.is_half_day
        if record_data.assignment_id:
            existing.assignment_id = record_data.assignment_id
        if record_data.ship_id:
            existing.ship_id = record_data.ship_id
        db.commit()
        db.refresh(existing)
        return existing
    
    record = AttendanceRecord(**record_data.model_dump())
    db.add(record)
    db.commit()
    db.refresh(record)
    return record


@router.get("/attendance/", response_model=List[AttendanceRecordResponse])
def list_attendance_records(
    crew_id: Optional[int] = None,
    ship_id: Optional[int] = None,
    start_date: Optional[date] = None,
    end_date: Optional[date] = None,
    db: Session = Depends(get_db)
):
    query = db.query(AttendanceRecord)
    
    if crew_id:
        query = query.filter(AttendanceRecord.crew_id == crew_id)
    if ship_id:
        query = query.filter(AttendanceRecord.ship_id == ship_id)
    if start_date:
        query = query.filter(AttendanceRecord.attendance_date >= start_date)
    if end_date:
        query = query.filter(AttendanceRecord.attendance_date <= end_date)
    
    return query.order_by(AttendanceRecord.attendance_date.desc()).all()


@router.get("/attendance/{record_id}", response_model=AttendanceRecordResponse)
def get_attendance_record(record_id: int, db: Session = Depends(get_db)):
    record = db.query(AttendanceRecord).filter(AttendanceRecord.id == record_id).first()
    if not record:
        raise HTTPException(status_code=404, detail="考勤记录不存在")
    return record


@router.put("/attendance/{record_id}", response_model=AttendanceRecordResponse)
def update_attendance_record(record_id: int, record_data: AttendanceRecordCreate, 
                             db: Session = Depends(get_db)):
    record = db.query(AttendanceRecord).filter(AttendanceRecord.id == record_id).first()
    if not record:
        raise HTTPException(status_code=404, detail="考勤记录不存在")
    
    valid_types = [e.value for e in AttendanceType]
    if record_data.attendance_type not in valid_types:
        raise HTTPException(status_code=400, detail=f"无效的考勤类型，有效类型: {valid_types}")
    
    update_data = record_data.model_dump()
    for field, value in update_data.items():
        setattr(record, field, value)
    
    db.commit()
    db.refresh(record)
    return record


@router.delete("/attendance/{record_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_attendance_record(record_id: int, db: Session = Depends(get_db)):
    record = db.query(AttendanceRecord).filter(AttendanceRecord.id == record_id).first()
    if not record:
        raise HTTPException(status_code=404, detail="考勤记录不存在")
    
    db.delete(record)
    db.commit()


@router.get("/attendance/summary/{crew_id}/{year}/{month}")
def get_attendance_summary(crew_id: int, year: int, month: int, db: Session = Depends(get_db)):
    crew = db.query(Crew).filter(Crew.id == crew_id).first()
    if not crew:
        raise HTTPException(status_code=404, detail="船员不存在")
    
    return get_monthly_attendance_summary(db, crew_id, year, month)


@router.post("/salary/calculate", response_model=SalaryRecordResponse)
def calculate_salary(salary_data: SalaryRecordCreate, db: Session = Depends(get_db)):
    crew = db.query(Crew).filter(Crew.id == salary_data.crew_id).first()
    if not crew:
        raise HTTPException(status_code=404, detail="船员不存在")
    
    salary = calculate_monthly_salary(
        db=db,
        crew_id=salary_data.crew_id,
        year=salary_data.year,
        month=salary_data.month,
        base_salary=salary_data.base_salary,
        ship_id=salary_data.ship_id,
        assignment_id=salary_data.assignment_id
    )
    
    return salary


@router.get("/salary/", response_model=List[SalaryRecordResponse])
def list_salary_records(
    crew_id: Optional[int] = None,
    ship_id: Optional[int] = None,
    year: Optional[int] = None,
    month: Optional[int] = None,
    db: Session = Depends(get_db)
):
    query = db.query(SalaryRecord)
    
    if crew_id:
        query = query.filter(SalaryRecord.crew_id == crew_id)
    if ship_id:
        query = query.filter(SalaryRecord.ship_id == ship_id)
    if year:
        query = query.filter(SalaryRecord.year == year)
    if month:
        query = query.filter(SalaryRecord.month == month)
    
    return query.order_by(SalaryRecord.year.desc(), SalaryRecord.month.desc()).all()


@router.get("/salary/{record_id}", response_model=SalaryRecordResponse)
def get_salary_record(record_id: int, db: Session = Depends(get_db)):
    record = db.query(SalaryRecord).filter(SalaryRecord.id == record_id).first()
    if not record:
        raise HTTPException(status_code=404, detail="工资记录不存在")
    return record


@router.post("/crew/off-board")
def off_board_crew_endpoint(
    request: SalarySettlementRequest,
    base_salary: float = Query(0.0, description="基本工资"),
    db: Session = Depends(get_db)
):
    crew = db.query(Crew).filter(Crew.id == request.crew_id).first()
    if not crew:
        raise HTTPException(status_code=404, detail="船员不存在")
    
    if crew.status != CrewStatus.ON_BOARD.value:
        raise HTTPException(status_code=400, detail="船员不在船，无法下船")
    
    try:
        off_board_crew(db, request.crew_id, request.end_date, base_salary)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
    
    salary_record = db.query(SalaryRecord).filter(
        SalaryRecord.crew_id == request.crew_id,
        SalaryRecord.year == request.end_date.year,
        SalaryRecord.month == request.end_date.month
    ).order_by(SalaryRecord.id.desc()).first()
    
    return {
        "message": "船员已成功下船，工资已结算",
        "crew_id": request.crew_id,
        "crew_name": crew.name,
        "end_date": request.end_date,
        "salary_record": SalaryRecordResponse.model_validate(salary_record) if salary_record else None
    }


@router.post("/salary/{record_id}/settle")
def settle_salary(record_id: int, db: Session = Depends(get_db)):
    record = db.query(SalaryRecord).filter(SalaryRecord.id == record_id).first()
    if not record:
        raise HTTPException(status_code=404, detail="工资记录不存在")
    
    if record.is_settled:
        raise HTTPException(status_code=400, detail="该工资记录已结算")
    
    record.is_settled = True
    record.settlement_date = date.today()
    db.commit()
    db.refresh(record)
    
    return SalaryRecordResponse.model_validate(record)


@router.get("/attendance/types")
def get_attendance_types():
    return [
        {"type": AttendanceType.NORMAL.value, "coefficient": 1.0, "description": "正常出勤"},
        {"type": AttendanceType.OVERTIME.value, "coefficient": 1.5, "description": "加班"},
        {"type": AttendanceType.SICK_LEAVE.value, "coefficient": 0.6, "description": "病假"},
        {"type": AttendanceType.PERSONAL_LEAVE.value, "coefficient": 0.0, "description": "事假（不计薪）"}
    ]
