from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List, Optional
from datetime import datetime, date
from server.database import get_db
from core.models import EmissionCreate, EmissionResponse
from core import services

router = APIRouter(prefix="/emissions", tags=["emissions"])


@router.post("", response_model=EmissionResponse)
def create(emission: EmissionCreate, db: Session = Depends(get_db)):
    facility = services.get_facility(db, emission.facility_id)
    if not facility:
        raise HTTPException(status_code=404, detail="Facility not found")
    record = services.create_emission(
        db,
        emission.facility_id,
        emission.recorded_at,
        emission.pollutant,
        emission.value,
        emission.limit_value
    )
    response_data = {
        "id": record.id,
        "facility_id": record.facility_id,
        "recorded_at": record.recorded_at,
        "pollutant": record.pollutant,
        "value": record.value,
        "limit_value": record.limit_value,
        "is_compliant": services.is_emission_compliant(record),
        "created_at": record.created_at
    }
    return response_data


@router.get("", response_model=List[EmissionResponse])
def list_all(
    facility_id: Optional[int] = None,
    start_time: Optional[datetime] = None,
    end_time: Optional[datetime] = None,
    db: Session = Depends(get_db)
):
    records = services.list_emissions(db, facility_id, start_time, end_time)
    return [
        {
            "id": r.id,
            "facility_id": r.facility_id,
            "recorded_at": r.recorded_at,
            "pollutant": r.pollutant,
            "value": r.value,
            "limit_value": r.limit_value,
            "is_compliant": services.is_emission_compliant(r),
            "created_at": r.created_at
        }
        for r in records
    ]


@router.get("/hourly-aggregate/{facility_id}/{pollutant}")
def hourly_aggregate(
    facility_id: int,
    pollutant: str,
    hour_start: datetime,
    db: Session = Depends(get_db)
):
    return services.aggregate_hourly_emissions(db, facility_id, pollutant, hour_start)


@router.get("/daily-compliance/{facility_id}")
def daily_compliance(
    facility_id: int,
    day: date,
    db: Session = Depends(get_db)
):
    rate = services.calculate_daily_compliance_rate(db, facility_id, day)
    return {"facility_id": facility_id, "day": day, "compliance_rate": rate}
