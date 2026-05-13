from pydantic import BaseModel
from datetime import datetime
from typing import Optional


class ConfigCreate(BaseModel):
    value: str


class ConfigResponse(BaseModel):
    project: str
    env: str
    key: str
    value: Optional[str]
    version: int

    class Config:
        from_attributes = True


class ConfigVersionResponse(BaseModel):
    version: int
    value: Optional[str]
    operation: str
    created_at: datetime

    class Config:
        from_attributes = True


class WatchCreate(BaseModel):
    project: str
    env: str
    key: str
    callback_url: str


class WatchResponse(BaseModel):
    id: int
    project: str
    env: str
    key: str
    callback_url: str
    status: str
    failed_count: int

    class Config:
        from_attributes = True
