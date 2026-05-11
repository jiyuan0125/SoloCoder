from typing import Optional
from fastapi import FastAPI, HTTPException, Depends
from pydantic import BaseModel
from datetime import date, datetime

from sys import path
from pathlib import Path

src_path = str(Path(__file__).resolve().parent.parent.parent / "src")
if src_path not in path:
    path.insert(0, src_path)

from core import (
    Storage,
    PlotService,
    HarvestService,
    ProcessingService,
    SchedulerService,
    FreshLeafGrade,
    FinishedGrade,
    TodoStatus,
    AlertStatus,
    ValidationException,
    NotFoundException,
    ConflictException,
)
from .config import settings


_storage: Optional[Storage] = None


def get_storage() -> Storage:
    global _storage
    if _storage is None:
        _storage = Storage()
    return _storage


def get_services(storage: Storage = Depends(get_storage)):
    return {
        "plot": PlotService(storage),
        "harvest": HarvestService(storage),
        "processing": ProcessingService(storage),
        "scheduler": SchedulerService(storage),
    }


class PlotCreate(BaseModel):
    name: str
    location: str
    area: float


class PlotUpdate(BaseModel):
    name: Optional[str] = None
    location: Optional[str] = None
    area: Optional[float] = None


class HarvestPlanCreate(BaseModel):
    plot_id: str
    plan_date: date
    expected_quantity: float


class HarvestPlanUpdate(BaseModel):
    plan_date: Optional[date] = None
    expected_quantity: Optional[float] = None


class HarvestRecordCreate(BaseModel):
    plan_id: str
    actual_quantity: float
    fresh_leaf_grade: FreshLeafGrade
    harvest_time: Optional[datetime] = None


class HarvestRecordUpdate(BaseModel):
    actual_quantity: Optional[float] = None
    fresh_leaf_grade: Optional[FreshLeafGrade] = None
    harvest_time: Optional[datetime] = None


class ProcessingBatchCreate(BaseModel):
    harvest_record_id: str
    input_quantity: float
    start_time: Optional[datetime] = None


class ProcessingBatchComplete(BaseModel):
    output_quantity: float


class QualityRatingCreate(BaseModel):
    batch_id: str
    grade: FinishedGrade
    sensory_description: str


class QualityRatingUpdate(BaseModel):
    grade: Optional[FinishedGrade] = None
    sensory_description: Optional[str] = None


