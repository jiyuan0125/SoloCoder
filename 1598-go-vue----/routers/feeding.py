from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List, Optional
from datetime import date
from database import get_db
from models import Feed, FeedingPlan, ShiftAssignment, FeedingRecord
from schemas import (
    FeedCreate, FeedUpdate, FeedResponse,
    FeedingPlanCreate, FeedingPlanUpdate, FeedingPlanResponse,
    ShiftAssignmentCreate, ShiftAssignmentResponse,
    FeedingRecordCreate, FeedingRecordUpdate, FeedingRecordResponse
)
from services import (
    create_feeding_record_with_stock_deduction,
    check_shift_coverage,
    get_feeding_standard_for_animal
)
from models import Animal

router = APIRouter(prefix="/api/feeding", tags=["feeding"])


@router.post("/feeds", response_model=FeedResponse)
def create_feed(feed: FeedCreate, db: Session = Depends(get_db)):
    db_feed = Feed(**feed.dict())
    db.add(db_feed)
    db.commit()
    db.refresh(db_feed)
    return db_feed


@router.get("/feeds", response_model=List[FeedResponse])
def list_feeds(name: Optional[str] = None, db: Session = Depends(get_db)):
    query = db.query(Feed)
    if name:
        query = query.filter(Feed.name.contains(name))
    return query.all()


@router.get("/feeds/{feed_id}", response_model=FeedResponse)
def get_feed(feed_id: int, db: Session = Depends(get_db)):
    feed = db.query(Feed).filter(Feed.id == feed_id).first()
    if not feed:
        raise HTTPException(status_code=404, detail="Feed not found")
    return feed


@router.put("/feeds/{feed_id}", response_model=FeedResponse)
def update_feed(feed_id: int, feed: FeedUpdate, db: Session = Depends(get_db)):
    db_feed = db.query(Feed).filter(Feed.id == feed_id).first()
    if not db_feed:
        raise HTTPException(status_code=404, detail="Feed not found")
    for key, value in feed.dict(exclude_unset=True).items():
        setattr(db_feed, key, value)
    db.commit()
    db.refresh(db_feed)
    return db_feed


@router.post("/feeds/{feed_id}/restock", response_model=FeedResponse)
def restock_feed(feed_id: int, amount: float, db: Session = Depends(get_db)):
    feed = db.query(Feed).filter(Feed.id == feed_id).first()
    if not feed:
        raise HTTPException(status_code=404, detail="Feed not found")
    feed.current_stock += amount
    db.commit()
    db.refresh(feed)
    return feed


@router.post("/plans", response_model=FeedingPlanResponse)
def create_feeding_plan(plan: FeedingPlanCreate, db: Session = Depends(get_db)):
    db_plan = FeedingPlan(**plan.dict())
    db.add(db_plan)
    db.commit()
    db.refresh(db_plan)
    return db_plan


@router.get("/plans", response_model=List[FeedingPlanResponse])
def list_feeding_plans(animal_id: Optional[int] = None, is_active: Optional[bool] = None, db: Session = Depends(get_db)):
    query = db.query(FeedingPlan)
    if animal_id:
        query = query.filter(FeedingPlan.animal_id == animal_id)
    if is_active is not None:
        query = query.filter(FeedingPlan.is_active == is_active)
    return query.all()


@router.get("/plans/{plan_id}", response_model=FeedingPlanResponse)
def get_feeding_plan(plan_id: int, db: Session = Depends(get_db)):
    plan = db.query(FeedingPlan).filter(FeedingPlan.id == plan_id).first()
    if not plan:
        raise HTTPException(status_code=404, detail="Feeding plan not found")
    return plan


