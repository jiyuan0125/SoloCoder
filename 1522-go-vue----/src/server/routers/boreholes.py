from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from src.core.models import Borehole
from src.core.services import BoreholeService
from src.core.exceptions import (
    NotFoundError,
    BusinessRuleError,
    DuplicateError
)
from src.server.dependencies import get_borehole_service


router = APIRouter(prefix="/boreholes", tags=["boreholes"])


@router.post("/", response_model=Borehole)
def create_borehole(
    borehole: Borehole,
    service: BoreholeService = Depends(get_borehole_service)
):
    try:
        return service.create_borehole(borehole)
    except NotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except (BusinessRuleError, DuplicateError) as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/", response_model=List[Borehole])
def list_boreholes(
    project_id: Optional[str] = Query(None, description="按项目ID筛选"),
    service: BoreholeService = Depends(get_borehole_service)
):
    return service.list_boreholes(project_id)


@router.get("/{borehole_id}", response_model=Borehole)
def get_borehole(
    borehole_id: str,
    service: BoreholeService = Depends(get_borehole_service)
):
    try:
        return service.get_borehole(borehole_id)
    except NotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))


@router.put("/{borehole_id}", response_model=Borehole)
def update_borehole(
    borehole_id: str,
    borehole: Borehole,
    service: BoreholeService = Depends(get_borehole_service)
):
    try:
        return service.update_borehole(borehole_id, borehole)
    except NotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except BusinessRuleError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.delete("/{borehole_id}")
def delete_borehole(
    borehole_id: str,
    service: BoreholeService = Depends(get_borehole_service)
):
    try:
        service.delete_borehole(borehole_id)
        return {"message": "钻孔已删除"}
    except NotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))
