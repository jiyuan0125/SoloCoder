from datetime import date
from typing import List, Optional

from fastapi import APIRouter, Depends, HTTPException

from src.core.models import Vaccine, VaccineRecord
from src.server.deps import get_vaccine_service

router = APIRouter(prefix="/api", tags=["vaccines"])


@router.post("/vaccines", response_model=Vaccine)
def create_vaccine(
    name: str,
    manufacturer: Optional[str] = None,
    recommended_interval_days: Optional[int] = None,
    vaccine_service=Depends(get_vaccine_service),
):
    return vaccine_service.create_vaccine(
        name=name,
        manufacturer=manufacturer,
        recommended_interval_days=recommended_interval_days,
    )


@router.get("/vaccines", response_model=List[Vaccine])
def list_vaccines(vaccine_service=Depends(get_vaccine_service)):
    return vaccine_service.list_vaccines()


@router.post("/vaccinations", response_model=VaccineRecord)
def record_vaccination(
    pet_id: str,
    vaccine_id: str,
    inoculation_date: date,
    doctor_id: Optional[str] = None,
    batch_number: Optional[str] = None,
    next_inoculation_date: Optional[date] = None,
    remarks: Optional[str] = None,
    vaccine_service=Depends(get_vaccine_service),
):
    try:
        return vaccine_service.record_vaccination(
            pet_id=pet_id,
            vaccine_id=vaccine_id,
            inoculation_date=inoculation_date,
            doctor_id=doctor_id,
            batch_number=batch_number,
            next_inoculation_date=next_inoculation_date,
            remarks=remarks,
        )
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/pets/{pet_id}/vaccinations", response_model=List[VaccineRecord])
def get_vaccine_history(
    pet_id: str,
    vaccine_service=Depends(get_vaccine_service),
):
    return vaccine_service.get_vaccine_history(pet_id)
