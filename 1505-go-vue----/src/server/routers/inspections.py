from typing import List

from fastapi import APIRouter, Depends, HTTPException, status

from core.models import QualityInspection, QualityInspectionCreate, ValidationError
from core.services import ProductionService
from server.deps import get_production_service

router = APIRouter(prefix="/inspections", tags=["inspections"])


@router.post("", response_model=QualityInspection, status_code=status.HTTP_201_CREATED)
def create_inspection(
    data: QualityInspectionCreate,
    service: ProductionService = Depends(get_production_service),
):
    try:
        return service.create_quality_inspection(data)
    except ValidationError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/batch/{batch_id}", response_model=List[QualityInspection])
def list_inspections_by_batch(
    batch_id: int,
    service: ProductionService = Depends(get_production_service),
):
    return service.get_inspections_by_batch(batch_id)
