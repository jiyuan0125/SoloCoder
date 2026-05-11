from typing import List

from fastapi import APIRouter, Depends, HTTPException, status

from core.models import Recipe, RecipeCreate, ValidationError
from core.services import ProductionService
from server.deps import get_production_service

router = APIRouter(prefix="/recipes", tags=["recipes"])


@router.post("", response_model=Recipe, status_code=status.HTTP_201_CREATED)
def create_recipe(
    data: RecipeCreate,
    service: ProductionService = Depends(get_production_service),
):
    try:
        return service.create_recipe(data)
    except ValidationError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("", response_model=List[Recipe])
def list_recipes(service: ProductionService = Depends(get_production_service)):
    return service.list_recipes()


@router.get("/{recipe_id}", response_model=Recipe)
def get_recipe(
    recipe_id: int,
    service: ProductionService = Depends(get_production_service),
):
    recipe = service.get_recipe(recipe_id)
    if not recipe:
        raise HTTPException(status_code=404, detail="配方不存在")
    return recipe
