from datetime import datetime
from typing import List, Optional
from fastapi import APIRouter, HTTPException, status

from src.core.models.schemas import (
    ProductionBatch, ProductionBatchCreate, FeedingRecord,
    FeedingRecordCreate, BatchStatus
)
from src.core.storage.memory_store import get_store
from src.core.services.feeding_service import FeedingService


router = APIRouter(prefix="/batches", tags=["batches"])
store = get_store()
feeding_service = FeedingService()


@router.post("/", response_model=ProductionBatch, status_code=status.HTTP_201_CREATED)
def create_batch(data: ProductionBatchCreate):
    reactor = store.get_reactor(data.reactor_id)
    if not reactor:
        raise HTTPException(status_code=404, detail="Reactor not found")
    
    recipe = store.get_recipe(data.recipe_id)
    if not recipe:
        raise HTTPException(status_code=404, detail="Recipe not found")
    
    existing_batch = store.get_running_batch_for_reactor(data.reactor_id)
    if existing_batch:
        raise HTTPException(
            status_code=400,
            detail="Reactor already has a running batch"
        )
    
    return store.create_batch(data)


@router.get("/", response_model=List[ProductionBatch])
def list_batches(
    reactor_id: Optional[str] = None,
    status_filter: Optional[BatchStatus] = None,
    start_date: Optional[datetime] = None,
    end_date: Optional[datetime] = None
):
    return store.list_batches(reactor_id, status_filter, start_date, end_date)


@router.get("/{batch_id}", response_model=ProductionBatch)
def get_batch(batch_id: str):
    batch = store.get_batch(batch_id)
    if not batch:
        raise HTTPException(status_code=404, detail="Batch not found")
    return batch


@router.post("/{batch_id}/feed", response_model=FeedingRecord, status_code=status.HTTP_201_CREATED)
def record_feeding(batch_id: str, data: FeedingRecordCreate):
    batch = store.get_batch(batch_id)
    if not batch:
        raise HTTPException(status_code=404, detail="Batch not found")
    
    if batch.status != BatchStatus.RUNNING:
        raise HTTPException(status_code=400, detail="Batch is not running")
    
    if batch_id != data.batch_id:
        raise HTTPException(status_code=400, detail="Batch ID mismatch")
    
    recipe = store.get_recipe(batch.recipe_id)
    if not recipe:
        raise HTTPException(status_code=404, detail="Recipe not found")
    
    if not feeding_service.validate_feeding_order(batch, data.order):
        raise HTTPException(
            status_code=400,
            detail="Invalid feeding order. Must follow recipe sequence."
        )
    
    record = feeding_service.create_feeding_record(data, recipe)
    record = store.add_feeding_record(record)
    
    return record


@router.post("/{batch_id}/complete", response_model=ProductionBatch)
def complete_batch(batch_id: str):
    batch = store.get_batch(batch_id)
    if not batch:
        raise HTTPException(status_code=404, detail="Batch not found")
    
    if batch.status != BatchStatus.RUNNING:
        raise HTTPException(status_code=400, detail="Batch is not running")
    
    recipe = store.get_recipe(batch.recipe_id)
    if not recipe:
        raise HTTPException(status_code=404, detail="Recipe not found")
    
    batch = feeding_service.finalize_batch(batch, recipe)
    batch = store.update_batch(batch)
    
    return batch


@router.get("/{batch_id}/feedings", response_model=List[FeedingRecord])
def list_feedings(batch_id: str):
    batch = store.get_batch(batch_id)
    if not batch:
        raise HTTPException(status_code=404, detail="Batch not found")
    return store.list_feeding_records(batch_id)
