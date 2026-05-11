from typing import List
from fastapi import APIRouter, HTTPException, status

from src.core.models.schemas import Recipe, RecipeCreate
from src.core.storage.memory_store import get_store


router = APIRouter(prefix="/recipes", tags=["recipes"])
store = get_store()


@router.post("/", response_model=Recipe, status_code=status.HTTP_201_CREATED)
def create_recipe(data: RecipeCreate):
    return store.create_recipe(data)


@router.get("/", response_model=List[Recipe])
def list_recipes():
    return store.list_recipes()


@router.get("/{recipe_id}", response_model=Recipe)
def get_recipe(recipe_id: str):
    recipe = store.get_recipe(recipe_id)
    if not recipe:
        raise HTTPException(status_code=404, detail="Recipe not found")
    return recipe
