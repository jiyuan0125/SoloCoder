import os
import logging
from datetime import datetime
from typing import Optional, List
from fastapi import FastAPI, Depends, HTTPException, Query, Request
from fastapi.responses import JSONResponse
from fastapi.exceptions import RequestValidationError
from pydantic import BaseModel, Field, field_validator
from sqlalchemy.orm import Session

from database import engine, get_db
from models import Base, Config, ConfigVersion, Subscriber
from notification_service import notify_subscribers_with_retry

Base.metadata.create_all(bind=engine)

app = FastAPI(title="配置中心", version="1.0.0")

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

MAX_VALUE_SIZE = 10 * 1024


class ConfigCreate(BaseModel):
    app_name: str = Field(..., description="应用名称")
    key: str = Field(..., description="配置键")
    value: str = Field(..., description="配置值")

    @field_validator('key')
    @classmethod
    def validate_key(cls, v: str) -> str:
        if not v or v.strip() == '':
            raise ValueError('key不能为空')
        if ' ' in v or '\n' in v:
            raise ValueError('key不能包含空格或换行')
        return v

    @field_validator('value')
    @classmethod
    def validate_value_size(cls, v: str) -> str:
        if len(v.encode('utf-8')) > MAX_VALUE_SIZE:
            raise ValueError('配置值超过10KB限制')
        return v


class ConfigUpdate(BaseModel):
    value: str = Field(..., description="配置值")

    @field_validator('value')
    @classmethod
    def validate_value_size(cls, v: str) -> str:
        if len(v.encode('utf-8')) > MAX_VALUE_SIZE:
            raise ValueError('配置值超过10KB限制')
        return v


class ConfigResponse(BaseModel):
    app_name: str
    key: str
    value: str
    version: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class ConfigVersionResponse(BaseModel):
    app_name: str
    key: str
    value: str
    version: int
    created_at: datetime

    class Config:
        from_attributes = True


class SubscriberCreate(BaseModel):
    app_name: str = Field(..., description="应用名称")
    callback_url: str = Field(..., description="回调地址")


class SubscriberResponse(BaseModel):
    id: int
    app_name: str
    callback_url: str
    created_at: datetime

    class Config:
        from_attributes = True


class RollbackRequest(BaseModel):
    version: int = Field(..., description="要回滚到的版本号")


@app.post("/configs", response_model=ConfigResponse)
async def create_config(config: ConfigCreate, db: Session = Depends(get_db)):
    existing = db.query(Config).filter(
        Config.app_name == config.app_name,
        Config.key == config.key
    ).first()

    if existing:
        raise HTTPException(
            status_code=409,
            detail=f"配置已存在: {config.app_name}/{config.key}"
        )

    db_config = Config(
        app_name=config.app_name,
        key=config.key,
        value=config.value,
        version=1
    )
    db.add(db_config)
    db.commit()
    db.refresh(db_config)

    db_version = ConfigVersion(
        config_id=db_config.id,
        app_name=config.app_name,
        key=config.key,
        value=config.value,
        version=1
    )
    db.add(db_version)
    db.commit()

    await notify_subscribers_with_retry(
        db=db,
        config_id=db_config.id,
        app_name=config.app_name,
        key=config.key,
        value=config.value,
        version=1
    )

    return db_config


@app.get("/configs/{app_name}/{key}", response_model=ConfigResponse)
async def read_config(
    app_name: str,
    key: str,
    version: Optional[int] = None,
    db: Session = Depends(get_db)
):
    if version is not None:
        db_version = db.query(ConfigVersion).filter(
            ConfigVersion.app_name == app_name,
            ConfigVersion.key == key,
            ConfigVersion.version == version
        ).first()

        if not db_version:
            raise HTTPException(
                status_code=404,
                detail=f"配置版本不存在: {app_name}/{key} v{version}"
            )

        return ConfigResponse(
            app_name=db_version.app_name,
            key=db_version.key,
            value=db_version.value,
            version=db_version.version,
            created_at=db_version.created_at,
            updated_at=db_version.created_at
        )

    db_config = db.query(Config).filter(
        Config.app_name == app_name,
        Config.key == key
    ).first()

    if not db_config:
        raise HTTPException(
            status_code=404,
            detail=f"配置不存在: {app_name}/{key}"
        )

    return db_config


@app.put("/configs/{app_name}/{key}", response_model=ConfigResponse)
async def update_config(
    app_name: str,
    key: str,
    config_update: ConfigUpdate,
    db: Session = Depends(get_db)
):
    db_config = db.query(Config).filter(
        Config.app_name == app_name,
        Config.key == key
    ).first()

    if not db_config:
        raise HTTPException(
            status_code=404,
            detail=f"配置不存在: {app_name}/{key}"
        )

    new_version = db_config.version + 1
    db_config.value = config_update.value
    db_config.version = new_version
    db_config.updated_at = datetime.utcnow()
    db.commit()
    db.refresh(db_config)

    db_version = ConfigVersion(
        config_id=db_config.id,
        app_name=app_name,
        key=key,
        value=config_update.value,
        version=new_version
    )
    db.add(db_version)
    db.commit()

    await notify_subscribers_with_retry(
        db=db,
        config_id=db_config.id,
        app_name=app_name,
        key=key,
        value=config_update.value,
        version=new_version
    )

    return db_config


