import os
from fastapi import FastAPI, HTTPException, status
from typing import List, Optional
from datetime import date

from src.core import (
    FurnaceService, BatchingService, TemperatureService, AlertService,
    ServiceException, scheduler, metrics_service,
    CreateFurnaceRequest, CreateBatchingOrderRequest, ChargeMaterialRequest,
    TemperatureReportRequest, UpdateFurnaceStatusRequest,
    Furnace, BatchingOrder, TemperatureReading, Alert, Metrics
)

app = FastAPI(
    title="冶炼厂炉温监控和配料管理系统",
    description="系统管理多个炉子，温度监控，配料管理，调度协调",
    version="1.0.0"
)


@app.get("/", tags=["系统"])
async def root():
    return {"message": "冶炼厂监控系统 API 服务运行中", "status": "ok"}


@app.get("/health", tags=["系统"])
async def health_check():
    return {"status": "healthy"}


@app.post("/api/furnaces", response_model=Furnace, tags=["炉子管理"], status_code=status.HTTP_201_CREATED)
async def create_furnace(request: CreateFurnaceRequest):
    try:
        return FurnaceService.create_furnace(
            name=request.name,
            design_capacity=request.design_capacity,
            min_temperature=request.min_temperature,
            max_temperature=request.max_temperature
        )
    except ServiceException as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/api/furnaces", response_model=List[Furnace], tags=["炉子管理"])
async def list_furnaces():
    return FurnaceService.get_all_furnaces()


@app.get("/api/furnaces/{furnace_id}", response_model=Furnace, tags=["炉子管理"])
async def get_furnace(furnace_id: str):
    furnace = FurnaceService.get_furnace(furnace_id)
    if not furnace:
        raise HTTPException(status_code=404, detail="炉子不存在")
    return furnace


@app.post("/api/batching-orders", response_model=BatchingOrder, tags=["配料管理"], status_code=status.HTTP_201_CREATED)
async def create_batching_order(request: CreateBatchingOrderRequest):
    try:
        return BatchingService.create_order(materials=request.materials)
    except ServiceException as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/api/batching-orders", response_model=List[BatchingOrder], tags=["配料管理"])
async def list_batching_orders():
    return BatchingService.get_all_orders()


@app.get("/api/batching-orders/{order_id}", response_model=BatchingOrder, tags=["配料管理"])
async def get_batching_order(order_id: str):
    order = BatchingService.get_order(order_id)
    if not order:
        raise HTTPException(status_code=404, detail="配料单不存在")
    return order


@app.post("/api/batching-orders/{order_id}/assign/{furnace_id}", response_model=BatchingOrder, tags=["配料管理"])
async def assign_order_to_furnace(order_id: str, furnace_id: str):
    try:
        return BatchingService.assign_order_to_furnace(order_id, furnace_id)
    except ServiceException as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.post("/api/furnaces/{furnace_id}/charge", response_model=BatchingOrder, tags=["配料管理"])
async def charge_materials(furnace_id: str, request: ChargeMaterialRequest):
    try:
        return BatchingService.charge_materials(furnace_id, request.materials)
    except ServiceException as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.post("/api/furnaces/{furnace_id}/complete", response_model=BatchingOrder, tags=["配料管理"])
async def complete_smelting(furnace_id: str, request: UpdateFurnaceStatusRequest):
    try:
        if request.output_weight is None or request.energy_consumed is None:
            raise HTTPException(status_code=400, detail="产出重量和能耗不能为空")
        return BatchingService.complete_smelting(
            furnace_id, request.output_weight, request.energy_consumed
        )
    except ServiceException as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.post("/api/furnaces/{furnace_id}/idle", response_model=Furnace, tags=["炉子管理"])
async def set_furnace_idle(furnace_id: str):
    try:
        return BatchingService.set_furnace_idle(furnace_id)
    except ServiceException as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.post("/api/temperature", tags=["温度监控"])
async def report_temperature(request: TemperatureReportRequest):
    try:
        reading, alert = TemperatureService.report_temperature(
            request.furnace_id, request.temperature
        )
        return {
            "reading": reading,
            "alert": alert
        }
    except ServiceException as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/api/furnaces/{furnace_id}/temperatures", response_model=List[TemperatureReading], tags=["温度监控"])
async def get_furnace_temperatures(furnace_id: str, limit: int = 100):
    return TemperatureService.get_furnace_temperatures(furnace_id, limit)


@app.get("/api/alerts", response_model=List[Alert], tags=["告警管理"])
async def list_alerts(unresolved_only: bool = False):
    return AlertService.get_all_alerts(unresolved_only)


@app.post("/api/alerts/{alert_id}/resolve", response_model=Alert, tags=["告警管理"])
async def resolve_alert(alert_id: str):
    alert = AlertService.resolve_alert(alert_id)
    if not alert:
        raise HTTPException(status_code=404, detail="告警不存在")
    return alert


@app.post("/api/scheduler/run", tags=["调度器"])
async def run_scheduler():
    assigned_orders = scheduler.schedule()
    return {
        "assigned_count": len(assigned_orders),
        "assigned_orders": assigned_orders
    }


@app.get("/api/metrics", response_model=Metrics, tags=["指标统计"])
async def get_daily_metrics(date_str: Optional[str] = None):
    try:
        target_date = None
        if date_str:
            target_date = date.fromisoformat(date_str)
        return metrics_service.get_daily_metrics(target_date)
    except ValueError:
        raise HTTPException(status_code=400, detail="日期格式错误，应为 YYYY-MM-DD")


def get_app():
    return app


def run_server():
    import uvicorn
    
    port = int(os.getenv("PORT", "8000"))
    host = os.getenv("HOST", "0.0.0.0")
    
    uvicorn.run(
        "src.server.app:app",
        host=host,
        port=port,
        reload=False
    )


if __name__ == "__main__":
    run_server()
