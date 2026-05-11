from typing import List, Optional
from fastapi import APIRouter, HTTPException, Query
from src.server.schemas import (
    BatchCreate, BatchResponse,
    BreedingRecordCreate, BreedingRecordResponse, BreedingFailureRequest,
    VaccinationRecordCreate, VaccinationRecordResponse,
    SlaughterRecordCreate, SlaughterRecordResponse,
    ReminderResponse, DashboardResponse
)
from src.core.services import farm_service

router = APIRouter()

@router.post("/batches", response_model=BatchResponse)
async def create_batch(batch: BatchCreate):
    try:
        result = farm_service.create_batch(
            breed=batch.breed,
            initial_count=batch.initial_count,
            gestation_period_days=batch.gestation_period_days
        )
        return result
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))

@router.get("/batches", response_model=List[BatchResponse])
async def list_batches():
    return farm_service.list_batches()

@router.get("/batches/{batch_id}", response_model=BatchResponse)
async def get_batch(batch_id: int):
    batch = farm_service.get_batch(batch_id)
    if not batch:
        raise HTTPException(status_code=404, detail="批次不存在")
    return batch

@router.post("/breeding", response_model=BreedingRecordResponse)
async def create_breeding_record(record: BreedingRecordCreate):
    try:
        result = farm_service.create_breeding_record(
            batch_id=record.batch_id,
            female_id=record.female_id,
            breeding_date=record.breeding_date
        )
        return result
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))

@router.get("/breeding", response_model=List[BreedingRecordResponse])
async def list_breeding_records(batch_id: Optional[int] = Query(None)):
    return farm_service.list_breeding_records(batch_id=batch_id)

@router.post("/breeding/{record_id}/failure", response_model=BreedingRecordResponse)
async def record_breeding_failure(record_id: int, request: BreedingFailureRequest):
    result = farm_service.record_breeding_failure(
        record_id=record_id,
        failure_reason=request.failure_reason
    )
    if not result:
        raise HTTPException(status_code=404, detail="配种记录不存在")
    return result

@router.post("/breeding/{record_id}/success", response_model=BreedingRecordResponse)
async def record_breeding_success(record_id: int):
    result = farm_service.record_breeding_success(record_id)
    if not result:
        raise HTTPException(status_code=404, detail="配种记录不存在")
    return result

@router.post("/vaccination", response_model=VaccinationRecordResponse)
async def create_vaccination_record(record: VaccinationRecordCreate):
    try:
        result = farm_service.create_vaccination_record(
            batch_id=record.batch_id,
            vaccine_name=record.vaccine_name,
            vaccination_date=record.vaccination_date,
            interval_days=record.interval_days
        )
        return result
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))

@router.get("/vaccination", response_model=List[VaccinationRecordResponse])
async def list_vaccination_records(batch_id: Optional[int] = Query(None)):
    return farm_service.list_vaccination_records(batch_id=batch_id)

@router.post("/slaughter", response_model=SlaughterRecordResponse)
async def create_slaughter_record(record: SlaughterRecordCreate):
    try:
        result = farm_service.create_slaughter_record(
            batch_id=record.batch_id,
            count=record.count,
            avg_weight=record.avg_weight,
            unit_price=record.unit_price,
            slaughter_date=record.slaughter_date
        )
        return result
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))

@router.get("/slaughter", response_model=List[SlaughterRecordResponse])
async def list_slaughter_records(batch_id: Optional[int] = Query(None)):
    return farm_service.list_slaughter_records(batch_id=batch_id)

@router.post("/reminders/generate", response_model=List[ReminderResponse])
async def generate_reminders():
    return farm_service.generate_reminders()

@router.get("/reminders", response_model=List[ReminderResponse])
async def list_reminders(is_read: Optional[bool] = Query(None)):
    return farm_service.list_reminders(is_read=is_read)

@router.post("/reminders/{reminder_id}/read", response_model=ReminderResponse)
async def mark_reminder_read(reminder_id: int):
    result = farm_service.mark_reminder_read(reminder_id)
    if not result:
        raise HTTPException(status_code=404, detail="提醒不存在")
    return result

@router.get("/dashboard", response_model=DashboardResponse)
async def get_dashboard():
    return farm_service.get_dashboard()
