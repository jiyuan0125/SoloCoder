import os
from contextlib import asynccontextmanager
from fastapi import FastAPI, HTTPException, Query
from typing import Optional, List
from datetime import date

from ..core.models import (
    Plot, PlotCreate,
    HarvestPlan, HarvestPlanCreate,
    HarvestRecord, HarvestRecordCreate,
    ProcessingBatch, ProcessingBatchCreate, ProcessingBatchUpdate,
    QualityEvaluation, QualityEvaluationCreate,
    Alert, AlertType
)
from ..core.storage import storage
from ..core.scheduler import scheduler


@asynccontextmanager
async def lifespan(app: FastAPI):
    scheduler.start()
    yield
    scheduler.stop()


app = FastAPI(
    title="茶叶加工厂管理系统 API",
    description="从采摘到分级的茶叶加工管理后端服务",
    version="1.0.0",
    lifespan=lifespan
)


@app.get("/")
def root():
    return {"message": "茶叶加工厂管理系统 API 运行中", "status": "ok"}


@app.post("/plots/", response_model=Plot, status_code=201)
def create_plot(plot_create: PlotCreate):
    return storage.create_plot(plot_create)


@app.get("/plots/", response_model=List[Plot])
def list_plots():
    return storage.list_plots()


@app.get("/plots/{plot_id}", response_model=Plot)
def get_plot(plot_id: int):
    plot = storage.get_plot(plot_id)
    if not plot:
        raise HTTPException(status_code=404, detail=f"茶园地块 ID {plot_id} 不存在")
    return plot


@app.post("/harvest-plans/", response_model=HarvestPlan, status_code=201)
def create_harvest_plan(plan_create: HarvestPlanCreate):
    plot = storage.get_plot(plan_create.plot_id)
    if not plot:
        raise HTTPException(status_code=400, detail=f"茶园地块 ID {plan_create.plot_id} 不存在")
    try:
        return storage.create_harvest_plan(plan_create)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/harvest-plans/", response_model=List[HarvestPlan])
def list_harvest_plans(
    plot_id: Optional[int] = Query(None, description="按地块筛选"),
    plan_date: Optional[date] = Query(None, description="按计划日期筛选")
):
    return storage.list_harvest_plans(plot_id, plan_date)


@app.get("/harvest-plans/{plan_id}", response_model=HarvestPlan)
def get_harvest_plan(plan_id: int):
    plan = storage.get_harvest_plan(plan_id)
    if not plan:
        raise HTTPException(status_code=404, detail=f"采摘计划 ID {plan_id} 不存在")
    return plan


@app.post("/harvest-records/", response_model=HarvestRecord, status_code=201)
def create_harvest_record(record_create: HarvestRecordCreate):
    plan = storage.get_harvest_plan(record_create.plan_id)
    if not plan:
        raise HTTPException(status_code=400, detail=f"采摘计划 ID {record_create.plan_id} 不存在")
    return storage.create_harvest_record(record_create)


@app.get("/harvest-records/", response_model=List[HarvestRecord])
def list_harvest_records(
    plan_id: Optional[int] = Query(None, description="按计划筛选")
):
    return storage.list_harvest_records(plan_id)


@app.get("/harvest-records/{record_id}", response_model=HarvestRecord)
def get_harvest_record(record_id: int):
    record = storage.get_harvest_record(record_id)
    if not record:
        raise HTTPException(status_code=404, detail=f"采摘记录 ID {record_id} 不存在")
    return record


@app.post("/processing-batches/", response_model=ProcessingBatch, status_code=201)
def create_processing_batch(batch_create: ProcessingBatchCreate):
    record = storage.get_harvest_record(batch_create.harvest_record_id)
    if not record:
        raise HTTPException(status_code=400, detail=f"采摘记录 ID {batch_create.harvest_record_id} 不存在")
    
    if storage.has_processing_batch_for_harvest(batch_create.harvest_record_id):
        raise HTTPException(
            status_code=400,
            detail=f"该采摘记录（ID {batch_create.harvest_record_id}）已存在炒制批次，无法重复创建"
        )
    
    return storage.create_processing_batch(batch_create)


@app.get("/processing-batches/", response_model=List[ProcessingBatch])
def list_processing_batches(
    harvest_record_id: Optional[int] = Query(None, description="按采摘记录筛选")
):
    return storage.list_processing_batches(harvest_record_id)


@app.get("/processing-batches/{batch_id}", response_model=ProcessingBatch)
def get_processing_batch(batch_id: int):
    batch = storage.get_processing_batch(batch_id)
    if not batch:
        raise HTTPException(status_code=404, detail=f"炒制批次 ID {batch_id} 不存在")
    return batch


@app.put("/processing-batches/{batch_id}/complete", response_model=ProcessingBatch)
def complete_processing_batch(batch_id: int, batch_update: ProcessingBatchUpdate):
    batch = storage.get_processing_batch(batch_id)
    if not batch:
        raise HTTPException(status_code=404, detail=f"炒制批次 ID {batch_id} 不存在")
    
    if batch.is_completed:
        raise HTTPException(status_code=400, detail=f"炒制批次 ID {batch_id} 已完成")
    
    try:
        updated = storage.update_processing_batch(batch_id, batch_update)
        if not updated:
            raise HTTPException(status_code=404, detail=f"炒制批次 ID {batch_id} 不存在")
        return updated
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.post("/quality-evaluations/", response_model=QualityEvaluation, status_code=201)
def create_quality_evaluation(eval_create: QualityEvaluationCreate):
    batch = storage.get_processing_batch(eval_create.batch_id)
    if not batch:
        raise HTTPException(status_code=400, detail=f"炒制批次 ID {eval_create.batch_id} 不存在")
    
    if not batch.is_completed:
        raise HTTPException(status_code=400, detail=f"炒制批次 ID {eval_create.batch_id} 尚未完成，无法评定等级")
    
    try:
        return storage.create_quality_evaluation(eval_create)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/quality-evaluations/", response_model=List[QualityEvaluation])
def list_quality_evaluations(
    batch_id: Optional[int] = Query(None, description="按炒制批次筛选")
):
    return storage.list_quality_evaluations(batch_id)


@app.get("/quality-evaluations/{eval_id}", response_model=QualityEvaluation)
def get_quality_evaluation(eval_id: int):
    evaluation = storage.get_quality_evaluation(eval_id)
    if not evaluation:
        raise HTTPException(status_code=404, detail=f"质量评定 ID {eval_id} 不存在")
    return evaluation


@app.get("/alerts/", response_model=List[Alert])
def list_alerts(
    alert_type: Optional[AlertType] = Query(None, description="按告警类型筛选"),
    is_read: Optional[bool] = Query(None, description="按是否已读筛选")
):
    return storage.list_alerts(alert_type, is_read)


@app.put("/alerts/{alert_id}/read", response_model=Alert)
def mark_alert_read(alert_id: int):
    alert = storage.mark_alert_read(alert_id)
    if not alert:
        raise HTTPException(status_code=404, detail=f"告警 ID {alert_id} 不存在")
    return alert


def main():
    import uvicorn
    port = int(os.getenv("PORT", "8000"))
    uvicorn.run(
        "src.server.app:app",
        host="0.0.0.0",
        port=port,
        reload=False
    )


if __name__ == "__main__":
    main()
