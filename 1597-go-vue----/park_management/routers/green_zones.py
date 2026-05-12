from datetime import datetime, timedelta
from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from ..database import get_db
from ..models import (
    Park, GreenZone, GradeAdjustment, MaintenanceTask,
    MaintenanceGrade, TaskStatus, get_maintenance_frequency
)
from .. import schemas

router = APIRouter(prefix="/parks")


def get_park_by_code(db: Session, park_code: str):
    park = db.query(Park).filter(Park.park_code == park_code).first()
    if not park:
        raise HTTPException(status_code=404, detail=f"Park with code {park_code} not found")
    return park


def get_green_zone_by_code(db: Session, park_id: int, zone_code: str):
    zone = db.query(GreenZone).filter(
        GreenZone.park_id == park_id,
        GreenZone.zone_code == zone_code
    ).first()
    if not zone:
        raise HTTPException(
            status_code=404, 
            detail=f"Green zone {zone_code} not found in park"
        )
    return zone


def generate_task_code():
    return f"TASK-{datetime.now().strftime('%Y%m%d%H%M%S')}"


def generate_next_task(db: Session, green_zone: GreenZone, scheduled_date: Optional[datetime] = None):
    if scheduled_date is None:
        latest_task = db.query(MaintenanceTask).filter(
            MaintenanceTask.green_zone_id == green_zone.id
        ).order_by(MaintenanceTask.scheduled_date.desc()).first()
        
        if latest_task:
            frequency = get_maintenance_frequency(latest_task.maintenance_grade)
            scheduled_date = latest_task.scheduled_date + timedelta(days=frequency)
        else:
            frequency = get_maintenance_frequency(green_zone.current_grade)
            scheduled_date = datetime.utcnow() + timedelta(days=frequency)
    
    task_code = generate_task_code()
    new_task = MaintenanceTask(
        task_code=task_code,
        green_zone_id=green_zone.id,
        maintenance_grade=green_zone.current_grade,
        scheduled_date=scheduled_date,
        status=TaskStatus.PENDING
    )
    db.add(new_task)
    db.commit()
    db.refresh(new_task)
    return new_task


def check_and_upgrade_grade(db: Session, green_zone: GreenZone):
    if green_zone.consecutive_missed_tasks >= 2:
        old_grade = green_zone.current_grade
        if old_grade == MaintenanceGrade.LEVEL_3:
            new_grade = MaintenanceGrade.LEVEL_2
        elif old_grade == MaintenanceGrade.LEVEL_2:
            new_grade = MaintenanceGrade.LEVEL_1
        else:
            return
        
        green_zone.current_grade = new_grade
        green_zone.consecutive_missed_tasks = 0
        
        adjustment = GradeAdjustment(
            green_zone_id=green_zone.id,
            old_grade=old_grade,
            new_grade=new_grade,
            reason="连续两次未按时完成养护任务",
            adjusted_by="SYSTEM"
        )
        db.add(adjustment)
        db.commit()


@router.post("/", response_model=schemas.Park, status_code=201)
def create_park(park: schemas.ParkCreate, db: Session = Depends(get_db)):
    existing = db.query(Park).filter(Park.park_code == park.park_code).first()
    if existing:
        raise HTTPException(status_code=400, detail=f"Park code {park.park_code} already exists")
    
    db_park = Park(**park.model_dump())
    db.add(db_park)
    db.commit()
    db.refresh(db_park)
    return db_park


@router.get("/", response_model=List[schemas.Park])
def list_parks(db: Session = Depends(get_db)):
    return db.query(Park).all()


@router.post("/{park_code}/green-zones/", response_model=schemas.GreenZone, status_code=201)
def create_green_zone(
    park_code: str,
    zone: schemas.GreenZoneCreate,
    db: Session = Depends(get_db)
):
    park = get_park_by_code(db, park_code)
    
    existing = db.query(GreenZone).filter(GreenZone.zone_code == zone.zone_code).first()
    if existing:
        raise HTTPException(status_code=400, detail=f"Zone code {zone.zone_code} already exists")
    
    zone_data = zone.model_dump(exclude={"park_code"})
    db_zone = GreenZone(park_id=park.id, **zone_data)
    db.add(db_zone)
    db.commit()
    db.refresh(db_zone)
    
    generate_next_task(db, db_zone)
    
    return db_zone


@router.get("/{park_code}/green-zones/", response_model=List[schemas.GreenZone])
def list_green_zones(park_code: str, db: Session = Depends(get_db)):
    park = get_park_by_code(db, park_code)
    return db.query(GreenZone).filter(GreenZone.park_id == park.id).all()


@router.get("/{park_code}/green-zones/{zone_code}", response_model=schemas.GreenZone)
def get_green_zone(
    park_code: str,
    zone_code: str,
    db: Session = Depends(get_db)
):
    park = get_park_by_code(db, park_code)
    return get_green_zone_by_code(db, park.id, zone_code)


