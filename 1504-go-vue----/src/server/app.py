import os
from typing import List, Optional
from fastapi import FastAPI, HTTPException, Query
from fastapi.responses import JSONResponse
from contextlib import asynccontextmanager

import sys
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', '..'))

from src.core import (
    get_farm_manager,
    Batch, BatchCreate, BatchUpdate,
    Breed, Breeding, BreedingCreate, BreedingUpdate,
    Vaccination, VaccinationCreate,
    Slaughter, SlaughterCreate,
    Reminder, ReminderUpdate,
    DashboardMetrics
)
from src.core.exceptions import (
    FarmManagerError, StockNegativeError, SlaughterExceedsStockError,
    DuplicateBreedingError, BatchEmptyError, InvalidPriceError,
    NotFoundError, InvalidOperationError
)


@asynccontextmanager
async def lifespan(app: FastAPI):
    app.state.farm_manager = get_farm_manager()
    app.state.farm_manager.generate_reminders()
    yield


app = FastAPI(
    title="养殖场数字化管理系统",
    description="FastAPI 后端服务，用于管理养殖批次、繁育、防疫和出栏数据",
    version="1.0.0",
    lifespan=lifespan
)


def get_manager():
    return app.state.farm_manager


@app.exception_handler(FarmManagerError)
async def farm_manager_error_handler(request, exc: FarmManagerError):
    return JSONResponse(
        status_code=400,
        content={"detail": str(exc), "error_type": type(exc).__name__}
    )


@app.exception_handler(NotFoundError)
async def not_found_error_handler(request, exc: NotFoundError):
    return JSONResponse(
        status_code=404,
        content={"detail": str(exc), "error_type": type(exc).__name__}
    )


@app.get("/api/breeds", response_model=List[Breed], tags=["品种管理"])
async def list_breeds():
    manager = get_manager()
    return manager.get_breeds()


@app.get("/api/breeds/{breed_id}", response_model=Breed, tags=["品种管理"])
async def get_breed(breed_id: str):
    manager = get_manager()
    breed = manager.get_breed(breed_id)
    if not breed:
        raise NotFoundError("品种", breed_id)
    return breed


@app.post("/api/batches", response_model=Batch, tags=["批次管理"], status_code=201)
async def create_batch(batch: BatchCreate):
    manager = get_manager()
    return manager.create_batch(batch)


@app.get("/api/batches", response_model=List[Batch], tags=["批次管理"])
async def list_batches():
    manager = get_manager()
    return manager.get_batches()


@app.get("/api/batches/{batch_id}", response_model=Batch, tags=["批次管理"])
async def get_batch(batch_id: str):
    manager = get_manager()
    batch = manager.get_batch(batch_id)
    if not batch:
        raise NotFoundError("批次", batch_id)
    return batch


@app.patch("/api/batches/{batch_id}", response_model=Batch, tags=["批次管理"])
async def update_batch(batch_id: str, batch_update: BatchUpdate):
    manager = get_manager()
    return manager.update_batch(batch_id, batch_update)


@app.post("/api/breedings", response_model=Breeding, tags=["繁育管理"], status_code=201)
async def create_breeding(breeding: BreedingCreate):
    manager = get_manager()
    return manager.create_breeding(breeding)


@app.get("/api/breedings", response_model=List[Breeding], tags=["繁育管理"])
async def list_breedings(batch_id: Optional[str] = Query(None, description="过滤批次ID")):
    manager = get_manager()
    return manager.get_breedings(batch_id)


@app.get("/api/breedings/{breeding_id}", response_model=Breeding, tags=["繁育管理"])
async def get_breeding(breeding_id: str):
    manager = get_manager()
    breeding = manager.get_breeding(breeding_id)
    if not breeding:
        raise NotFoundError("配种记录", breeding_id)
    return breeding


@app.patch("/api/breedings/{breeding_id}", response_model=Breeding, tags=["繁育管理"])
async def update_breeding(breeding_id: str, breeding_update: BreedingUpdate):
    manager = get_manager()
    return manager.update_breeding(breeding_id, breeding_update)


@app.post("/api/vaccinations", response_model=Vaccination, tags=["防疫管理"], status_code=201)
async def create_vaccination(vaccination: VaccinationCreate):
    manager = get_manager()
    return manager.create_vaccination(vaccination)


@app.get("/api/vaccinations", response_model=List[Vaccination], tags=["防疫管理"])
async def list_vaccinations(batch_id: Optional[str] = Query(None, description="过滤批次ID")):
    manager = get_manager()
    return manager.get_vaccinations(batch_id)


@app.get("/api/vaccinations/{vaccination_id}", response_model=Vaccination, tags=["防疫管理"])
async def get_vaccination(vaccination_id: str):
    manager = get_manager()
    vaccination = manager.get_vaccination(vaccination_id)
    if not vaccination:
        raise NotFoundError("防疫记录", vaccination_id)
    return vaccination


@app.post("/api/slaughters", response_model=Slaughter, tags=["出栏管理"], status_code=201)
async def create_slaughter(slaughter: SlaughterCreate):
    manager = get_manager()
    return manager.create_slaughter(slaughter)


@app.get("/api/slaughters", response_model=List[Slaughter], tags=["出栏管理"])
async def list_slaughters(batch_id: Optional[str] = Query(None, description="过滤批次ID")):
    manager = get_manager()
    return manager.get_slaughters(batch_id)


@app.get("/api/slaughters/{slaughter_id}", response_model=Slaughter, tags=["出栏管理"])
async def get_slaughter(slaughter_id: str):
    manager = get_manager()
    slaughter = manager.get_slaughter(slaughter_id)
    if not slaughter:
        raise NotFoundError("出栏记录", slaughter_id)
    return slaughter


@app.post("/api/reminders/generate", response_model=List[Reminder], tags=["提醒管理"])
async def generate_reminders():
    manager = get_manager()
    return manager.generate_reminders()


@app.get("/api/reminders", response_model=List[Reminder], tags=["提醒管理"])
async def list_reminders(include_read: bool = Query(False, description="是否包含已读提醒")):
    manager = get_manager()
    return manager.get_reminders(include_read)


@app.patch("/api/reminders/{reminder_id}", response_model=Reminder, tags=["提醒管理"])
async def update_reminder(reminder_id: str, reminder_update: ReminderUpdate):
    manager = get_manager()
    return manager.update_reminder(reminder_id, reminder_update)


@app.get("/api/dashboard", response_model=DashboardMetrics, tags=["指标看板"])
async def get_dashboard():
    manager = get_manager()
    return manager.get_dashboard_metrics()


@app.get("/health", tags=["健康检查"])
async def health_check():
    return {"status": "healthy"}


if __name__ == "__main__":
    import uvicorn
    port = int(os.environ.get("PORT", 8000))
    uvicorn.run(app, host="0.0.0.0", port=port)
