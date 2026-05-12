from datetime import datetime
from typing import Optional
from pydantic import BaseModel

from app.models import (
    DeviceStatus, AlarmLevel, AlarmStatus, ControlMode,
    LightingGroup, MaintenanceTaskStatus
)


class VentilationDeviceBase(BaseModel):
    name: str
    co_threshold: float = 100.0
    visibility_threshold: float = 10.0


class VentilationDeviceCreate(VentilationDeviceBase):
    pass


class VentilationDeviceResponse(VentilationDeviceBase):
    id: int
    status: str
    mode: str
    fault: bool
    co_level: float
    visibility: float
    is_running: bool
    start_time: Optional[datetime] = None
    shutdown_delay_active: bool
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class VentilationModeUpdate(BaseModel):
    mode: ControlMode


class LightingDeviceBase(BaseModel):
    name: str
    group: LightingGroup = LightingGroup.MIDDLE


class LightingDeviceCreate(LightingDeviceBase):
    pass


class LightingDeviceResponse(LightingDeviceBase):
    id: int
    status: str
    mode: str
    fault: bool
    brightness: int
    target_brightness: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class LightingBrightnessUpdate(BaseModel):
    brightness: int


class LightingGroupUpdate(BaseModel):
    group: LightingGroup


class DrainageDeviceBase(BaseModel):
    name: str
    model: str
    parameters: Optional[str] = None
    warning_level: float = 5.0
    safe_level: float = 2.0
    maintenance_interval_hours: float = 500.0


class DrainageDeviceCreate(DrainageDeviceBase):
    pass


class DrainageDeviceResponse(DrainageDeviceBase):
    id: int
    status: str
    mode: str
    fault: bool
    water_level: float
    is_running: bool
    start_time: Optional[datetime] = None
    shutdown_delay_active: bool
    cumulative_runtime: float
    last_maintenance_runtime: float
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class DrainageLevelUpdate(BaseModel):
    level: float


class AlarmBase(BaseModel):
    subsystem: str
    device_id: int
    device_name: str
    level: AlarmLevel
    message: str


class AlarmCreate(AlarmBase):
    pass


class AlarmResponse(AlarmBase):
    id: int
    status: str
    created_at: datetime
    acknowledged_at: Optional[datetime] = None
    resolved_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class MaintenanceTaskResponse(BaseModel):
    id: int
    drainage_device_id: int
    device_name: str
    device_model: str
    status: str
    content: str
    created_at: datetime
    completed_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class DeviceOperationResponse(BaseModel):
    success: bool
    message: str
