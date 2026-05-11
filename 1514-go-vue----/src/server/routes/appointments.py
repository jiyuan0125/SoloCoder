from datetime import date, time
from typing import List, Optional

from fastapi import APIRouter, Depends, HTTPException

from src.core.models import Appointment, WorkSlot
from src.server.deps import get_appointment_service

router = APIRouter(prefix="/api", tags=["appointments", "work-slots"])


@router.post("/work-slots", response_model=WorkSlot)
def create_work_slot(
    doctor_id: str,
    slot_date: date,
    start_time: time,
    end_time: time,
    max_appointments: int = 1,
    appointment_service=Depends(get_appointment_service),
):
    try:
        return appointment_service.create_work_slot(
            doctor_id=doctor_id,
            slot_date=slot_date,
            start_time=start_time,
            end_time=end_time,
            max_appointments=max_appointments,
        )
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/work-slots", response_model=List[WorkSlot])
def list_work_slots(
    doctor_id: Optional[str] = None,
    slot_date: Optional[date] = None,
    appointment_service=Depends(get_appointment_service),
):
    return appointment_service.list_work_slots(
        doctor_id=doctor_id, slot_date=slot_date
    )


@router.post("/appointments", response_model=Appointment)
def create_appointment(
    owner_id: str,
    pet_id: str,
    work_slot_id: str,
    remarks: Optional[str] = None,
    appointment_service=Depends(get_appointment_service),
):
    try:
        return appointment_service.create_appointment(
            owner_id=owner_id,
            pet_id=pet_id,
            work_slot_id=work_slot_id,
            remarks=remarks,
        )
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/appointments", response_model=List[Appointment])
def list_appointments(
    owner_id: Optional[str] = None,
    appointment_date: Optional[date] = None,
    appointment_service=Depends(get_appointment_service),
):
    return appointment_service.list_appointments(
        owner_id=owner_id, appointment_date=appointment_date
    )


@router.post("/appointments/{appointment_id}/cancel", response_model=Appointment)
def cancel_appointment(
    appointment_id: str,
    appointment_service=Depends(get_appointment_service),
):
    try:
        return appointment_service.cancel_appointment(appointment_id)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
