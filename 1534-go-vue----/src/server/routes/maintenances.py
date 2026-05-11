from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List, Optional
from datetime import date
from server.database import get_db
from core.models import MaintenanceCreate, MaintenanceResponse
from core import services

router = APIRouter(prefix="/maintenances", tags=["maintenances"])


@router.post("", response_model=MaintenanceResponse)
def create(maintenance: MaintenanceCreate, db: Session = Depends(get_db)):
    facility = services.get_facility(db, maintenance.facility_id)
    if not facility:
        raise HTTPException(status_code=404, detail="Facility not found")
    parts_dicts = [p.model_dump() for p in maintenance.parts]
    return services.create_maintenance(
        db,
        maintenance.facility_id,
        maintenance.maintenance_date,
        maintenance.description,
        parts_dicts
    )


@router.get("", response_model=List[MaintenanceResponse])
def list_all(
    facility_id: Optional[int] = None,
    start_date: Optional[date] = None,
    end_date: Optional[date] = None,
    db: Session = Depends(get_db)
):
    return services.list_maintenances(db, facility_id, start_date, end_date)


@router.get("/{maintenance_id}", response_model=MaintenanceResponse)
def get_one(maintenance_id: int, db: Session = Depends(get_db)):
    maintenance = services.get_maintenance(db, maintenance_id)
    if not maintenance:
        raise HTTPException(status_code=404, detail="Maintenance not found")
    return maintenance


@router.get("/monthly-cost/{year}/{month}")
def monthly_cost(
    year: int,
    month: int,
    enterprise_id: Optional[int] = None,
    db: Session = Depends(get_db)
):
    return services.aggregate_monthly_maintenance_cost(db, year, month, enterprise_id)
