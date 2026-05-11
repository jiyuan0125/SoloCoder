from datetime import date
from typing import Optional

from fastapi import FastAPI, HTTPException
from fastapi.responses import JSONResponse

from core import (
    PlotCreate, Plot,
    HarvestCreate, Harvest,
    BatchCreate, Batch,
    CellarCreate, Cellar,
    StorageCreate, Storage,
    TastingCreate, Tasting,
    TodoItem,
    WineryService,
)
from core.service import ValidationError


app = FastAPI(title="Winery Management API", version="1.0.0")

_service = WineryService()


@app.exception_handler(ValidationError)
async def validation_error_handler(request, exc: ValidationError):
    return JSONResponse(
        status_code=400,
        content={"detail": exc.message}
    )


@app.get("/")
def root():
    return {"name": "Winery Management API", "version": "1.0.0"}


@app.post("/plots", response_model=Plot)
def create_plot(data: PlotCreate):
    return _service.create_plot(data)


@app.get("/plots", response_model=list[Plot])
def list_plots():
    return _service.list_plots()


@app.get("/plots/{plot_id}", response_model=Plot)
def get_plot(plot_id: str):
    plot = _service.get_plot(plot_id)
    if not plot:
        raise HTTPException(status_code=404, detail="Plot not found")
    return plot


@app.post("/harvests", response_model=Harvest)
def create_harvest(data: HarvestCreate):
    return _service.create_harvest(data)


@app.get("/harvests", response_model=list[Harvest])
def list_harvests(plot_id: Optional[str] = None):
    return _service.list_harvests(plot_id)


@app.get("/harvests/{harvest_id}", response_model=Harvest)
def get_harvest(harvest_id: str):
    harvest = _service.get_harvest(harvest_id)
    if not harvest:
        raise HTTPException(status_code=404, detail="Harvest not found")
    return harvest


@app.post("/batches", response_model=Batch)
def create_batch(data: BatchCreate):
    return _service.create_batch(data)


@app.get("/batches", response_model=list[Batch])
def list_batches():
    return _service.list_batches()


@app.get("/batches/{batch_id}", response_model=Batch)
def get_batch(batch_id: str):
    batch = _service.get_batch(batch_id)
    if not batch:
        raise HTTPException(status_code=404, detail="Batch not found")
    return batch


@app.post("/batches/{batch_id}/complete-fermentation", response_model=Batch)
def complete_fermentation(batch_id: str, end_date: date):
    return _service.complete_fermentation(batch_id, end_date)


@app.post("/cellars", response_model=Cellar)
def create_cellar(data: CellarCreate):
    return _service.create_cellar(data)


@app.get("/cellars", response_model=list[Cellar])
def list_cellars():
    return _service.list_cellars()


@app.get("/cellars/{cellar_id}", response_model=Cellar)
def get_cellar(cellar_id: str):
    cellar = _service.get_cellar(cellar_id)
    if not cellar:
        raise HTTPException(status_code=404, detail="Cellar not found")
    return cellar


@app.get("/cellars/{cellar_id}/usage")
def get_cellar_usage(cellar_id: str):
    usage = _service.get_cellar_usage(cellar_id)
    return {
        "cellar_id": usage["cellar"].id,
        "cellar_name": usage["cellar"].name,
        "total_slots": usage["total_slots"],
        "occupied_count": usage["occupied_count"],
        "available_count": usage["available_count"],
        "occupied_positions": [
            {
                "shelf_number": s.shelf_number,
                "position": s.position,
                "batch_id": s.batch_id,
            }
            for s in usage["by_shelf"].values()
            for s in s
        ]
    }


@app.post("/storages", response_model=Storage)
def create_storage(data: StorageCreate):
    return _service.create_storage(data)


@app.get("/storages", response_model=list[Storage])
def list_storages(
    cellar_id: Optional[str] = None,
    batch_id: Optional[str] = None,
    active_only: bool = False
):
    return _service.list_storages(cellar_id, batch_id, active_only)


@app.get("/storages/{storage_id}", response_model=Storage)
def get_storage(storage_id: str):
    storage = _service.get_storage(storage_id)
    if not storage:
        raise HTTPException(status_code=404, detail="Storage not found")
    return storage


@app.get("/todos/pending-tastings", response_model=list[TodoItem])
def get_pending_tastings():
    return _service.get_pending_tastings()


@app.post("/tastings", response_model=Tasting)
def record_tasting(data: TastingCreate):
    return _service.record_tasting(data)


@app.get("/tastings", response_model=list[Tasting])
def list_tastings(batch_id: Optional[str] = None):
    return _service.list_tastings(batch_id)


@app.get("/tastings/{tasting_id}", response_model=Tasting)
def get_tasting(tasting_id: str):
    tasting = _service.list_tastings()
    for t in tasting:
        if t.id == tasting_id:
            return t
    raise HTTPException(status_code=404, detail="Tasting not found")
