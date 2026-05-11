import asyncio
from datetime import datetime
from typing import List, Optional
from fastapi import FastAPI, HTTPException, Query
from fastapi.responses import JSONResponse

from core import (
    DataStore,
    GreenhouseService,
    Greenhouse,
    GreenhouseCreate,
    EnvironmentData,
    EnvironmentDataCreate,
    IrrigationPlan,
    IrrigationPlanCreate,
    FertilizerPlan,
    FertilizerPlanCreate,
    PlanStatus,
    AggregatedData,
)


store = DataStore()
service = GreenhouseService(store)

app = FastAPI(title="温室大棚管理系统 API", version="1.0.0")


@app.on_event("startup")
async def start_scheduler():
    asyncio.create_task(scheduler_loop())


async def scheduler_loop():
    while True:
        await asyncio.sleep(5)
        try:
            service.run_scheduler()
        except Exception:
            pass


@app.post("/api/greenhouses", response_model=Greenhouse, status_code=201)
def create_greenhouse(data: GreenhouseCreate):
    return service.create_greenhouse(data)


@app.get("/api/greenhouses", response_model=List[Greenhouse])
def list_greenhouses():
    return service.list_greenhouses()


@app.get("/api/greenhouses/{greenhouse_id}", response_model=Greenhouse)
def get_greenhouse(greenhouse_id: str):
    greenhouse = service.get_greenhouse(greenhouse_id)
    if not greenhouse:
        raise HTTPException(status_code=404, detail="大棚不存在")
    return greenhouse


@app.post(
    "/api/greenhouses/{greenhouse_id}/environment",
    response_model=EnvironmentData,
    status_code=201,
)
def report_environment_data(greenhouse_id: str, data: EnvironmentDataCreate):
    result = service.report_environment_data(greenhouse_id, data)
    if not result:
        raise HTTPException(status_code=404, detail="大棚不存在")
    return result


@app.get("/api/greenhouses/{greenhouse_id}/environment/aggregate", response_model=AggregatedData)
def aggregate_environment_data(
    greenhouse_id: str,
    metric: str = Query(..., description="指标: temperature, humidity, soil_moisture, light"),
    start_time: datetime = Query(..., description="开始时间"),
    end_time: datetime = Query(..., description="结束时间"),
):
    result = service.aggregate_environment_data(greenhouse_id, metric, start_time, end_time)
    if not result:
        raise HTTPException(status_code=404, detail="没有找到数据")
    return result


@app.post(
    "/api/greenhouses/{greenhouse_id}/irrigation-plans",
    response_model=IrrigationPlan,
    status_code=201,
)
def create_irrigation_plan(greenhouse_id: str, data: IrrigationPlanCreate):
    result = service.create_irrigation_plan(greenhouse_id, data)
    if not result:
        raise HTTPException(status_code=404, detail="大棚不存在")
    return result


@app.get("/api/irrigation-plans", response_model=List[IrrigationPlan])
def list_irrigation_plans(
    greenhouse_id: Optional[str] = None,
    status: Optional[PlanStatus] = None,
):
    return service.list_irrigation_plans(greenhouse_id, status)


@app.post(
    "/api/greenhouses/{greenhouse_id}/fertilizer-plans",
    response_model=FertilizerPlan,
    status_code=201,
)
def create_fertilizer_plan(greenhouse_id: str, data: FertilizerPlanCreate):
    result = service.create_fertilizer_plan(greenhouse_id, data)
    if not result:
        raise HTTPException(status_code=404, detail="大棚不存在")
    return result


@app.get("/api/fertilizer-plans", response_model=List[FertilizerPlan])
def list_fertilizer_plans(
    greenhouse_id: Optional[str] = None,
    status: Optional[PlanStatus] = None,
):
    return service.list_fertilizer_plans(greenhouse_id, status)


@app.post("/api/scheduler/run")
def run_scheduler():
    return service.run_scheduler()


@app.get("/health")
def health_check():
    return {"status": "ok"}
