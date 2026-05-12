from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session

from .. import models, schemas, services
from ..database import get_db

router = APIRouter()


@router.post("/docks/", response_model=schemas.Dock)
def create_dock(dock: schemas.DockCreate, db: Session = Depends(get_db)):
    db_dock = models.Dock(**dock.model_dump())
    db.add(db_dock)
    db.commit()
    db.refresh(db_dock)
    return db_dock


@router.get("/docks/", response_model=List[schemas.Dock])
def list_docks(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    docks = db.query(models.Dock).offset(skip).limit(limit).all()
    return docks


@router.get("/docks/{dock_id}/availability")
def check_dock_availability_api(
    dock_id: int,
    start_date: str = Query(...),
    end_date: str = Query(...),
    db: Session = Depends(get_db),
):
    from datetime import datetime

    try:
        start = datetime.strptime(start_date, "%Y-%m-%d").date()
        end = datetime.strptime(end_date, "%Y-%m-%d").date()
    except ValueError:
        raise HTTPException(status_code=400, detail="日期格式错误，应为 YYYY-MM-DD")

    available, message = services.check_dock_availability(db, dock_id, start, end)
    return {"available": available, "message": message}


@router.post("/", response_model=schemas.Docking)
def create_docking(docking: schemas.DockingCreate, db: Session = Depends(get_db)):
    vessel = (
        db.query(models.Vessel)
        .filter(models.Vessel.id == docking.vessel_id)
        .first()
    )
    if not vessel:
        raise HTTPException(status_code=404, detail="船舶不存在")

    dock = db.query(models.Dock).filter(models.Dock.id == docking.dock_id).first()
    if not dock:
        raise HTTPException(status_code=404, detail="船坞不存在")

    available, message = services.check_dock_availability(
        db, docking.dock_id, docking.start_date, docking.end_date
    )
    if not available:
        raise HTTPException(status_code=400, detail=message)

    db_docking = models.Docking(**docking.model_dump())
    db.add(db_docking)
    db.commit()
    db.refresh(db_docking)
    return db_docking


@router.get("/", response_model=List[schemas.Docking])
def list_dockings(
    vessel_id: Optional[int] = Query(None),
    dock_id: Optional[int] = Query(None),
    status: Optional[str] = Query(None),
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db),
):
    query = db.query(models.Docking)

    if vessel_id:
        query = query.filter(models.Docking.vessel_id == vessel_id)
    if dock_id:
        query = query.filter(models.Docking.dock_id == dock_id)
    if status:
        query = query.filter(models.Docking.status == status)

    dockings = query.order_by(models.Docking.start_date.desc()).offset(skip).limit(limit).all()
    return dockings


@router.get("/{docking_id}", response_model=schemas.Docking)
def get_docking(docking_id: int, db: Session = Depends(get_db)):
    docking = (
        db.query(models.Docking)
        .filter(models.Docking.id == docking_id)
        .first()
    )
    if not docking:
        raise HTTPException(status_code=404, detail="坞修安排不存在")
    return docking


@router.put("/{docking_id}", response_model=schemas.Docking)
def update_docking(
    docking_id: int,
    docking_update: schemas.DockingUpdate,
    db: Session = Depends(get_db),
):
    docking = (
        db.query(models.Docking)
        .filter(models.Docking.id == docking_id)
        .first()
    )
    if not docking:
        raise HTTPException(status_code=404, detail="坞修安排不存在")

    update_data = docking_update.model_dump(exclude_unset=True)

    if "start_date" in update_data or "end_date" in update_data:
        new_start = update_data.get("start_date", docking.start_date)
        new_end = update_data.get("end_date", docking.end_date)

        available, message = services.check_dock_availability(
            db, docking.dock_id, new_start, new_end
        )
        if not available:
            raise HTTPException(status_code=400, detail=message)

    for key, value in update_data.items():
        setattr(docking, key, value)

    db.commit()
    db.refresh(docking)
    return docking
