from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List, Optional
from server.database import get_db
from core.models import FacilityCreate, FacilityResponse
from core import services

router = APIRouter(prefix="/facilities", tags=["facilities"])


@router.post("", response_model=FacilityResponse)
def create(facility: FacilityCreate, db: Session = Depends(get_db)):
    enterprise = services.get_enterprise(db, facility.enterprise_id)
    if not enterprise:
        raise HTTPException(status_code=404, detail="Enterprise not found")
    return services.create_facility(
        db,
        facility.enterprise_id,
        facility.name,
        facility.facility_type,
        facility.installed_at,
        facility.is_running
    )


@router.get("", response_model=List[FacilityResponse])
def list_all(enterprise_id: Optional[int] = None, db: Session = Depends(get_db)):
    return services.list_facilities(db, enterprise_id)


@router.get("/{facility_id}", response_model=FacilityResponse)
def get_one(facility_id: int, db: Session = Depends(get_db)):
    facility = services.get_facility(db, facility_id)
    if not facility:
        raise HTTPException(status_code=404, detail="Facility not found")
    return facility
