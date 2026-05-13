from enum import Enum
from pydantic import BaseModel, Field
from typing import Literal


class ConnectionStatus(str, Enum):
    IDLE = "idle"
    IN_USE = "in_use"
    VALIDATING = "validating"
    DISCONNECTED = "disconnected"


class DatasourceType(str, Enum):
    HTTP = "http"
    TCP = "tcp"


class DatasourceCreate(BaseModel):
    name: str
    type: DatasourceType
    max_connections: int = Field(gt=0)
    idle_timeout: int = Field(gt=0)


class ConnectionStats(BaseModel):
    idle: int
    in_use: int
    validating: int
    disconnected: int


class DatasourceStats(BaseModel):
    name: str
    type: DatasourceType
    max_connections: int
    current_pool_size: int
    connections: ConnectionStats


class DatasourceOverview(BaseModel):
    name: str
    type: DatasourceType
    max_connections: int
    current_pool_size: int
    idle_connections: int
    in_use_connections: int
