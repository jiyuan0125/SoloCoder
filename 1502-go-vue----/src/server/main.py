import os
from typing import List, Optional
from datetime import datetime, date, time
from fastapi import FastAPI, HTTPException
from fastapi.responses import JSONResponse
from pydantic import BaseModel

from ..core.models import (
    Pond, WaterQualityThreshold, WaterQualityRecord,
    FeedingPlan, FeedingRecord, HarvestRecord, WaterQualityAggregation
)
from ..core.services import (
    PondService, WaterQualityService, FeedingService, HarvestService
)


app = FastAPI(title="水产养殖管理系统", version="1.0.0")


class PondCreate(BaseModel):
    code: str
    area: float
    species: str
    stock_quantity: int


class PondUpdate(BaseModel):
    code: Optional[str] = None
    area: Optional[float] = None
    species: Optional[str] = None
    stock_quantity: Optional[int] = None


class ThresholdCreate(BaseModel):
    pond_id: int
    parameter: str
    min_value: Optional[float] = None
    max_value: Optional[float] = None


class WaterQualityRecordCreate(BaseModel):
    pond_id: int
    water_temp: float
    dissolved_oxygen: float
    ph: float
    ammonia: float


class FeedingPlanCreate(BaseModel):
    pond_id: int
    feed_date: date
    feed_time: time
    amount: float


class HarvestCreate(BaseModel):
    pond_id: int
    species: str
    quantity: int
    weight: float


@app.get("/")
def root():
    return {"message": "水产养殖管理系统 API", "version": "1.0.0"}


# Ponds
@app.post("/ponds", response_model=Pond)
def create_pond(data: PondCreate):
    try:
        return PondService.create_pond(
            code=data.code,
            area=data.area,
            species=data.species,
            stock_quantity=data.stock_quantity
        )
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/ponds", response_model=List[Pond])
def list_ponds():
    return PondService.list_ponds()


@app.get("/ponds/{pond_id}", response_model=Pond)
def get_pond(pond_id: int):
    pond = PondService.get_pond(pond_id)
    if not pond:
        raise HTTPException(status_code=404, detail="池塘不存在")
    return pond


@app.put("/ponds/{pond_id}", response_model=Pond)
def update_pond(pond_id: int, data: PondUpdate):
    update_data = {k: v for k, v in data.dict().items() if v is not None}
    pond = PondService.update_pond(pond_id, **update_data)
    if not pond:
        raise HTTPException(status_code=404, detail="池塘不存在")
    return pond


@app.delete("/ponds/{pond_id}")
def delete_pond(pond_id: int):
    success = PondService.delete_pond(pond_id)
    if not success:
        raise HTTPException(status_code=404, detail="池塘不存在")
    return {"message": "删除成功"}


# Water Quality Thresholds
@app.post("/thresholds", response_model=WaterQualityThreshold)
def set_threshold(data: ThresholdCreate):
    try:
        return WaterQualityService.set_threshold(
            pond_id=data.pond_id,
            parameter=data.parameter,
            min_value=data.min_value,
            max_value=data.max_value
        )
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/thresholds", response_model=List[WaterQualityThreshold])
def list_thresholds(pond_id: Optional[int] = None):
    return WaterQualityService.get_thresholds(pond_id)


# Water Quality Records
@app.post("/water-quality", response_model=WaterQualityRecord)
def add_water_quality(data: WaterQualityRecordCreate):
    try:
        return WaterQualityService.add_record(
            pond_id=data.pond_id,
            water_temp=data.water_temp,
            dissolved_oxygen=data.dissolved_oxygen,
            ph=data.ph,
            ammonia=data.ammonia
        )
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/water-quality", response_model=List[WaterQualityRecord])
def list_water_quality(
    pond_id: Optional[int] = None,
    start_time: Optional[datetime] = None,
    end_time: Optional[datetime] = None
):
    return WaterQualityService.get_records(pond_id, start_time, end_time)


@app.get("/water-quality/aggregate", response_model=List[WaterQualityAggregation])
def aggregate_water_quality(
    pond_id: Optional[int] = None,
    start_time: Optional[datetime] = None,
    end_time: Optional[datetime] = None
):
    return WaterQualityService.aggregate_data(pond_id, start_time, end_time)


# Feeding Plans
@app.post("/feeding-plans", response_model=FeedingPlan)
def create_feeding_plan(data: FeedingPlanCreate):
    try:
        return FeedingService.create_plan(
            pond_id=data.pond_id,
            feed_date=data.feed_date,
            feed_time=data.feed_time,
            amount=data.amount
        )
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/feeding-plans", response_model=List[FeedingPlan])
def list_feeding_plans(pond_id: Optional[int] = None):
    return FeedingService.get_plans(pond_id)


@app.delete("/feeding-plans/{plan_id}")
def delete_feeding_plan(plan_id: int):
    success = FeedingService.delete_plan(plan_id)
    if not success:
        raise HTTPException(status_code=404, detail="投喂计划不存在")
    return {"message": "删除成功"}


# Feeding Records
@app.post("/feeding-execute", response_model=List[FeedingRecord])
def execute_feeding():
    return FeedingService.execute_scheduled_plans()


@app.get("/feeding-records", response_model=List[FeedingRecord])
def list_feeding_records(pond_id: Optional[int] = None):
    return FeedingService.get_records(pond_id)


# Harvest
@app.post("/harvests", response_model=HarvestRecord)
def create_harvest(data: HarvestCreate):
    try:
        return HarvestService.create_harvest(
            pond_id=data.pond_id,
            species=data.species,
            quantity=data.quantity,
            weight=data.weight
        )
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/harvests", response_model=List[HarvestRecord])
def list_harvests(pond_id: Optional[int] = None):
    return HarvestService.get_records(pond_id)
