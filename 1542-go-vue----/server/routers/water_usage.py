from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from typing import List, Optional
from datetime import date

from server.database import get_db
from server import crud, schemas
from server.config import settings

router = APIRouter(prefix="/api/water-usage", tags=["用水量管理"])


@router.post("/record", response_model=schemas.WaterUsageResponse, summary="记录月度用水量")
def record_water_usage(usage: schemas.WaterUsageCreate, db: Session = Depends(get_db)):
    return crud.record_water_usage(db=db, usage=usage)


@router.get("/", response_model=List[schemas.WaterUsageResponse], summary="获取用水量列表")
def read_water_usages(
    skip: int = 0,
    limit: int = 100,
    canal_id: Optional[int] = None,
    year: Optional[int] = None,
    month: Optional[int] = Query(None, ge=1, le=12),
    db: Session = Depends(get_db)
):
    return crud.get_water_usages(
        db, skip=skip, limit=limit,
        canal_id=canal_id, year=year, month=month
    )


@router.get("/{usage_id}", response_model=schemas.WaterUsageResponse, summary="获取单个用水量记录")
def read_water_usage(usage_id: int, db: Session = Depends(get_db)):
    usage = crud.get_water_usage(db, usage_id=usage_id)
    if usage is None:
        raise HTTPException(status_code=404, detail="用水量记录不存在")
    return usage


@router.get("/statistics/monthly", response_model=List[schemas.MonthlyStatistics], summary="月度统计")
def get_monthly_statistics(
    year: int,
    month: int = Query(..., ge=1, le=12),
    db: Session = Depends(get_db)
):
    canals = crud.get_canals(db, limit=1000)
    results = []
    
    for canal in canals:
        usage = crud.get_water_usage_by_canal_year_month(db, canal.id, year, month)
        quota = crud.get_monthly_quota(db, canal.id, year, month)
        
        usage_amount = usage.usage_amount if usage else 0.0
        usage_rate = usage_amount / quota if quota > 0 else 0.0
        
        stats = schemas.MonthlyStatistics(
            canal_id=canal.id,
            canal_name=canal.name,
            year=year,
            month=month,
            total_usage=usage_amount,
            quota=quota,
            usage_rate=usage_rate,
            is_warning=usage_rate >= settings.WARNING_THRESHOLD if quota > 0 else False,
            is_over_quota=usage.is_over_quota if usage else False
        )
        results.append(stats)
    
    return results


@router.post("/irrigation-end/{canal_id}/{year}", summary="灌溉结束后更新年度累计执行量")
def update_annual_execution(canal_id: int, year: int, db: Session = Depends(get_db)):
    plan = crud.update_annual_execution_after_irrigation(db, canal_id=canal_id, year=year)
    if plan is None:
        raise HTTPException(status_code=404, detail="该渠道年度用水计划不存在")
    return {
        "message": "更新成功",
        "canal_id": canal_id,
        "year": year,
        "annual_used": plan.annual_used
    }
