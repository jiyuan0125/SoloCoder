from datetime import datetime
from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from ..database import get_db
from ..models import Part, Repair
from ..schemas import RepairCreate, RepairUpdate, RepairResponse

router = APIRouter(prefix="/repairs", tags=["repairs"])


@router.post("", response_model=RepairResponse, status_code=201)
def create_repair(
    repair_data: RepairCreate,
    db: Session = Depends(get_db)
):
    part = db.query(Part).filter(Part.id == repair_data.part_id).first()
    if not part:
        raise HTTPException(status_code=404, detail="Part not found")

    repair = Repair(
        part_id=repair_data.part_id,
        quantity=repair_data.quantity,
        repair_vendor=repair_data.repair_vendor,
        repair_cost=repair_data.repair_cost
    )
    db.add(repair)
    db.commit()
    db.refresh(repair)

    return repair


@router.get("", response_model=List[RepairResponse])
def list_repairs(
    skip: int = 0,
    limit: int = 100,
    status: Optional[str] = None,
    overdue: Optional[bool] = None,
    part_id: Optional[int] = None,
    db: Session = Depends(get_db)
):
    query = db.query(Repair)
    if status:
        query = query.filter(Repair.status == status)
    if part_id:
        query = query.filter(Repair.part_id == part_id)

    repairs = query.offset(skip).limit(limit).all()

    if overdue is not None:
        repairs = [r for r in repairs if r.is_overdue == overdue]

    return repairs


@router.get("/{repair_id}", response_model=RepairResponse)
def get_repair(repair_id: int, db: Session = Depends(get_db)):
    repair = db.query(Repair).filter(Repair.id == repair_id).first()
    if not repair:
        raise HTTPException(status_code=404, detail="Repair not found")
    return repair


@router.put("/{repair_id}", response_model=RepairResponse)
def update_repair(
    repair_id: int,
    repair_data: RepairUpdate,
    db: Session = Depends(get_db)
):
    repair = db.query(Repair).filter(Repair.id == repair_id).first()
    if not repair:
        raise HTTPException(status_code=404, detail="Repair not found")

    update_data = repair_data.model_dump(exclude_unset=True)

    if "status" in update_data and update_data["status"] == "completed":
        repair.status = "completed"
        repair.completed_at = datetime.utcnow()
        update_data.pop("status", None)

    for key, value in update_data.items():
        setattr(repair, key, value)

    repair.last_updated_at = datetime.utcnow()
    db.commit()
    db.refresh(repair)

    return repair


@router.post("/{repair_id}/complete", response_model=RepairResponse)
def complete_repair(
    repair_id: int,
    repair_data: Optional[RepairUpdate] = None,
    db: Session = Depends(get_db)
):
    repair = db.query(Repair).filter(Repair.id == repair_id).first()
    if not repair:
        raise HTTPException(status_code=404, detail="Repair not found")

    if repair.status == "completed":
        raise HTTPException(status_code=400, detail="Repair already completed")

    repair.status = "completed"
    repair.completed_at = datetime.utcnow()
    repair.last_updated_at = datetime.utcnow()

    if repair_data:
        if repair_data.repair_vendor:
            repair.repair_vendor = repair_data.repair_vendor
        if repair_data.repair_cost is not None:
            repair.repair_cost = repair_data.repair_cost

    db.commit()
    db.refresh(repair)

    return repair
