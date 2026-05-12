from datetime import datetime
from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from ..database import get_db
from ..models import Part, InventoryCount
from ..schemas import InventoryCountCreate, InventoryCountResponse

router = APIRouter(prefix="/inventory", tags=["inventory"])


@router.post("/count", response_model=InventoryCountResponse, status_code=201)
def create_inventory_count(
    count_data: InventoryCountCreate,
    db: Session = Depends(get_db)
):
    part = db.query(Part).filter(Part.id == count_data.part_id).first()
    if not part:
        raise HTTPException(status_code=404, detail="Part not found")

    difference = count_data.counted_quantity - part.available_quantity

    inventory_count = InventoryCount(
        part_id=count_data.part_id,
        counted_quantity=count_data.counted_quantity,
        system_quantity=part.available_quantity,
        difference=difference,
        reason=count_data.reason,
        adjuster=count_data.adjuster
    )
    db.add(inventory_count)

    part.available_quantity = count_data.counted_quantity

    db.commit()
    db.refresh(inventory_count)
    db.refresh(part)

    return inventory_count


@router.get("/counts", response_model=List[InventoryCountResponse])
def list_inventory_counts(
    skip: int = 0,
    limit: int = 100,
    part_id: Optional[int] = None,
    db: Session = Depends(get_db)
):
    query = db.query(InventoryCount)
    if part_id:
        query = query.filter(InventoryCount.part_id == part_id)
    return query.order_by(InventoryCount.created_at.desc()).offset(skip).limit(limit).all()


@router.get("/counts/{count_id}", response_model=InventoryCountResponse)
def get_inventory_count(count_id: int, db: Session = Depends(get_db)):
    inventory_count = db.query(InventoryCount).filter(InventoryCount.id == count_id).first()
    if not inventory_count:
        raise HTTPException(status_code=404, detail="Inventory count not found")
    return inventory_count
