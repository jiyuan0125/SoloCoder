import os
from fastapi import FastAPI, HTTPException, Query
from fastapi.responses import PlainTextResponse
from typing import List, Optional

from src.core.storage import Storage
from src.core.service import ProductionService
from src.core.export import DataExporter
from src.core.models import (
    Plot,
    PlotCreate,
    PlotUpdate,
    FarmOperation,
    FarmOperationCreate,
    Harvest,
    HarvestCreate,
    HarvestUpdate,
    Processing,
    ProcessingCreate,
    ProcessingUpdate,
    Todo,
    TodoUpdate,
)
from src.core.exceptions import (
    GAPValidationError,
    NotFoundError,
)

app = FastAPI(title="中药材GAP生产管理系统", version="1.0.0")

storage = Storage()
service = ProductionService(storage)
exporter = DataExporter(storage)


@app.on_event("startup")
def on_startup():
    service.refresh_todos()


@app.get("/api/health")
def health_check():
    return {"status": "ok", "service": "GAP Production Management"}


@app.post("/api/plots", response_model=Plot)
def create_plot(data: PlotCreate):
    try:
        return service.create_plot(data)
    except GAPValidationError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/api/plots", response_model=List[Plot])
def list_plots():
    return service.get_plots()


@app.get("/api/plots/{plot_id}", response_model=Plot)
def get_plot(plot_id: str):
    try:
        return service.get_plot(plot_id)
    except NotFoundError as e:
        raise HTTPException(status_code=404, detail=e.message)


@app.put("/api/plots/{plot_id}", response_model=Plot)
def update_plot(plot_id: str, data: PlotUpdate):
    try:
        return service.update_plot(plot_id, data)
    except NotFoundError as e:
        raise HTTPException(status_code=404, detail=e.message)
    except GAPValidationError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.delete("/api/plots/{plot_id}")
def delete_plot(plot_id: str):
    try:
        service.delete_plot(plot_id)
        return {"message": "Plot deleted successfully"}
    except NotFoundError as e:
        raise HTTPException(status_code=404, detail=e.message)


@app.post("/api/operations", response_model=FarmOperation)
def create_operation(data: FarmOperationCreate):
    try:
        return service.create_operation(data)
    except NotFoundError as e:
        raise HTTPException(status_code=404, detail=e.message)


@app.get("/api/operations", response_model=List[FarmOperation])
def list_operations(plot_id: str = Query(..., description="地块ID")):
    try:
        return service.get_operations(plot_id)
    except NotFoundError as e:
        raise HTTPException(status_code=404, detail=e.message)


@app.post("/api/harvests", response_model=Harvest)
def create_harvest(data: HarvestCreate):
    try:
        return service.create_harvest(data)
    except NotFoundError as e:
        raise HTTPException(status_code=404, detail=e.message)
    except GAPValidationError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/api/harvests", response_model=List[Harvest])
def list_harvests(plot_id: Optional[str] = Query(None, description="地块ID（可选）")):
    return service.get_harvests(plot_id)


@app.get("/api/harvests/{harvest_id}", response_model=Harvest)
def get_harvest(harvest_id: str):
    try:
        return service.get_harvest(harvest_id)
    except NotFoundError as e:
        raise HTTPException(status_code=404, detail=e.message)


@app.put("/api/harvests/{harvest_id}", response_model=Harvest)
def update_harvest(harvest_id: str, data: HarvestUpdate):
    try:
        return service.update_harvest(harvest_id, data)
    except NotFoundError as e:
        raise HTTPException(status_code=404, detail=e.message)
    except GAPValidationError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.post("/api/processings", response_model=Processing)
def create_processing(data: ProcessingCreate):
    try:
        return service.create_processing(data)
    except NotFoundError as e:
        raise HTTPException(status_code=404, detail=e.message)
    except GAPValidationError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/api/processings", response_model=List[Processing])
def list_processings(harvest_id: Optional[str] = Query(None, description="采收ID（可选）")):
    return service.get_processings(harvest_id)


@app.put("/api/processings/{processing_id}", response_model=Processing)
def update_processing(processing_id: str, data: ProcessingUpdate):
    try:
        return service.update_processing(processing_id, data)
    except NotFoundError as e:
        raise HTTPException(status_code=404, detail=e.message)
    except GAPValidationError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/api/todos", response_model=List[Todo])
def list_todos():
    service.refresh_todos()
    return service.get_todos()


@app.put("/api/todos/{todo_id}", response_model=Todo)
def update_todo(todo_id: str, data: TodoUpdate):
    try:
        return service.update_todo(todo_id, data)
    except NotFoundError as e:
        raise HTTPException(status_code=404, detail=e.message)


@app.post("/api/todos/refresh")
def refresh_todos():
    service.refresh_todos()
    return {"message": "Todos refreshed successfully"}


@app.get("/api/export/batch/{harvest_id}", response_class=PlainTextResponse)
def export_batch(harvest_id: str):
    content = exporter.export_batch(harvest_id)
    if content.startswith("Error"):
        raise HTTPException(status_code=404, detail=content)
    return PlainTextResponse(
        content=content,
        media_type="text/plain; charset=utf-8",
        headers={"Content-Disposition": f"attachment; filename=gap_batch_{harvest_id}.txt"}
    )


@app.get("/api/export/all", response_class=PlainTextResponse)
def export_all():
    content = exporter.export_all()
    return PlainTextResponse(
        content=content,
        media_type="text/plain; charset=utf-8",
        headers={"Content-Disposition": "attachment; filename=gap_all_records.txt"}
    )


def run_server():
    import uvicorn
    port = int(os.environ.get("GAP_SERVER_PORT", "8000"))
    host = os.environ.get("GAP_SERVER_HOST", "0.0.0.0")
    uvicorn.run(app, host=host, port=port)


if __name__ == "__main__":
    run_server()
