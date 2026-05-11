from typing import List, Optional

from fastapi import APIRouter, Body, Depends, HTTPException
from pydantic import BaseModel

from src.core.models import (
    Diagnosis,
    FeeRecord,
    PrescriptionItem,
    Registration,
    Treatment,
)
from src.server.deps import (
    get_diagnosis_service,
    get_registration_service,
)


class DiagnosisCreateRequest(BaseModel):
    doctor_id: str
    diagnosis: str
    prescription_items: Optional[List[PrescriptionItem]] = None
    treatments: Optional[List[Treatment]] = None
    remarks: Optional[str] = None


router = APIRouter(prefix="/api", tags=["registrations", "diagnosis", "fees"])


@router.post("/registrations", response_model=Registration)
def create_registration(
    pet_id: str,
    owner_id: str,
    symptoms: Optional[str] = None,
    doctor_id: Optional[str] = None,
    registration_service=Depends(get_registration_service),
):
    try:
        return registration_service.create_registration(
            pet_id=pet_id,
            owner_id=owner_id,
            symptoms=symptoms,
            doctor_id=doctor_id,
        )
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/registrations", response_model=List[Registration])
def list_registrations(registration_service=Depends(get_registration_service)):
    return registration_service.list_registrations()


@router.get("/registrations/waiting", response_model=List[Registration])
def get_waiting_queue(registration_service=Depends(get_registration_service)):
    return registration_service.get_waiting_queue()


@router.get("/registrations/{registration_id}", response_model=Registration)
def get_registration(
    registration_id: str,
    registration_service=Depends(get_registration_service),
):
    registration = registration_service.get_registration(registration_id)
    if not registration:
        raise HTTPException(status_code=404, detail="挂号记录不存在")
    return registration


@router.post("/registrations/{registration_id}/start", response_model=Registration)
def start_treatment(
    registration_id: str,
    doctor_id: str,
    registration_service=Depends(get_registration_service),
):
    try:
        return registration_service.start_treatment(
            registration_id=registration_id,
            doctor_id=doctor_id,
        )
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.post("/registrations/{registration_id}/complete", response_model=Registration)
def complete_registration(
    registration_id: str,
    registration_service=Depends(get_registration_service),
):
    try:
        return registration_service.complete_registration(registration_id)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.post("/registrations/{registration_id}/cancel", response_model=Registration)
def cancel_registration(
    registration_id: str,
    registration_service=Depends(get_registration_service),
):
    try:
        return registration_service.cancel_registration(registration_id)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.post("/registrations/{registration_id}/diagnosis", response_model=Diagnosis)
def create_diagnosis(
    registration_id: str,
    request: DiagnosisCreateRequest,
    diagnosis_service=Depends(get_diagnosis_service),
):
    try:
        return diagnosis_service.create_diagnosis(
            registration_id=registration_id,
            doctor_id=request.doctor_id,
            diagnosis_text=request.diagnosis,
            prescription_items=request.prescription_items,
            treatments=request.treatments,
            remarks=request.remarks,
        )
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/registrations/{registration_id}/fee", response_model=FeeRecord)
def get_fee_record(
    registration_id: str,
    diagnosis_service=Depends(get_diagnosis_service),
):
    fee = diagnosis_service.get_fee_record(registration_id)
    if not fee:
        raise HTTPException(status_code=404, detail="费用记录不存在")
    return fee


@router.post("/fees/{fee_id}/pay", response_model=FeeRecord)
def pay_fee(
    fee_id: str,
    diagnosis_service=Depends(get_diagnosis_service),
):
    try:
        return diagnosis_service.pay_fee(fee_id)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/fees/unpaid", response_model=List[FeeRecord])
def list_unpaid_fees(diagnosis_service=Depends(get_diagnosis_service)):
    return diagnosis_service.list_unpaid_fees()
