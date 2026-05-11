from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session

from ..dependencies import get_db
from ...core.schemas import IngredientCreate, IngredientUpdate, IngredientResponse
from ...core.service import IngredientService, BusinessException


router = APIRouter(prefix="/ingredients", tags=["ingredients"])


@router.post("", response_model=IngredientResponse)
def create_ingredient(
    ingredient_create: IngredientCreate, db: Session = Depends(get_db)
):
    try:
        ingredient = IngredientService(db).create(ingredient_create)
        return IngredientService(db)._to_response(ingredient)
    except BusinessException as e:
        raise HTTPException(status_code=400, detail=e.detail)


@router.get("", response_model=List[IngredientResponse])
def list_ingredients(
    store_id: Optional[int] = Query(None), db: Session = Depends(get_db)
):
    service = IngredientService(db)
    if store_id:
        return service.list_by_store(store_id)
    return service.list_all()


@router.get("/{ingredient_id}", response_model=IngredientResponse)
def get_ingredient(ingredient_id: int, db: Session = Depends(get_db)):
    ingredient = IngredientService(db).get_by_id(ingredient_id)
    if not ingredient:
        raise HTTPException(status_code=404, detail=f"食材 ID {ingredient_id} 不存在")
    return IngredientService(db)._to_response(ingredient)


@router.patch("/{ingredient_id}", response_model=IngredientResponse)
def update_ingredient(
    ingredient_id: int,
    update_data: IngredientUpdate,
    db: Session = Depends(get_db),
):
    try:
        updated = IngredientService(db).update(ingredient_id, update_data)
        return IngredientService(db)._to_response(updated)
    except BusinessException as e:
        raise HTTPException(status_code=400, detail=e.detail)
