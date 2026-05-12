from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session

from .. import models, schemas, services
from ..database import get_db
from ..config import INSPECTION_TYPES

router = APIRouter()


@router.post("/", response_model=schemas.Inspection)
def create_inspection(
    inspection: schemas.InspectionCreate, db: Session = Depends(get_db)
):
    if inspection.inspection_type not in INSPECTION_TYPES:
        raise HTTPException(
            status_code=400,
            detail=f"无效的检验类型，可选值: {list(INSPECTION_TYPES.keys())}",
        )

    vessel = (
        db.query(models.Vessel)
        .filter(models.Vessel.id == inspection.vessel_id)
        .first()
    )
    if not vessel:
        raise HTTPException(status_code=404, detail="船舶不存在")

    db_inspection = models.Inspection(**inspection.model_dump())
    db.add(db_inspection)
    db.commit()
    db.refresh(db_inspection)

    services.generate_inspection_todos(db, inspection.vessel_id)

    return db_inspection


@router.get("/", response_model=List[schemas.Inspection])
def list_inspections(
    vessel_id: Optional[int] = Query(None),
    inspection_type: Optional[str] = Query(None),
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db),
):
    query = db.query(models.Inspection)

    if vessel_id:
        query = query.filter(models.Inspection.vessel_id == vessel_id)
    if inspection_type:
        query = query.filter(models.Inspection.inspection_type == inspection_type)

    inspections = query.order_by(models.Inspection.inspection_date.desc()).offset(skip).limit(limit).all()
    return inspections


@router.get("/{inspection_id}", response_model=schemas.Inspection)
def get_inspection(inspection_id: int, db: Session = Depends(get_db)):
    inspection = (
        db.query(models.Inspection)
        .filter(models.Inspection.id == inspection_id)
        .first()
    )
    if not inspection:
        raise HTTPException(status_code=404, detail="检验记录不存在")
    return inspection


@router.put("/{inspection_id}/complete", response_model=schemas.Inspection)
def complete_inspection(inspection_id: int, result: str = "passed", db: Session = Depends(get_db)):
    try:
        return services.complete_inspection(db, inspection_id, result)
    except ValueError as e:
        raise HTTPException(status_code=404, detail=str(e))


@router.put("/{inspection_id}", response_model=schemas.Inspection)
def update_inspection(
    inspection_id: int,
    inspection_update: schemas.InspectionUpdate,
    db: Session = Depends(get_db),
):
    inspection = (
        db.query(models.Inspection)
        .filter(models.Inspection.id == inspection_id)
        .first()
    )
    if not inspection:
        raise HTTPException(status_code=404, detail="检验记录不存在")

    update_data = inspection_update.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(inspection, key, value)

    db.commit()
    db.refresh(inspection)
    return inspection
