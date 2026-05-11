import os
import uuid
from datetime import date
from typing import List, Optional

from fastapi import FastAPI, HTTPException
from fastapi.responses import JSONResponse

from core.aggregator import MetricsAggregator
from core.business_rules import (
    BusinessRuleError,
    update_equipment_running_hours,
    validate_crushing_record,
    validate_equipment_status_transition,
    validate_flotation_record,
)
from core.calculator import GradeCalculator
from core.models import (
    CrushingRecord,
    CrushingRecordCreate,
    Equipment,
    EquipmentCreate,
    EquipmentUpdateStatus,
    FlotationRecord,
    FlotationRecordCreate,
    GradeCalculationInput,
    GradeCalculationResult,
    DailyMetrics,
    EquipmentUtilization,
)
from .storage import storage

app = FastAPI(
    title="选矿厂工艺管理系统",
    description="破碎、浮选工段数据管理系统",
    version="1.0.0",
)


@app.get("/")
def root():
    return {"message": "选矿厂工艺管理系统 API", "version": "1.0.0"}


@app.post("/equipment/", response_model=Equipment, status_code=201)
def create_equipment(data: EquipmentCreate):
    existing = storage.get_equipment(data.id)
    if existing:
        raise HTTPException(
            status_code=409,
            detail=f"设备 {data.id} 已存在",
        )
    equipment = Equipment(**data.dict())
    storage.add_equipment(equipment)
    return equipment


@app.get("/equipment/", response_model=List[Equipment])
def list_equipment():
    return storage.get_all_equipment()


@app.get("/equipment/{equipment_id}", response_model=Equipment)
def get_equipment(equipment_id: str):
    equipment = storage.get_equipment(equipment_id)
    if not equipment:
        raise HTTPException(
            status_code=404,
            detail=f"设备 {equipment_id} 不存在",
        )
    return equipment


@app.patch("/equipment/{equipment_id}/status", response_model=Equipment)
def update_equipment_status(
    equipment_id: str,
    data: EquipmentUpdateStatus,
):
    equipment = storage.get_equipment(equipment_id)
    if not equipment:
        raise HTTPException(
            status_code=404,
            detail=f"设备 {equipment_id} 不存在",
        )
    try:
        validate_equipment_status_transition(equipment, data)
    except BusinessRuleError as e:
        raise HTTPException(status_code=400, detail=str(e))

    updated = update_equipment_running_hours(equipment, data)
    storage.update_equipment(updated)
    return updated


@app.post("/crushing/", response_model=CrushingRecord, status_code=201)
def create_crushing_record(data: CrushingRecordCreate):
    existing_records = storage.get_all_crushing_records()
    try:
        validate_crushing_record(data, existing_records)
    except BusinessRuleError as e:
        raise HTTPException(status_code=400, detail=str(e))

    record_id = str(uuid.uuid4())
    record = CrushingRecord(id=record_id, **data.dict())
    storage.add_crushing_record(record)
    return record


@app.get("/crushing/", response_model=List[CrushingRecord])
def list_crushing_records(
    start_date: Optional[date] = None,
    end_date: Optional[date] = None,
    crusher_id: Optional[str] = None,
):
    records = storage.get_all_crushing_records()
    if start_date:
        records = [r for r in records if r.record_date >= start_date]
    if end_date:
        records = [r for r in records if r.record_date <= end_date]
    if crusher_id:
        records = [r for r in records if r.crusher_id == crusher_id]
    return records


@app.post("/flotation/", response_model=FlotationRecord, status_code=201)
def create_flotation_record(data: FlotationRecordCreate):
    try:
        validate_flotation_record(data)
    except BusinessRuleError as e:
        raise HTTPException(status_code=400, detail=str(e))

    record_id = str(uuid.uuid4())
    record = FlotationRecord(id=record_id, **data.dict())
    storage.add_flotation_record(record)
    return record


@app.get("/flotation/", response_model=List[FlotationRecord])
def list_flotation_records(
    start_date: Optional[date] = None,
    end_date: Optional[date] = None,
):
    records = storage.get_all_flotation_records()
    if start_date:
        records = [r for r in records if r.date >= start_date]
    if end_date:
        records = [r for r in records if r.date <= end_date]
    return records


@app.post("/calculate/grade", response_model=GradeCalculationResult)
def calculate_grade(data: GradeCalculationInput):
    result = GradeCalculator.calculate(data)
    return result


@app.get("/metrics/daily/{target_date}", response_model=DailyMetrics)
def get_daily_metrics(target_date: date):
    aggregator = MetricsAggregator(
        crushing_records=storage.get_all_crushing_records(),
        flotation_records=storage.get_all_flotation_records(),
        equipment_list=storage.get_all_equipment(),
    )
    return aggregator.get_daily_metrics(target_date)


@app.get("/metrics/equipment/{target_date}", response_model=List[EquipmentUtilization])
def get_equipment_utilization(target_date: date):
    aggregator = MetricsAggregator(
        crushing_records=storage.get_all_crushing_records(),
        flotation_records=storage.get_all_flotation_records(),
        equipment_list=storage.get_all_equipment(),
    )
    return aggregator.get_equipment_utilization_list(target_date)


@app.exception_handler(BusinessRuleError)
async def business_rule_exception_handler(request, exc):
    return JSONResponse(
        status_code=400,
        content={"detail": str(exc)},
    )


def run_server():
    import uvicorn

    port = int(os.getenv("PORT", "8000"))
    host = os.getenv("HOST", "0.0.0.0")
    uvicorn.run("server.main:app", host=host, port=port, reload=False)


if __name__ == "__main__":
    run_server()
