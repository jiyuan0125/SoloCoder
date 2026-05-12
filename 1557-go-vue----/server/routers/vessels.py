from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session

from .. import models, schemas, services
from ..database import get_db

router = APIRouter()


@router.post("/", response_model=schemas.Vessel)
def create_vessel(vessel: schemas.VesselCreate, db: Session = Depends(get_db)):
    if vessel.imo_number:
        existing = (
            db.query(models.Vessel)
            .filter(models.Vessel.imo_number == vessel.imo_number)
            .first()
        )
        if existing:
            raise HTTPException(status_code=400, detail="IMO编号已存在")

    db_vessel = models.Vessel(**vessel.model_dump())
    db.add(db_vessel)
    db.commit()
    db.refresh(db_vessel)
    return db_vessel


@router.get("/", response_model=List[schemas.Vessel])
def list_vessels(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    vessels = db.query(models.Vessel).offset(skip).limit(limit).all()
    return vessels


@router.get("/{vessel_id}", response_model=schemas.Vessel)
def get_vessel(vessel_id: int, db: Session = Depends(get_db)):
    vessel = db.query(models.Vessel).filter(models.Vessel.id == vessel_id).first()
    if not vessel:
        raise HTTPException(status_code=404, detail="船舶不存在")
    return vessel


@router.put("/{vessel_id}", response_model=schemas.Vessel)
def update_vessel(
    vessel_id: int, vessel_update: schemas.VesselUpdate, db: Session = Depends(get_db)
):
    vessel = db.query(models.Vessel).filter(models.Vessel.id == vessel_id).first()
    if not vessel:
        raise HTTPException(status_code=404, detail="船舶不存在")

    update_data = vessel_update.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(vessel, key, value)

    db.commit()
    db.refresh(vessel)
    return vessel


@router.get("/{vessel_id}/inspections", response_model=List[schemas.Inspection])
def get_vessel_inspections(vessel_id: int, db: Session = Depends(get_db)):
    vessel = db.query(models.Vessel).filter(models.Vessel.id == vessel_id).first()
    if not vessel:
        raise HTTPException(status_code=404, detail="船舶不存在")

    return services.get_vessel_inspections(db, vessel_id)


@router.get("/{vessel_id}/certificates", response_model=List[schemas.Certificate])
def get_vessel_certificates(vessel_id: int, db: Session = Depends(get_db)):
    vessel = db.query(models.Vessel).filter(models.Vessel.id == vessel_id).first()
    if not vessel:
        raise HTTPException(status_code=404, detail="船舶不存在")

    return services.get_vessel_certificates(db, vessel_id)


@router.get("/{vessel_id}/dockings", response_model=List[schemas.Docking])
def get_vessel_dockings(vessel_id: int, db: Session = Depends(get_db)):
    vessel = db.query(models.Vessel).filter(models.Vessel.id == vessel_id).first()
    if not vessel:
        raise HTTPException(status_code=404, detail="船舶不存在")

    return services.get_vessel_dockings(db, vessel_id)


@router.get("/{vessel_id}/todos", response_model=List[schemas.TodoItem])
def get_vessel_todos(
    vessel_id: int,
    status: Optional[str] = Query(None),
    db: Session = Depends(get_db),
):
    vessel = db.query(models.Vessel).filter(models.Vessel.id == vessel_id).first()
    if not vessel:
        raise HTTPException(status_code=404, detail="船舶不存在")

    return services.get_vessel_todos(db, vessel_id, status)


@router.get("/{vessel_id}/next-inspection")
def get_next_inspection(vessel_id: int, db: Session = Depends(get_db)):
    vessel = db.query(models.Vessel).filter(models.Vessel.id == vessel_id).first()
    if not vessel:
        raise HTTPException(status_code=404, detail="船舶不存在")

    next_type, next_due = services.get_next_inspection_info(db, vessel_id)

    return {
        "vessel_id": vessel_id,
        "next_inspection_type": next_type,
        "next_inspection_due": next_due,
    }
