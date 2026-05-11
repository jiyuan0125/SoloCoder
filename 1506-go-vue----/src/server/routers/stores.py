from typing import List
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from ..dependencies import get_db
from ...core.schemas import StoreCreate, StoreResponse
from ...core.service import StoreService, BusinessException


router = APIRouter(prefix="/stores", tags=["stores"])


@router.post("", response_model=StoreResponse)
def create_store(store_create: StoreCreate, db: Session = Depends(get_db)):
    try:
        return StoreService(db).create(store_create)
    except BusinessException as e:
        raise HTTPException(status_code=400, detail=e.detail)


@router.get("", response_model=List[StoreResponse])
def list_stores(db: Session = Depends(get_db)):
    return StoreService(db).list_all()


@router.get("/{store_id}", response_model=StoreResponse)
def get_store(store_id: int, db: Session = Depends(get_db)):
    store = StoreService(db).get_by_id(store_id)
    if not store:
        raise HTTPException(status_code=404, detail=f"门店 ID {store_id} 不存在")
    return store
