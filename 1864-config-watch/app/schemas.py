from pydantic import BaseModel, Field
from typing import Optional, Dict, Any, List
from datetime import datetime
from app.models import ConfigStatus, Environment


class ConfigCreate(BaseModel):
    key: str
    environment: Environment
    value: str


class ConfigUpdate(BaseModel):
    value: str


class ConfigResponse(BaseModel):
    id: int
    key: str
    environment: Environment
    version: int
    value: str
    status: ConfigStatus
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class WatchCreate(BaseModel):
    key: str
    callback_url: str
    environment: Environment


class WatchResponse(BaseModel):
    id: int
    key: str
    callback_url: str
    environment: Environment
    is_active: str
    failure_count: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class CallbackPayload(BaseModel):
    key: str
    environment: Environment
    value: str
    version: int


class DiffResponse(BaseModel):
    from_version: int
    to_version: int
    key: str
    environment: Environment
    from_value: str
    to_value: str
    changed: bool