@router.put("/plans/{plan_id}", response_model=FeedingPlanResponse)
def update_feeding_plan(plan_id: int, plan: FeedingPlanUpdate, db: Session = Depends(get_db)):
    db_plan = db.query(FeedingPlan).filter(FeedingPlan.id == plan_id).first()
    if not db_plan:
        raise HTTPException(status_code=404, detail="Feeding plan not found")
    for key, value in plan.dict(exclude_unset=True).items():
        setattr(db_plan, key, value)
    db.commit()
    db.refresh(db_plan)
    return db_plan


@router.get("/animals/{animal_id}/recommended-plan")
def get_recommended_feeding_plan(animal_id: int, db: Session = Depends(get_db)):
    animal = db.query(Animal).filter(Animal.id == animal_id).first()
    if not animal:
        raise HTTPException(status_code=404, detail="Animal not found")
    
    standard = get_feeding_standard_for_animal(db, animal)
    if not standard:
        raise HTTPException(status_code=404, detail="No feeding standard found for this animal")
    
    return {
        "animal_id": animal_id,
        "animal_name": animal.name,
        "weight": animal.weight,
        "health_status": animal.health_status.value,
        "feed_type": standard.feed_type,
        "daily_amount": standard.daily_amount,
        "frequency": standard.frequency,
        "per_feeding_amount": round(standard.daily_amount / standard.frequency, 2),
        "is_special": standard.is_special
    }


@router.post("/shifts", response_model=ShiftAssignmentResponse)
def create_shift_assignment(assignment: ShiftAssignmentCreate, db: Session = Depends(get_db)):
    db_assignment = ShiftAssignment(**assignment.dict())
    db.add(db_assignment)
    db.commit()
    db.refresh(db_assignment)
    return db_assignment


@router.get("/shifts", response_model=List[ShiftAssignmentResponse])
def list_shift_assignments(shift_date: Optional[date] = None, zone_id: Optional[int] = None, db: Session = Depends(get_db)):
    query = db.query(ShiftAssignment)
    if shift_date:
        query = query.filter(ShiftAssignment.shift_date == shift_date)
    if zone_id:
        query = query.filter(ShiftAssignment.zone_id == zone_id)
    return query.all()


@router.get("/shifts/coverage")
def check_shift_coverage_api(check_date: date, db: Session = Depends(get_db)):
    return check_shift_coverage(db, check_date)


@router.post("/records", response_model=FeedingRecordResponse)
def create_feeding_record(record: FeedingRecordCreate, db: Session = Depends(get_db)):
    try:
        return create_feeding_record_with_stock_deduction(db, record)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/records", response_model=List[FeedingRecordResponse])
def list_feeding_records(
    animal_id: Optional[int] = None,
    start_date: Optional[date] = None,
    end_date: Optional[date] = None,
    status: Optional[str] = None,
    db: Session = Depends(get_db)
):
    query = db.query(FeedingRecord)
    if animal_id:
        query = query.filter(FeedingRecord.animal_id == animal_id)
    if start_date:
        query = query.filter(FeedingRecord.feeding_date >= start_date)
    if end_date:
        query = query.filter(FeedingRecord.feeding_date <= end_date)
    if status:
        query = query.filter(FeedingRecord.status == status)
    return query.order_by(FeedingRecord.feeding_date.desc()).all()


@router.get("/records/{record_id}", response_model=FeedingRecordResponse)
def get_feeding_record(record_id: int, db: Session = Depends(get_db)):
    record = db.query(FeedingRecord).filter(FeedingRecord.id == record_id).first()
    if not record:
        raise HTTPException(status_code=404, detail="Feeding record not found")
    return record


@router.put("/records/{record_id}", response_model=FeedingRecordResponse)
def update_feeding_record(record_id: int, record: FeedingRecordUpdate, db: Session = Depends(get_db)):
    db_record = db.query(FeedingRecord).filter(FeedingRecord.id == record_id).first()
    if not db_record:
        raise HTTPException(status_code=404, detail="Feeding record not found")
    for key, value in record.dict(exclude_unset=True).items():
        setattr(db_record, key, value)
    db.commit()
    db.refresh(db_record)
    return db_record
