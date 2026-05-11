from typing import List, Optional, Dict
from fastapi import APIRouter, Depends, HTTPException, Query
from fastapi.responses import PlainTextResponse
from src.core.models import Sample
from src.core.services import SampleService, ExportService
from src.core.exceptions import (
    NotFoundError,
    BusinessRuleError
)
from src.server.dependencies import get_sample_service, get_export_service


router = APIRouter(prefix="/samples", tags=["samples"])


@router.post("/", response_model=Sample)
def create_sample(
    sample: Sample,
    service: SampleService = Depends(get_sample_service)
):
    try:
        return service.create_sample(sample)
    except NotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except BusinessRuleError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/", response_model=List[Sample])
def list_samples(
    borehole_id: Optional[str] = Query(None, description="按钻孔ID筛选"),
    service: SampleService = Depends(get_sample_service)
):
    return service.list_samples(borehole_id)


@router.get("/{sample_id}", response_model=Sample)
def get_sample(
    sample_id: str,
    service: SampleService = Depends(get_sample_service)
):
    try:
        return service.get_sample(sample_id)
    except NotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))


@router.put("/{sample_id}", response_model=Sample)
def update_sample(
    sample_id: str,
    sample: Sample,
    service: SampleService = Depends(get_sample_service)
):
    try:
        return service.update_sample(sample_id, sample)
    except NotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except BusinessRuleError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.delete("/{sample_id}")
def delete_sample(
    sample_id: str,
    service: SampleService = Depends(get_sample_service)
):
    try:
        service.delete_sample(sample_id)
        return {"message": "样品已删除"}
    except NotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))


@router.patch("/{sample_id}/analysis", response_model=Sample)
def update_analysis_results(
    sample_id: str,
    results: Dict[str, float],
    service: SampleService = Depends(get_sample_service)
):
    try:
        return service.update_analysis_results(sample_id, results)
    except NotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except BusinessRuleError as e:
        raise HTTPException(status_code=400, detail=str(e))
