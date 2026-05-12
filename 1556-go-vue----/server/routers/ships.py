from datetime import date
from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from sqlalchemy import and_

from ..config import get_db
from ..models import Ship, Assignment, Crew, AttendanceRecord
from ..schemas import (
    ShipCreate, ShipUpdate, ShipResponse, ShipDetailResponse,
    AssignmentCreate, AssignmentUpdate, AssignmentResponse,
    CrewCertificateStatus, CrewResponse, MonthlyAttendanceSummary
)
from ..services import (
    validate_assignment, check_crew_certificates_valid,
    update_on_board_status, check_ship_capacity,
    get_monthly_attendance_summary
)

router = APIRouter(prefix="/api", tags=["ships"])


@router.post("/ships/", response_model=ShipResponse, status_code=status.HTTP_201_CREATED)
def create_ship(ship_data: ShipCreate, db: Session = Depends(get_db)):
    existing = db.query(Ship).filter(Ship.name == ship_data.name).first()
    if existing:
        raise HTTPException(status_code=400, detail="船舶名称已存在")
    
    if ship_data.imo_number:
        existing_imo = db.query(Ship).filter(Ship.imo_number == ship_data.imo_number).first()
        if existing_imo:
            raise HTTPException(status_code=400, detail="IMO编号已存在")
    
    ship = Ship(**ship_data.model_dump())
    db.add(ship)
    db.commit()
    db.refresh(ship)
    return ship


@router.get("/ships/", response_model=List[ShipResponse])
def list_ships(db: Session = Depends(get_db)):
    return db.query(Ship).all()


@router.get("/ships/{ship_id}", response_model=ShipDetailResponse)
def get_ship_detail(ship_id: int, db: Session = Depends(get_db)):
    ship = db.query(Ship).filter(Ship.id == ship_id).first()
    if not ship:
        raise HTTPException(status_code=404, detail="船舶不存在")
    
    active_assignments = db.query(Assignment).filter(
        Assignment.ship_id == ship_id,
        Assignment.is_active == True
    ).all()
    
    certificate_statuses = []
    for assignment in active_assignments:
        all_valid, cert_summary = check_crew_certificates_valid(db, assignment.crew_id)
        certificate_statuses.append(CrewCertificateStatus(
            crew_id=assignment.crew_id,
            crew_name=assignment.crew.name,
            has_expired_certificates=not all_valid,
            certificate_summary=cert_summary
        ))
    
    today = date.today()
    attendance_summary = MonthlyAttendanceSummary(
        year=today.year,
        month=today.month,
        normal_days=0.0,
        overtime_days=0.0,
        sick_leave_days=0.0,
        personal_leave_days=0.0,
        total_days=0.0
    )
    
    for assignment in active_assignments:
        crew_summary = get_monthly_attendance_summary(db, assignment.crew_id, today.year, today.month)
        attendance_summary.normal_days += crew_summary["normal_days"]
        attendance_summary.overtime_days += crew_summary["overtime_days"]
        attendance_summary.sick_leave_days += crew_summary["sick_leave_days"]
        attendance_summary.personal_leave_days += crew_summary["personal_leave_days"]
        attendance_summary.total_days += crew_summary["total_days"]
    
    assignment_responses = []
    for a in active_assignments:
        crew_resp = CrewResponse.model_validate(a.crew) if a.crew else None
        ship_resp = ShipResponse.model_validate(a.ship) if a.ship else None
        assignment_responses.append(AssignmentResponse(
            id=a.id,
            crew_id=a.crew_id,
            ship_id=a.ship_id,
            position=a.position,
            is_watch_keeper=a.is_watch_keeper,
            start_date=a.start_date,
            end_date=a.end_date,
            is_active=a.is_active,
            created_at=a.created_at,
            updated_at=a.updated_at,
            ship=ship_resp,
            crew=crew_resp
        ))
    
    return ShipDetailResponse(
        id=ship.id,
        name=ship.name,
        imo_number=ship.imo_number,
        capacity=ship.capacity,
        description=ship.description,
        created_at=ship.created_at,
        updated_at=ship.updated_at,
        crew_list=assignment_responses,
        attendance_summary=[attendance_summary],
        certificate_statuses=certificate_statuses
    )


