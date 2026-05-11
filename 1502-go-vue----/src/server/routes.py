from datetime import datetime
from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session

from core import models, schemas, services
from core.database import get_db

router = APIRouter()


@router.post("/ponds", response_model=schemas.PondResponse, status_code=201)
def create_pond(pond: schemas.PondCreate, db: Session = Depends(get_db)):
    existing = services.get_pond_by_code(db, pond.code)
    if existing:
        raise HTTPException(status_code=400, detail="池塘编号已存在")
    return services.create_pond(db, pond)


@router.get("/ponds", response_model=List[schemas.PondResponse])
def list_ponds(db: Session = Depends(get_db)):
    return services.get_all_ponds(db)


@router.get("/ponds/{pond_id}", response_model=schemas.PondResponse)
def get_pond(pond_id: int, db: Session = Depends(get_db)):
    pond = services.get_pond(db, pond_id)
    if not pond:
        raise HTTPException(status_code=404, detail="池塘不存在")
    return pond


@router.patch("/ponds/{pond_id}", response_model=schemas.PondResponse)
def update_pond(
    pond_id: int, pond_update: schemas.PondUpdate, db: Session = Depends(get_db)
):
    pond = services.update_pond(db, pond_id, pond_update)
    if not pond:
        raise HTTPException(status_code=404, detail="池塘不存在")
    return pond


@router.delete("/ponds/{pond_id}", status_code=204)
def delete_pond(pond_id: int, db: Session = Depends(get_db)):
    success = services.delete_pond(db, pond_id)
    if not success:
        raise HTTPException(status_code=404, detail="池塘不存在")


@router.post(
    "/ponds/{pond_id}/thresholds",
    response_model=schemas.ThresholdResponse,
    status_code=201,
)
def set_threshold(
    pond_id: int, threshold: schemas.ThresholdCreate, db: Session = Depends(get_db)
):
    pond = services.get_pond(db, pond_id)
    if not pond:
        raise HTTPException(status_code=404, detail="池塘不存在")
    valid_indicators = ["temperature", "dissolved_oxygen", "ph", "ammonia_nitrogen"]
    if threshold.indicator not in valid_indicators:
        raise HTTPException(status_code=400, detail=f"无效指标，有效值: {valid_indicators}")
    return services.set_threshold(db, pond_id, threshold)


@router.get("/ponds/{pond_id}/thresholds", response_model=List[schemas.ThresholdResponse])
def list_thresholds(pond_id: int, db: Session = Depends(get_db)):
    pond = services.get_pond(db, pond_id)
    if not pond:
        raise HTTPException(status_code=404, detail="池塘不存在")
    return services.get_thresholds(db, pond_id)


@router.post(
    "/ponds/{pond_id}/water-records",
    response_model=schemas.WaterRecordResponse,
    status_code=201,
)
def create_water_record(
    pond_id: int, record: schemas.WaterRecordCreate, db: Session = Depends(get_db)
):
    pond = services.get_pond(db, pond_id)
    if not pond:
        raise HTTPException(status_code=404, detail="池塘不存在")
    return services.create_water_record(db, pond_id, record)


@router.get("/ponds/{pond_id}/water-records", response_model=List[schemas.WaterRecordResponse])
def list_water_records(
    pond_id: int,
    limit: int = Query(100, gt=0, le=1000),
    db: Session = Depends(get_db),
):
    pond = services.get_pond(db, pond_id)
    if not pond:
        raise HTTPException(status_code=404, detail="池塘不存在")
    return services.get_water_records(db, pond_id, limit=limit)


@router.get("/ponds/{pond_id}/water-stats", response_model=schemas.WaterStatsResponse)
def get_water_stats(
    pond_id: int,
    start_time: datetime = Query(...),
    end_time: datetime = Query(...),
    db: Session = Depends(get_db),
):
    pond = services.get_pond(db, pond_id)
    if not pond:
        raise HTTPException(status_code=404, detail="池塘不存在")
    if start_time >= end_time:
        raise HTTPException(status_code=400, detail="开始时间必须早于结束时间")
    stats = services.get_water_stats(db, pond_id, start_time, end_time)
    return schemas.WaterStatsResponse(
        pond_id=pond_id,
        start_time=start_time,
        end_time=end_time,
        stats=stats,
    )


@router.post(
    "/ponds/{pond_id}/feeding-plans",
    response_model=schemas.FeedingPlanResponse,
    status_code=201,
)
def create_feeding_plan(
    pond_id: int, plan: schemas.FeedingPlanCreate, db: Session = Depends(get_db)
):
    pond = services.get_pond(db, pond_id)
    if not pond:
        raise HTTPException(status_code=404, detail="池塘不存在")
    return services.create_feeding_plan(db, pond_id, plan)


@router.get("/ponds/{pond_id}/feeding-plans", response_model=List[schemas.FeedingPlanResponse])
def list_feeding_plans(
    pond_id: int,
    limit: int = Query(100, gt=0, le=1000),
    db: Session = Depends(get_db),
):
    pond = services.get_pond(db, pond_id)
    if not pond:
        raise HTTPException(status_code=404, detail="池塘不存在")
    return services.get_feeding_plans(db, pond_id, limit=limit)


@router.post(
    "/feeding-plans/{plan_id}/execute",
    response_model=schemas.FeedingRecordResponse,
    status_code=201,
)
def execute_feeding_plan(plan_id: int, db: Session = Depends(get_db)):
    record = services.execute_feeding_plan(db, plan_id)
    if not record:
        raise HTTPException(status_code=400, detail="投喂计划无法执行（已执行或已过期）")
    return record


@router.get("/ponds/{pond_id}/feeding-records", response_model=List[schemas.FeedingRecordResponse])
def list_feeding_records(
    pond_id: int,
    limit: int = Query(100, gt=0, le=1000),
    db: Session = Depends(get_db),
):
    pond = services.get_pond(db, pond_id)
    if not pond:
        raise HTTPException(status_code=404, detail="池塘不存在")
    return services.get_feeding_records(db, pond_id, limit=limit)


@router.post(
    "/ponds/{pond_id}/harvests",
    response_model=schemas.HarvestResponse,
    status_code=201,
)
def create_harvest(
    pond_id: int, harvest: schemas.HarvestCreate, db: Session = Depends(get_db)
):
    pond = services.get_pond(db, pond_id)
    if not pond:
        raise HTTPException(status_code=404, detail="池塘不存在")
    result = services.create_harvest(db, pond_id, harvest)
    if not result:
        raise HTTPException(status_code=400, detail="存塘量不足，无法创建捕捞记录")
    return result


@router.get("/ponds/{pond_id}/harvests", response_model=List[schemas.HarvestResponse])
def list_harvests(
    pond_id: int,
    limit: int = Query(100, gt=0, le=1000),
    db: Session = Depends(get_db),
):
    pond = services.get_pond(db, pond_id)
    if not pond:
        raise HTTPException(status_code=404, detail="池塘不存在")
    return services.get_harvest_records(db, pond_id, limit=limit)
