from pydantic import BaseModel
from datetime import datetime
from typing import Optional, List


class SwitchBase(BaseModel):
    device_id: str
    name: str


class SwitchCreate(SwitchBase):
    pass


class SwitchResponse(BaseModel):
    device_id: str
    name: str
    position: str
    locked: bool
    has_indication: bool
    occupied: bool
    status: str
    last_updated: datetime

    class Config:
        from_attributes = True


class InterlockingBase(BaseModel):
    device_id: str
    name: str


class InterlockingCreate(InterlockingBase):
    pass


class InterlockingResponse(BaseModel):
    device_id: str
    name: str
    status: str
    current_route: Optional[str]
    route_status: str
    last_updated: datetime

    class Config:
        from_attributes = True


class BlockSectionBase(BaseModel):
    device_id: str
    name: str


class BlockSectionCreate(BlockSectionBase):
    pass


class BlockSectionResponse(BaseModel):
    device_id: str
    name: str
    occupied: bool
    mode: str
    circuit_ok: bool
    last_updated: datetime

    class Config:
        from_attributes = True


class SemaphoreBase(BaseModel):
    device_id: str
    name: str


class SemaphoreCreate(SemaphoreBase):
    pass


class SemaphoreResponse(BaseModel):
    device_id: str
    name: str
    status: str
    fault: bool
    last_updated: datetime

    class Config:
        from_attributes = True


class AlertResponse(BaseModel):
    id: int
    device_type: str
    device_id: str
    message: str
    timestamp: datetime
    resolved: bool

    class Config:
        from_attributes = True


class StatsResponse(BaseModel):
    period: str
    switch_operations: int
    switch_faults: int
    block_occupancy_rate: float


class RouteRequest(BaseModel):
    route_id: str
    switches: List[str]
    blocks: List[str]
    semaphores: List[str]
    conflicting_routes: List[str] = []
