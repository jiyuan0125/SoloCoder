from typing import List, Optional
from datetime import date
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session

from ..dependencies import get_db
from ...core.schemas import WastageCreate, WastageResponse
from ...core.service import WastageService, BusinessException


router = APIRouter(prefix="/wastages", tags=["wastages"])


@router.post("", response_model=WastageResponse)
def create_wastage(wastage_create: WastageCreate, db: Session = Depends(get_db)):
    try:
        wastage = WastageService(db).create(wastage_create)
        return WastageService(db)._to_response(wastage)
    except BusinessException as e:
        raise HTTPException(status_code=400, detail=e.detail)


@router.get("", response_model=List[WastageResponse])
def list_wastages(
    store_id: Optional[int] = Query(None),
    start_date: Optional[date] = Query(None),
    end_date: Optional[date] = Query(None),
    db: Session = Depends(get_db),
):
    service = WastageService(db)
    if store_id:
        return service.list_by_store(store_id, start_date, end_date)
    return service.list_all(start_date, end_date)


@router.get("/{wastage_id}", response_model=WastageResponse)
def get_wastage(wastage_id: int, db: Session = Depends(get_db)):
    wastage = WastageService(db).get_by_id(wastage_id)
    if not wastage:
        raise HTTPException(status_code=404, detail=f"损耗记录 ID {wastage_id} 不存在")
    return WastageService(db)._to_response(wastage)
