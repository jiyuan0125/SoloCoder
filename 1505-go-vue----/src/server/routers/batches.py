from datetime import datetime
from typing import List, Optional

from fastapi import APIRouter, Depends, HTTPException, Query, status
from pydantic import BaseModel

from core.models import Batch, BatchCreate, ValidationError
from core.services import ProductionService
from server.deps import get_production_service

router = APIRouter(prefix="/batches", tags=["batches"])


class UpdateActualQuantityRequest(BaseModel):
    actual_quantity: float


@router.post("", response_model=Batch, status_code=status.HTTP_201_CREATED)
def create_batch(
    data: BatchCreate,
    service: ProductionService = Depends(get_production_service),
):
    try:
        return service.create_batch(data)
    except ValidationError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("", response_model=List[Batch])
def list_batches(service: ProductionService = Depends(get_production_service)):
    return service.list_batches()


@router.get("/{batch_id}", response_model=Batch)
def get_batch(
    batch_id: int,
    service: ProductionService = Depends(get_production_service),
):
    batch = service.get_batch(batch_id)
    if not batch:
        raise HTTPException(status_code=404, detail="批次不存在")
    return batch


@router.patch("/{batch_id}/actual-quantity", response_model=Batch)
def update_actual_quantity(
    batch_id: int,
    data: UpdateActualQuantityRequest,
    service: ProductionService = Depends(get_production_service),
):
    try:
        batch = service.update_batch_actual_quantity(batch_id, data.actual_quantity)
    except ValidationError as e:
        raise HTTPException(status_code=400, detail=str(e))
    
    if not batch:
        raise HTTPException(status_code=404, detail="批次不存在")
    return batch
