import os
from datetime import datetime, date
from typing import Optional, List
from uuid import UUID

from fastapi import FastAPI, HTTPException
from fastapi.responses import JSONResponse
from pydantic import BaseModel

from ..core.models import (
    Storehouse, WarehouseArea, StockInRecord, StockOutRecord,
    TemperatureRecord, PestInspectionRecord, TodoItem,
    AreaType, GrainVariety
)
from ..core.services import (
    storehouse_service, area_service, stock_service,
    temperature_service, pest_service, todo_service, metrics_service,
    BusinessError
)

app = FastAPI(title="粮油仓储库存管理系统", version="1.0.0")


def _handle_business_error(e: BusinessError):
    raise HTTPException(status_code=400, detail=str(e))


class StorehouseCreate(BaseModel):
    name: str
    location: str


class WarehouseAreaCreate(BaseModel):
    storehouse_id: UUID
    name: str
    area_type: AreaType
    capacity_kg: float
    min_temp: float = 10.0
    max_temp: float = 25.0


class StockInCreate(BaseModel):
    area_id: UUID
    variety: GrainVariety
    quantity_kg: float
    source: str
    in_date: Optional[date] = None


class StockOutCreate(BaseModel):
    area_id: UUID
    variety: GrainVariety
    quantity_kg: float
    destination: str
    purpose: str
    out_date: Optional[date] = None


class TemperatureReport(BaseModel):
    area_id: UUID
    temperature: float
    record_time: Optional[datetime] = None


class PestInspectionCreate(BaseModel):
    area_id: UUID
    pest_count_per_kg: float
    variety: Optional[GrainVariety] = None
    inspection_date: Optional[date] = None
    notes: Optional[str] = None


@app.get("/")
def root():
    return {"name": "粮油仓储库存管理系统", "version": "1.0.0"}


@app.get("/storehouses", response_model=List[Storehouse])
def list_storehouses():
    return storehouse_service.list()


@app.post("/storehouses", response_model=Storehouse)
def create_storehouse(data: StorehouseCreate):
    return storehouse_service.create(name=data.name, location=data.location)


@app.get("/storehouses/{storehouse_id}", response_model=Storehouse)
def get_storehouse(storehouse_id: UUID):
    storehouse = storehouse_service.get(storehouse_id)
    if not storehouse:
        raise HTTPException(status_code=404, detail="仓库不存在")
    return storehouse


@app.get("/areas", response_model=List[WarehouseArea])
def list_areas(storehouse_id: Optional[UUID] = None):
    return area_service.list(storehouse_id=storehouse_id)


@app.post("/areas", response_model=WarehouseArea)
def create_area(data: WarehouseAreaCreate):
    try:
        return area_service.create(
            storehouse_id=data.storehouse_id,
            name=data.name,
            area_type=data.area_type,
            capacity_kg=data.capacity_kg,
            min_temp=data.min_temp,
            max_temp=data.max_temp,
        )
    except BusinessError as e:
        _handle_business_error(e)


@app.get("/areas/{area_id}", response_model=WarehouseArea)
def get_area(area_id: UUID):
    area = area_service.get(area_id)
    if not area:
        raise HTTPException(status_code=404, detail="库区不存在")
    return area


@app.get("/stock-in", response_model=List[StockInRecord])
def list_stock_ins(area_id: Optional[UUID] = None):
    return stock_service.list_stock_ins(area_id=area_id)


@app.post("/stock-in", response_model=StockInRecord)
def create_stock_in(data: StockInCreate):
    try:
        return stock_service.stock_in(
            area_id=data.area_id,
            variety=data.variety.value,
            quantity_kg=data.quantity_kg,
            source=data.source,
            in_date=data.in_date,
        )
    except BusinessError as e:
        _handle_business_error(e)


@app.get("/stock-in/fifo-recommendation")
def get_fifo_recommendation(area_id: UUID, variety: GrainVariety, quantity_kg: float):
    return stock_service.get_fifo_recommendation(area_id, variety.value, quantity_kg)


@app.get("/stock-in/expiring-soon", response_model=List[StockInRecord])
def get_expiring_soon(days: int = 30):
    return stock_service.get_expiring_soon(days=days)


@app.get("/stock-out", response_model=List[StockOutRecord])
def list_stock_outs(area_id: Optional[UUID] = None):
    return stock_service.list_stock_outs(area_id=area_id)


@app.post("/stock-out", response_model=List[StockOutRecord])
def create_stock_out(data: StockOutCreate):
    try:
        return stock_service.stock_out(
            area_id=data.area_id,
            variety=data.variety.value,
            quantity_kg=data.quantity_kg,
            destination=data.destination,
            purpose=data.purpose,
            out_date=data.out_date,
        )
    except BusinessError as e:
        _handle_business_error(e)


@app.get("/temperatures", response_model=List[TemperatureRecord])
def list_temperatures(area_id: Optional[UUID] = None):
    return temperature_service.list_records(area_id=area_id)


@app.post("/temperatures", response_model=TemperatureRecord)
def report_temperature(data: TemperatureReport):
    try:
        return temperature_service.report_temperature(
            area_id=data.area_id,
            temperature=data.temperature,
            record_time=data.record_time,
        )
    except BusinessError as e:
        _handle_business_error(e)


@app.get("/temperatures/alarms", response_model=List[TemperatureRecord])
def get_temperature_alarms():
    return temperature_service.get_alarming_records()


@app.get("/pest-inspections", response_model=List[PestInspectionRecord])
def list_pest_inspections(area_id: Optional[UUID] = None):
    return pest_service.list_inspections(area_id=area_id)


@app.post("/pest-inspections", response_model=PestInspectionRecord)
def create_pest_inspection(data: PestInspectionCreate):
    try:
        return pest_service.record_inspection(
            area_id=data.area_id,
            pest_count_per_kg=data.pest_count_per_kg,
            variety=data.variety.value if data.variety else None,
            inspection_date=data.inspection_date,
            notes=data.notes,
        )
    except BusinessError as e:
        _handle_business_error(e)


@app.get("/todos", response_model=List[TodoItem])
def list_todos(area_id: Optional[UUID] = None, pending_only: bool = False):
    return todo_service.list_todos(area_id=area_id, pending_only=pending_only)


@app.post("/todos/{todo_id}/complete", response_model=TodoItem)
def complete_todo(todo_id: UUID):
    todo = todo_service.complete_todo(todo_id)
    if not todo:
        raise HTTPException(status_code=404, detail="待办不存在")
    return todo


@app.get("/metrics")
def get_metrics():
    return metrics_service.get_all_metrics()
