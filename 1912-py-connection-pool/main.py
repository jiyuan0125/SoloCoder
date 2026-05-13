import os
from fastapi import FastAPI, HTTPException, status
from fastapi.responses import JSONResponse
from pydantic import BaseModel
from typing import List, Optional
from contextlib import asynccontextmanager

from connection_pool.models import DatasourceCreate, DatasourceStats, DatasourceOverview
from connection_pool.manager import get_manager


@asynccontextmanager
async def lifespan(app: FastAPI):
    manager = get_manager()
    manager.start_background_task()
    yield


app = FastAPI(title="Connection Pool Service", lifespan=lifespan)


class UpdateMaxConnectionsRequest(BaseModel):
    max_connections: int


@app.post("/datasources", status_code=status.HTTP_201_CREATED)
async def register_datasource(ds: DatasourceCreate):
    manager = get_manager()
    if not manager.register_datasource(ds):
        raise HTTPException(status_code=status.HTTP_409_CONFLICT, detail=f"Datasource '{ds.name}' already exists")
    return {"message": f"Datasource '{ds.name}' registered successfully"}


@app.get("/datasources", response_model=List[DatasourceOverview])
async def list_datasources():
    manager = get_manager()
    return manager.list_datasources()


@app.get("/datasources/{name}/stats", response_model=DatasourceStats)
async def get_datasource_stats(name: str):
    manager = get_manager()
    stats = manager.get_datasource_stats(name)
    if not stats:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=f"Datasource '{name}' not found")
    return stats


@app.put("/datasources/{name}/max-connections")
async def update_max_connections(name: str, request: UpdateMaxConnectionsRequest):
    if request.max_connections <= 0:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail="max_connections must be greater than 0")
    manager = get_manager()
    if not manager.update_max_connections(name, request.max_connections):
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=f"Datasource '{name}' not found")
    return {"message": f"max_connections updated to {request.max_connections}"}


if __name__ == "__main__":
    import uvicorn
    port = int(os.environ.get("PORT", 8000))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=False)