@router.post("/{park_code}/green-zones/{zone_code}/grade/adjust", response_model=schemas.GreenZone)
def adjust_green_zone_grade(
    park_code: str,
    zone_code: str,
    adjustment: schemas.GradeAdjustmentCreate,
    db: Session = Depends(get_db)
):
    park = get_park_by_code(db, park_code)
    zone = get_green_zone_by_code(db, park.id, zone_code)
    
    old_grade = zone.current_grade
    if old_grade == adjustment.new_grade:
        raise HTTPException(status_code=400, detail="New grade is the same as current grade")
    
    db_adjustment = GradeAdjustment(
        green_zone_id=zone.id,
        old_grade=old_grade,
        new_grade=adjustment.new_grade,
        reason=adjustment.reason,
        season=adjustment.season,
        adjusted_by=adjustment.adjusted_by
    )
    db.add(db_adjustment)
    
    zone.current_grade = adjustment.new_grade
    zone.consecutive_missed_tasks = 0
    db.commit()
    db.refresh(zone)
    
    return zone


@router.get("/{park_code}/green-zones/{zone_code}/grade/history", response_model=List[schemas.GradeAdjustment])
def get_grade_history(
    park_code: str,
    zone_code: str,
    db: Session = Depends(get_db)
):
    park = get_park_by_code(db, park_code)
    zone = get_green_zone_by_code(db, park.id, zone_code)
    
    return db.query(GradeAdjustment).filter(
        GradeAdjustment.green_zone_id == zone.id
    ).order_by(GradeAdjustment.adjusted_at.desc()).all()


@router.post("/{park_code}/green-zones/{zone_code}/maintenance/tasks", response_model=schemas.MaintenanceTask, status_code=201)
def create_maintenance_task(
    park_code: str,
    zone_code: str,
    task_data: schemas.MaintenanceTaskCreate,
    db: Session = Depends(get_db)
):
    park = get_park_by_code(db, park_code)
    zone = get_green_zone_by_code(db, park.id, zone_code)
    
    return generate_next_task(db, zone, task_data.scheduled_date)


@router.get("/{park_code}/green-zones/{zone_code}/maintenance/tasks", response_model=List[schemas.MaintenanceTask])
def list_maintenance_tasks(
    park_code: str,
    zone_code: str,
    status: Optional[TaskStatus] = Query(None),
    db: Session = Depends(get_db)
):
    park = get_park_by_code(db, park_code)
    zone = get_green_zone_by_code(db, park.id, zone_code)
    
    query = db.query(MaintenanceTask).filter(MaintenanceTask.green_zone_id == zone.id)
    if status:
        query = query.filter(MaintenanceTask.status == status)
    
    return query.order_by(MaintenanceTask.scheduled_date.desc()).all()


@router.post("/{park_code}/green-zones/{zone_code}/maintenance/record", response_model=schemas.MaintenanceTask)
def record_maintenance(
    park_code: str,
    zone_code: str,
    record: schemas.MaintenanceTaskRecord,
    task_id: Optional[int] = Query(None, description="Task ID to update. If not provided, uses latest pending task."),
    db: Session = Depends(get_db)
):
    park = get_park_by_code(db, park_code)
    zone = get_green_zone_by_code(db, park.id, zone_code)
    
    if task_id:
        task = db.query(MaintenanceTask).filter(
            MaintenanceTask.id == task_id,
            MaintenanceTask.green_zone_id == zone.id
        ).first()
        if not task:
            raise HTTPException(status_code=404, detail="Maintenance task not found")
    else:
        task = db.query(MaintenanceTask).filter(
            MaintenanceTask.green_zone_id == zone.id,
            MaintenanceTask.status.in_([TaskStatus.PENDING, TaskStatus.IN_PROGRESS, TaskStatus.OVERDUE])
        ).order_by(MaintenanceTask.scheduled_date.asc()).first()
        
        if not task:
            raise HTTPException(
                status_code=404, 
                detail="No pending maintenance tasks found for this green zone"
            )
    
    if record.status == TaskStatus.COMPLETED:
        task.actual_completion_date = record.actual_completion_date or datetime.utcnow()
        
        if task.actual_completion_date > task.scheduled_date:
            zone.consecutive_missed_tasks += 1
        else:
            zone.consecutive_missed_tasks = 0
        
        check_and_upgrade_grade(db, zone)
        generate_next_task(db, zone)
    
    task.status = record.status
    task.executed_by = record.executed_by
    task.notes = record.notes
    
    db.commit()
    db.refresh(task)
    return task


@router.put("/{park_code}/green-zones/{zone_code}/maintenance/tasks/{task_id}", response_model=schemas.MaintenanceTask)
def update_maintenance_task(
    park_code: str,
    zone_code: str,
    task_id: int,
    record: schemas.MaintenanceTaskRecord,
    db: Session = Depends(get_db)
):
    park = get_park_by_code(db, park_code)
    zone = get_green_zone_by_code(db, park.id, zone_code)
    
    task = db.query(MaintenanceTask).filter(
        MaintenanceTask.id == task_id,
        MaintenanceTask.green_zone_id == zone.id
    ).first()
    
    if not task:
        raise HTTPException(status_code=404, detail="Maintenance task not found")
    
    if record.status == TaskStatus.COMPLETED and task.status != TaskStatus.COMPLETED:
        task.actual_completion_date = record.actual_completion_date or datetime.utcnow()
        
        if task.actual_completion_date > task.scheduled_date:
            zone.consecutive_missed_tasks += 1
        else:
            zone.consecutive_missed_tasks = 0
        
        check_and_upgrade_grade(db, zone)
        generate_next_task(db, zone)
    
    task.status = record.status
    if record.executed_by:
        task.executed_by = record.executed_by
    if record.notes:
        task.notes = record.notes
    
    db.commit()
    db.refresh(task)
    return task
