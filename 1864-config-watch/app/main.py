import os
import asyncio
from fastapi import FastAPI, Depends, HTTPException, Query, BackgroundTasks
from fastapi.responses import JSONResponse
from sqlalchemy.orm import Session
from typing import List, Optional

from app.database import get_db, engine, Base
from app.models import Config, Environment, ConfigStatus
from app.schemas import ConfigCreate, ConfigUpdate, ConfigResponse, WatchCreate, WatchResponse, DiffResponse
from app import crud
from app.callback_service import notify_watches


Base.metadata.create_all(bind=engine)

app = FastAPI(title="Config Center API")


@app.post("/configs", response_model=ConfigResponse)
def create_config(config: ConfigCreate, db: Session = Depends(get_db)):
    return crud.create_config(db, config)


@app.get("/configs")
def get_configs(
    environment: Environment = Query(..., description="Environment to filter"),
    db: Session = Depends(get_db)
):
    configs = crud.get_configs_by_environment(db, environment)
    return {
        "environment": environment,
        "count": len(configs),
        "configs": [
            {
                "key": c.key,
                "value": c.value,
                "version": c.version,
                "status": c.status
            }
            for c in configs
        ]
    }


@app.get("/configs/{key}", response_model=ConfigResponse)
def get_config(key: str, environment: Environment = Query(...), db: Session = Depends(get_db)):
    config = db.query(Config).filter(
        Config.key == key,
        Config.environment == environment,
        Config.status == ConfigStatus.ACTIVE
    ).first()
    
    if not config:
        config = crud.get_config_by_key_env(db, key, environment)
    
    if not config:
        raise HTTPException(status_code=404, detail="Config not found")
    return config


@app.put("/configs/{key}", response_model=ConfigResponse)
def update_config(key: str, config: ConfigUpdate, environment: Environment = Query(...), db: Session = Depends(get_db)):
    updated = crud.update_config(db, key, environment, config)
    if not updated:
        raise HTTPException(status_code=400, detail="No draft config found to update")
    return updated


@app.post("/configs/{key}/publish", response_model=ConfigResponse)
def publish_config(
    key: str,
    background_tasks: BackgroundTasks,
    environment: Environment = Query(...),
    db: Session = Depends(get_db)
):
    published = crud.publish_config(db, key, environment)
    if not published:
        raise HTTPException(status_code=400, detail="No draft config found to publish")
    
    background_tasks.add_task(
        run_async_task,
        notify_watches,
        db,
        key,
        environment.value,
        published.value,
        published.version
    )
    
    return published


@app.post("/configs/{key}/rollback", response_model=ConfigResponse)
def rollback_config(
    key: str,
    background_tasks: BackgroundTasks,
    version: int = Query(..., description="Version to rollback to"),
    environment: Environment = Query(...),
    db: Session = Depends(get_db)
):
    rolled = crud.rollback_config(db, key, environment, version)
    if not rolled:
        raise HTTPException(status_code=400, detail="Cannot rollback to specified version")
    
    background_tasks.add_task(
        run_async_task,
        notify_watches,
        db,
        key,
        environment.value,
        rolled.value,
        rolled.version
    )
    
    return rolled


@app.get("/configs/{key}/versions", response_model=List[ConfigResponse])
def get_versions(key: str, environment: Environment = Query(...), db: Session = Depends(get_db)):
    versions = crud.get_config_versions(db, key, environment)
    if not versions:
        raise HTTPException(status_code=404, detail="No versions found for this config")
    return versions


@app.get("/configs/{key}/diff", response_model=DiffResponse)
def get_diff(
    key: str,
    from_: int = Query(..., alias="from", description="From version"),
    to: int = Query(..., description="To version"),
    environment: Environment = Query(...),
    db: Session = Depends(get_db)
):
    diff = crud.get_config_diff(db, key, environment, from_, to)
    if not diff:
        raise HTTPException(status_code=404, detail="One or both versions not found")
    return diff


@app.post("/watches", response_model=WatchResponse)
def create_watch(watch: WatchCreate, db: Session = Depends(get_db)):
    return crud.create_watch(db, watch)


@app.get("/watches/{watch_id}", response_model=WatchResponse)
def get_watch(watch_id: int, db: Session = Depends(get_db)):
    watch = crud.get_watch(db, watch_id)
    if not watch:
        raise HTTPException(status_code=404, detail="Watch not found")
    return watch


@app.put("/watches/{watch_id}/activate", response_model=WatchResponse)
def activate_watch(watch_id: int, db: Session = Depends(get_db)):
    watch = crud.activate_watch(db, watch_id)
    if not watch:
        raise HTTPException(status_code=404, detail="Watch not found")
    return watch


def run_async_task(async_func, *args):
    loop = asyncio.new_event_loop()
    asyncio.set_event_loop(loop)
    try:
        loop.run_until_complete(async_func(*args))
    finally:
        loop.close()


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", 8000))
    uvicorn.run("app.main:app", host="0.0.0.0", port=port, reload=True)