@app.delete("/configs/{app_name}/{key}")
async def delete_config(
    app_name: str,
    key: str,
    db: Session = Depends(get_db)
):
    db_config = db.query(Config).filter(
        Config.app_name == app_name,
        Config.key == key
    ).first()

    if not db_config:
        raise HTTPException(
            status_code=404,
            detail=f"配置不存在: {app_name}/{key}"
        )

    db.delete(db_config)
    db.commit()

    return {"message": "配置已删除"}


@app.get("/configs/{app_name}/{key}/history", response_model=List[ConfigVersionResponse])
async def get_config_history(
    app_name: str,
    key: str,
    start_time: Optional[datetime] = Query(None, description="开始时间"),
    end_time: Optional[datetime] = Query(None, description="结束时间"),
    db: Session = Depends(get_db)
):
    query = db.query(ConfigVersion).filter(
        ConfigVersion.app_name == app_name,
        ConfigVersion.key == key
    )

    if start_time:
        query = query.filter(ConfigVersion.created_at >= start_time)
    if end_time:
        query = query.filter(ConfigVersion.created_at <= end_time)

    versions = query.order_by(ConfigVersion.version.desc()).all()

    if not versions:
        raise HTTPException(
            status_code=404,
            detail=f"配置历史不存在: {app_name}/{key}"
        )

    return versions


@app.post("/configs/{app_name}/{key}/rollback", response_model=ConfigResponse)
async def rollback_config(
    app_name: str,
    key: str,
    rollback: RollbackRequest,
    db: Session = Depends(get_db)
):
    target_version = db.query(ConfigVersion).filter(
        ConfigVersion.app_name == app_name,
        ConfigVersion.key == key,
        ConfigVersion.version == rollback.version
    ).first()

    if not target_version:
        raise HTTPException(
            status_code=404,
            detail=f"版本不存在: v{rollback.version}"
        )

    db_config = db.query(Config).filter(
        Config.app_name == app_name,
        Config.key == key
    ).first()

    if not db_config:
        raise HTTPException(
            status_code=404,
            detail=f"配置不存在: {app_name}/{key}"
        )

    new_version = db_config.version + 1
    db_config.value = target_version.value
    db_config.version = new_version
    db_config.updated_at = datetime.utcnow()
    db.commit()
    db.refresh(db_config)

    db_new_version = ConfigVersion(
        config_id=db_config.id,
        app_name=app_name,
        key=key,
        value=target_version.value,
        version=new_version
    )
    db.add(db_new_version)
    db.commit()

    await notify_subscribers_with_retry(
        db=db,
        config_id=db_config.id,
        app_name=app_name,
        key=key,
        value=target_version.value,
        version=new_version
    )

    return db_config


@app.get("/configs/{app_name}", response_model=List[ConfigResponse])
async def list_app_configs(
    app_name: str,
    db: Session = Depends(get_db)
):
    configs = db.query(Config).filter(Config.app_name == app_name).all()
    return configs


@app.post("/subscribers", response_model=SubscriberResponse)
async def register_subscriber(
    subscriber: SubscriberCreate,
    db: Session = Depends(get_db)
):
    existing = db.query(Subscriber).filter(
        Subscriber.app_name == subscriber.app_name,
        Subscriber.callback_url == subscriber.callback_url
    ).first()

    if existing:
        raise HTTPException(
            status_code=409,
            detail="订阅者已存在"
        )

    db_subscriber = Subscriber(
        app_name=subscriber.app_name,
        callback_url=subscriber.callback_url
    )
    db.add(db_subscriber)
    db.commit()
    db.refresh(db_subscriber)

    return db_subscriber


@app.delete("/subscribers")
async def unregister_subscriber(
    app_name: str,
    callback_url: str,
    db: Session = Depends(get_db)
):
    subscriber = db.query(Subscriber).filter(
        Subscriber.app_name == app_name,
        Subscriber.callback_url == callback_url
    ).first()

    if not subscriber:
        raise HTTPException(
            status_code=404,
            detail="订阅者不存在"
        )

    db.delete(subscriber)
    db.commit()

    return {"message": "订阅者已注销"}


@app.get("/subscribers/{app_name}", response_model=List[SubscriberResponse])
async def list_subscribers(
    app_name: str,
    db: Session = Depends(get_db)
):
    subscribers = db.query(Subscriber).filter(
        Subscriber.app_name == app_name
    ).all()
    return subscribers


@app.exception_handler(RequestValidationError)
async def validation_exception_handler(request: Request, exc: RequestValidationError):
    for error in exc.errors():
        error_msg = error.get("msg", "")
        if "配置值超过10KB限制" in error_msg:
            return JSONResponse(
                status_code=400,
                content={"detail": "配置值超过 10KB 限制"}
            )
        if "key不能为空" in error_msg or "key不能包含空格或换行" in error_msg:
            return JSONResponse(
                status_code=400,
                content={"detail": error_msg.replace("Value error, ", "")}
            )
    return JSONResponse(
        status_code=422,
        content={"detail": exc.errors()}
    )


@app.exception_handler(ValueError)
async def value_error_handler(request, exc):
    if "配置值超过10KB限制" in str(exc):
        return JSONResponse(
            status_code=400,
            content={"detail": "配置值超过 10KB 限制"}
        )
    if "key不能为空" in str(exc) or "key不能包含空格或换行" in str(exc):
        return JSONResponse(
            status_code=400,
            content={"detail": str(exc)}
        )
    return JSONResponse(
        status_code=400,
        content={"detail": str(exc)}
    )


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", 8000))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=True)
