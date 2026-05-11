from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from typing import List, Optional
from datetime import date

from server.database import get_db
from server import crud, schemas
from server.config import settings

router = APIRouter(prefix="/api/dispatch", tags=["闸门调度"])


@router.post("/generate", response_model=schemas.DailyDispatchSummary, summary="生成当日调度方案")
def generate_dispatch(target_date: Optional[date] = None, db: Session = Depends(get_db)):
    if target_date is None:
        target_date = date.today()
    
    schemes = crud.generate_daily_dispatch(db, target_date=target_date)
    is_irrigation = crud.is_irrigation_season(target_date)
    
    return schemas.DailyDispatchSummary(
        date=target_date,
        is_irrigation_season=is_irrigation,
        total_schemes=len(schemes),
        ecological_flow=settings.ECOLOGICAL_FLOW
    )


@router.get("/schemes", response_model=List[schemas.DispatchSchemeResponse], summary="获取指定日期调度方案")
def get_dispatch_schemes(target_date: date, db: Session = Depends(get_db)):
    schemes = crud.get_dispatch_schemes(db, target_date=target_date)
    return schemes


@router.get("/today", response_model=List[schemas.DispatchSchemeResponse], summary="获取今日调度方案")
def get_today_dispatch(db: Session = Depends(get_db)):
    today = date.today()
    schemes = crud.get_dispatch_schemes(db, target_date=today)
    if not schemes:
        schemes = crud.generate_daily_dispatch(db, target_date=today)
    return schemes


@router.get("/quota-remaining/{canal_id}", summary="获取渠道配额余量")
def get_quota_remaining(
    canal_id: int,
    year: Optional[int] = None,
    month: Optional[int] = Query(None, ge=1, le=12),
    db: Session = Depends(get_db)
):
    today = date.today()
    if year is None:
        year = today.year
    if month is None:
        month = today.month
    
    remaining = crud.get_quota_remaining(db, canal_id=canal_id, year=year, month=month)
    canal = crud.get_canal(db, canal_id=canal_id)
    
    if canal is None:
        raise HTTPException(status_code=404, detail="渠道不存在")
    
    return {
        "canal_id": canal_id,
        "canal_name": canal.name,
        "year": year,
        "month": month,
        "remaining_amount": remaining,
        "has_plan": remaining > 0 or crud.get_water_plan_by_canal_and_year(db, canal_id, year) is not None
    }
