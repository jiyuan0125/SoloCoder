from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List, Optional

from server.database import get_db
from server import crud, schemas

router = APIRouter(prefix="/api/water-plans", tags=["用水计划管理"])


def format_plan_response(plan):
    response_data = {
        "id": plan.id,
        "canal_id": plan.canal_id,
        "year": plan.year,
        "initial_annual_quota": plan.initial_annual_quota,
        "current_annual_quota": plan.current_annual_quota,
        "annual_used": plan.annual_used,
        "status": plan.status,
        "created_at": plan.created_at,
        "updated_at": plan.updated_at,
        "quarterly_quotas": []
    }
    for qq in plan.quarterly_quotas:
        response_data["quarterly_quotas"].append({
            "id": qq.id,
            "quarter": qq.quarter,
            "initial_quota": qq.initial_quota,
            "current_quota": qq.current_quota,
            "used_amount": qq.used_amount,
            "remaining_amount": max(0.0, qq.current_quota - qq.used_amount),
            "is_critical": qq.is_critical,
            "updated_at": qq.updated_at
        })
    return schemas.WaterPlanResponse(**response_data)


@router.post("/", response_model=schemas.WaterPlanResponse, summary="创建年度用水计划")
def create_water_plan(plan: schemas.WaterPlanCreate, db: Session = Depends(get_db)):
    existing = crud.get_water_plan_by_canal_and_year(
        db, canal_id=plan.canal_id, year=plan.year
    )
    if existing:
        raise HTTPException(
            status_code=400, 
            detail="该渠道本年度用水计划已存在"
        )
    
    try:
        db_plan = crud.create_water_plan(db=db, plan=plan)
        return format_plan_response(db_plan)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/", response_model=List[schemas.WaterPlanResponse], summary="获取用水计划列表")
def read_water_plans(
    skip: int = 0,
    limit: int = 100,
    year: Optional[int] = None,
    canal_id: Optional[int] = None,
    db: Session = Depends(get_db)
):
    plans = crud.get_water_plans(db, skip=skip, limit=limit, year=year, canal_id=canal_id)
    return [format_plan_response(p) for p in plans]


@router.get("/{plan_id}", response_model=schemas.WaterPlanResponse, summary="获取单个用水计划")
def read_water_plan(plan_id: int, db: Session = Depends(get_db)):
    db_plan = crud.get_water_plan(db, plan_id=plan_id)
    if db_plan is None:
        raise HTTPException(status_code=404, detail="用水计划不存在")
    return format_plan_response(db_plan)


@router.post("/{plan_id}/adjust-quota", response_model=schemas.QuarterlyQuotaResponse, summary="调整季度配额")
def adjust_quarterly_quota(
    plan_id: int,
    adjustment: schemas.QuotaAdjustment,
    db: Session = Depends(get_db)
):
    try:
        qq = crud.adjust_quarterly_quota(
            db,
            plan_id=plan_id,
            quarter=adjustment.quarter,
            new_quota=adjustment.new_quota
        )
        return schemas.QuarterlyQuotaResponse(
            id=qq.id,
            quarter=qq.quarter,
            initial_quota=qq.initial_quota,
            current_quota=qq.current_quota,
            used_amount=qq.used_amount,
            remaining_amount=max(0.0, qq.current_quota - qq.used_amount),
            is_critical=qq.is_critical,
            updated_at=qq.updated_at
        )
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