@router.put("/ships/{ship_id}", response_model=ShipResponse)
def update_ship(ship_id: int, ship_data: ShipUpdate, db: Session = Depends(get_db)):
    ship = db.query(Ship).filter(Ship.id == ship_id).first()
    if not ship:
        raise HTTPException(status_code=404, detail="船舶不存在")
    
    update_data = ship_data.model_dump(exclude_unset=True)
    
    if "name" in update_data:
        existing = db.query(Ship).filter(
            Ship.name == update_data["name"],
            Ship.id != ship_id
        ).first()
        if existing:
            raise HTTPException(status_code=400, detail="船舶名称已存在")
    
    if "imo_number" in update_data and update_data["imo_number"]:
        existing_imo = db.query(Ship).filter(
            Ship.imo_number == update_data["imo_number"],
            Ship.id != ship_id
        ).first()
        if existing_imo:
            raise HTTPException(status_code=400, detail="IMO编号已存在")
    
    for field, value in update_data.items():
        setattr(ship, field, value)
    
    db.commit()
    db.refresh(ship)
    return ship


@router.delete("/ships/{ship_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_ship(ship_id: int, db: Session = Depends(get_db)):
    ship = db.query(Ship).filter(Ship.id == ship_id).first()
    if not ship:
        raise HTTPException(status_code=404, detail="船舶不存在")
    
    active_assignments = db.query(Assignment).filter(
        Assignment.ship_id == ship_id,
        Assignment.is_active == True
    ).count()
    
    if active_assignments > 0:
        raise HTTPException(status_code=400, detail="船舶上还有在船船员，不能删除")
    
    db.delete(ship)
    db.commit()


@router.post("/assignments/", response_model=AssignmentResponse, status_code=status.HTTP_201_CREATED)
def create_assignment(assign_data: AssignmentCreate, db: Session = Depends(get_db)):
    is_valid, errors = validate_assignment(
        db=db,
        crew_id=assign_data.crew_id,
        ship_id=assign_data.ship_id,
        is_watch_keeper=assign_data.is_watch_keeper,
        start_date=assign_data.start_date
    )
    
    if not is_valid:
        raise HTTPException(status_code=400, detail="; ".join(errors))
    
    assignment = Assignment(**assign_data.model_dump())
    db.add(assignment)
    db.commit()
    db.refresh(assignment)
    
    if assign_data.start_date <= date.today():
        update_on_board_status(db, assign_data.crew_id, assignment.id, assign_data.start_date)
    
    return assignment


@router.get("/assignments/", response_model=List[AssignmentResponse])
def list_assignments(ship_id: Optional[int] = None, 
                     crew_id: Optional[int] = None,
                     is_active: Optional[bool] = None,
                     db: Session = Depends(get_db)):
    query = db.query(Assignment)
    
    if ship_id:
        query = query.filter(Assignment.ship_id == ship_id)
    if crew_id:
        query = query.filter(Assignment.crew_id == crew_id)
    if is_active is not None:
        query = query.filter(Assignment.is_active == is_active)
    
    return query.all()


@router.get("/assignments/{assignment_id}", response_model=AssignmentResponse)
def get_assignment(assignment_id: int, db: Session = Depends(get_db)):
    assignment = db.query(Assignment).filter(Assignment.id == assignment_id).first()
    if not assignment:
        raise HTTPException(status_code=404, detail="分配记录不存在")
    return assignment


@router.put("/assignments/{assignment_id}", response_model=AssignmentResponse)
def update_assignment(assignment_id: int, assign_data: AssignmentUpdate, 
                      db: Session = Depends(get_db)):
    assignment = db.query(Assignment).filter(Assignment.id == assignment_id).first()
    if not assignment:
        raise HTTPException(status_code=404, detail="分配记录不存在")
    
    update_data = assign_data.model_dump(exclude_unset=True)
    
    if "is_active" in update_data and update_data["is_active"] is False:
        if "end_date" not in update_data:
            update_data["end_date"] = date.today()
    
    for field, value in update_data.items():
        setattr(assignment, field, value)
    
    db.commit()
    db.refresh(assignment)
    return assignment


@router.post("/ships/{ship_id}/vacancy-check")
def check_ship_vacancy(ship_id: int, db: Session = Depends(get_db)):
    ship = db.query(Ship).filter(Ship.id == ship_id).first()
    if not ship:
        raise HTTPException(status_code=404, detail="船舶不存在")
    
    has_vacancy = check_ship_capacity(db, ship_id)
    active_count = db.query(Assignment).filter(
        Assignment.ship_id == ship_id,
        Assignment.is_active == True
    ).count()
    
    return {
        "ship_id": ship_id,
        "ship_name": ship.name,
        "capacity": ship.capacity,
        "current_crew": active_count,
        "vacancies": ship.capacity - active_count,
        "has_vacancy": has_vacancy
    }
