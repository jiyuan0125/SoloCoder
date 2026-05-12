from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from ..database import get_db
from ..models import Part, Purchase
from ..schemas import (
    PartCreate, PartUpdate, PartResponse, PartDetailResponse,
    RepairResponse, RequisitionResponse, PurchaseResponse, PurchaseHistoryResponse
)

router = APIRouter(prefix="/parts", tags=["parts"])


def get_part_or_404(db: Session, part_id: int):
    part = db.query(Part).filter(Part.id == part_id).first()
    if not part:
        raise HTTPException(status_code=404, detail="Part not found")
    return part


@router.post("", response_model=PartResponse, status_code=201)
def create_part(part_data: PartCreate, db: Session = Depends(get_db)):
    existing = db.query(Part).filter(
        Part.part_number == part_data.part_number,
        Part.serial_number == part_data.serial_number
    ).first()
    if existing:
        raise HTTPException(
            status_code=400,
            detail="Part with same part_number and serial_number already exists"
        )

    part = Part(
        part_number=part_data.part_number,
        serial_number=part_data.serial_number,
        name=part_data.name,
        unit_price=part_data.unit_price,
        is_controllable=part_data.is_controllable,
        minimum_stock=part_data.minimum_stock,
        available_quantity=part_data.available_quantity
    )
    db.add(part)
    db.commit()
    db.refresh(part)

    return part


@router.get("", response_model=List[PartResponse])
def list_parts(
    skip: int = 0,
    limit: int = 100,
    part_number: Optional[str] = None,
    status: Optional[str] = None,
    db: Session = Depends(get_db)
):
    query = db.query(Part)
    if part_number:
        query = query.filter(Part.part_number.ilike(f"%{part_number}%"))
    if status:
        query = query.filter(Part.status == status)
    return query.offset(skip).limit(limit).all()


@router.get("/{part_id}", response_model=PartDetailResponse)
def get_part(part_id: int, db: Session = Depends(get_db)):
    return get_part_or_404(db, part_id)


@router.put("/{part_id}", response_model=PartResponse)
def update_part(part_id: int, part_data: PartUpdate, db: Session = Depends(get_db)):
    part = get_part_or_404(db, part_id)
    update_data = part_data.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(part, key, value)
    db.commit()
    db.refresh(part)
    return part


@router.delete("/{part_id}", status_code=204)
def delete_part(part_id: int, db: Session = Depends(get_db)):
    part = get_part_or_404(db, part_id)
    db.delete(part)
    db.commit()


@router.get("/{part_id}/repairs", response_model=List[RepairResponse])
def get_part_repairs(part_id: int, db: Session = Depends(get_db)):
    part = get_part_or_404(db, part_id)
    return part.repairs


@router.get("/{part_id}/requisitions", response_model=List[RequisitionResponse])
def get_part_requisitions(part_id: int, db: Session = Depends(get_db)):
    part = get_part_or_404(db, part_id)
    return part.requisitions


@router.get("/{part_id}/purchases", response_model=List[PurchaseResponse])
def get_part_purchases(part_id: int, db: Session = Depends(get_db)):
    part = get_part_or_404(db, part_id)
    return part.purchases


@router.post("/{part_id}/scrap", response_model=PartResponse)
def scrap_part(part_id: int, db: Session = Depends(get_db)):
    part = get_part_or_404(db, part_id)
    part.status = "scrapped"
    part.available_quantity = 0
    db.commit()
    db.refresh(part)
    return part
