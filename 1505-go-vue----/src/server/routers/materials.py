from typing import List

from fastapi import APIRouter, Depends, HTTPException, status
from pydantic import BaseModel

from core.models import MaterialInventory, ValidationError
from core.services import ProductionService
from server.deps import get_production_service

router = APIRouter(prefix="/materials", tags=["materials"])


class CreateMaterialRequest(BaseModel):
    name: str
    current_stock: float
    safety_stock: float


class UpdateStockRequest(BaseModel):
    current_stock: float


@router.post("", response_model=MaterialInventory, status_code=status.HTTP_201_CREATED)
def create_material(
    data: CreateMaterialRequest,
    service: ProductionService = Depends(get_production_service),
):
    try:
        return service.create_material_inventory(
            name=data.name,
            current_stock=data.current_stock,
            safety_stock=data.safety_stock,
        )
    except ValidationError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("", response_model=List[MaterialInventory])
def list_materials(service: ProductionService = Depends(get_production_service)):
    return service.list_materials()


@router.get("/{material_id}", response_model=MaterialInventory)
def get_material(
    material_id: int,
    service: ProductionService = Depends(get_production_service),
):
    material = service.get_material(material_id)
    if not material:
        raise HTTPException(status_code=404, detail="原材料不存在")
    return material


@router.patch("/{material_id}/stock", response_model=MaterialInventory)
def update_stock(
    material_id: int,
    data: UpdateStockRequest,
    service: ProductionService = Depends(get_production_service),
):
    try:
        material = service.update_material_stock(material_id, data.current_stock)
    except ValidationError as e:
        raise HTTPException(status_code=400, detail=str(e))
    
    if not material:
        raise HTTPException(status_code=404, detail="原材料不存在")
    return material


@router.get("/alert/low-stock", response_model=List[MaterialInventory])
def list_low_stock_materials(service: ProductionService = Depends(get_production_service)):
    return service.get_low_stock_materials()