def create_app() -> FastAPI:
    app = FastAPI(
        title="茶叶加工厂管理系统",
        description="从采摘到分级的全流程管理后端",
        version="0.1.0"
    )

    @app.exception_handler(ValidationException)
    async def validation_exception_handler(request, exc):
        raise HTTPException(status_code=400, detail={"error": "ValidationError", "message": exc.message})

    @app.exception_handler(NotFoundException)
    async def not_found_exception_handler(request, exc):
        raise HTTPException(status_code=404, detail={"error": "NotFound", "message": exc.message})

    @app.exception_handler(ConflictException)
    async def conflict_exception_handler(request, exc):
        raise HTTPException(status_code=409, detail={"error": "Conflict", "message": exc.message})

    @app.get("/")
    def root():
        return {"name": "茶叶加工厂管理系统", "version": "0.1.0"}

    @app.get("/plots")
    def list_plots(services: dict = Depends(get_services)):
        return {"plots": services["plot"].list()}

    @app.post("/plots")
    def create_plot(data: PlotCreate, services: dict = Depends(get_services)):
        return services["plot"].create(
            name=data.name,
            location=data.location,
            area=data.area
        )

    @app.get("/plots/{plot_id}")
    def get_plot(plot_id: str, services: dict = Depends(get_services)):
        return services["plot"].get(plot_id)

    @app.put("/plots/{plot_id}")
    def update_plot(plot_id: str, data: PlotUpdate, services: dict = Depends(get_services)):
        return services["plot"].update(
            plot_id=plot_id,
            name=data.name,
            location=data.location,
            area=data.area
        )

    @app.delete("/plots/{plot_id}")
    def delete_plot(plot_id: str, services: dict = Depends(get_services)):
        services["plot"].delete(plot_id)
        return {"success": True}

    @app.get("/harvest-plans")
    def list_harvest_plans(services: dict = Depends(get_services)):
        return {"plans": services["harvest"].list_plans()}

    @app.post("/harvest-plans")
    def create_harvest_plan(data: HarvestPlanCreate, services: dict = Depends(get_services)):
        return services["harvest"].create_plan(
            plot_id=data.plot_id,
            plan_date=data.plan_date,
            expected_quantity=data.expected_quantity
        )

    @app.get("/harvest-plans/{plan_id}")
    def get_harvest_plan(plan_id: str, services: dict = Depends(get_services)):
        return services["harvest"].get_plan(plan_id)

    @app.put("/harvest-plans/{plan_id}")
    def update_harvest_plan(plan_id: str, data: HarvestPlanUpdate, services: dict = Depends(get_services)):
        return services["harvest"].update_plan(
            plan_id=plan_id,
            plan_date=data.plan_date,
            expected_quantity=data.expected_quantity
        )

    @app.delete("/harvest-plans/{plan_id}")
    def delete_harvest_plan(plan_id: str, services: dict = Depends(get_services)):
        services["harvest"].delete_plan(plan_id)
        return {"success": True}

    @app.get("/harvest-records")
    def list_harvest_records(services: dict = Depends(get_services)):
        return {"records": services["harvest"].list_records()}

    @app.post("/harvest-records")
    def create_harvest_record(data: HarvestRecordCreate, services: dict = Depends(get_services)):
        return services["harvest"].create_record(
            plan_id=data.plan_id,
            actual_quantity=data.actual_quantity,
            fresh_leaf_grade=data.fresh_leaf_grade,
            harvest_time=data.harvest_time
        )

    @app.get("/harvest-records/{record_id}")
    def get_harvest_record(record_id: str, services: dict = Depends(get_services)):
        return services["harvest"].get_record(record_id)

    @app.put("/harvest-records/{record_id}")
    def update_harvest_record(record_id: str, data: HarvestRecordUpdate, services: dict = Depends(get_services)):
        return services["harvest"].update_record(
            record_id=record_id,
            actual_quantity=data.actual_quantity,
            fresh_leaf_grade=data.fresh_leaf_grade,
            harvest_time=data.harvest_time
        )

    @app.delete("/harvest-records/{record_id}")
    def delete_harvest_record(record_id: str, services: dict = Depends(get_services)):
        services["harvest"].delete_record(record_id)
        return {"success": True}

    @app.get("/processing-batches")
    def list_processing_batches(services: dict = Depends(get_services)):
        return {"batches": services["processing"].list_batches()}

    @app.post("/processing-batches")
    def create_processing_batch(data: ProcessingBatchCreate, services: dict = Depends(get_services)):
        return services["processing"].create_batch(
            harvest_record_id=data.harvest_record_id,
            input_quantity=data.input_quantity,
            start_time=data.start_time
        )

    @app.get("/processing-batches/{batch_id}")
    def get_processing_batch(batch_id: str, services: dict = Depends(get_services)):
        return services["processing"].get_batch(batch_id)

    @app.post("/processing-batches/{batch_id}/complete")
    def complete_processing_batch(batch_id: str, data: ProcessingBatchComplete, services: dict = Depends(get_services)):
        return services["processing"].complete_batch(
            batch_id=batch_id,
            output_quantity=data.output_quantity
        )

    @app.delete("/processing-batches/{batch_id}")
    def delete_processing_batch(batch_id: str, services: dict = Depends(get_services)):
        services["processing"].delete_batch(batch_id)
        return {"success": True}

    @app.get("/quality-ratings")
    def list_quality_ratings(services: dict = Depends(get_services)):
        return {"ratings": services["processing"].list_ratings()}

    @app.post("/quality-ratings")
    def create_quality_rating(data: QualityRatingCreate, services: dict = Depends(get_services)):
        return services["processing"].create_rating(
            batch_id=data.batch_id,
            grade=data.grade,
            sensory_description=data.sensory_description
        )

    @app.get("/quality-ratings/{rating_id}")
    def get_quality_rating(rating_id: str, services: dict = Depends(get_services)):
        return services["processing"].get_rating(rating_id)

    @app.put("/quality-ratings/{rating_id}")
    def update_quality_rating(rating_id: str, data: QualityRatingUpdate, services: dict = Depends(get_services)):
        return services["processing"].update_rating(
            rating_id=rating_id,
            grade=data.grade,
            sensory_description=data.sensory_description
        )

    @app.delete("/quality-ratings/{rating_id}")
    def delete_quality_rating(rating_id: str, services: dict = Depends(get_services)):
        services["processing"].delete_rating(rating_id)
        return {"success": True}

    @app.post("/scheduler/run")
    def run_scheduler(services: dict = Depends(get_services)):
        return services["scheduler"].run_scheduler()

    @app.get("/todos")
    def list_todos(status: Optional[TodoStatus] = None, services: dict = Depends(get_services)):
        return {"todos": services["scheduler"].list_todos(status)}

    @app.post("/todos/{todo_id}/complete")
    def complete_todo(todo_id: str, services: dict = Depends(get_services)):
        return services["scheduler"].complete_todo(todo_id)

    @app.get("/alerts")
    def list_alerts(status: Optional[AlertStatus] = None, services: dict = Depends(get_services)):
        return {"alerts": services["scheduler"].list_alerts(status)}

    @app.post("/alerts/{alert_id}/resolve")
    def resolve_alert(alert_id: str, services: dict = Depends(get_services)):
        return services["scheduler"].resolve_alert(alert_id)

    return app


def main():
    import uvicorn
    app = create_app()
    port = settings.port_from_env
    uvicorn.run(
        app,
        host=settings.host,
        port=port
    )


if __name__ == "__main__":
    main()
