from pydantic import BaseModel, Field
from datetime import datetime
from typing import Optional, Dict, List
from server.models import StationType, AlertLevel, DispatchStatus, FloodPhaseType


class WatershedBase(BaseModel):
    name: str
    description: Optional[str] = None
    area: Optional[float] = None

class WatershedCreate(WatershedBase):
    pass

class WatershedUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    area: Optional[float] = None

class WatershedResponse(WatershedBase):
    id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class ReservoirBase(BaseModel):
    watershed_id: int
    name: str
    code: str
    capacity: Optional[float] = None
    flood_limit_level: Optional[float] = None
    warning_level: Optional[float] = None
    capacity_curve: Optional[Dict[str, float]] = None
    downstream_safe_discharge: Optional[float] = None

class ReservoirCreate(ReservoirBase):
    pass

class ReservoirUpdate(BaseModel):
    watershed_id: Optional[int] = None
    name: Optional[str] = None
    code: Optional[str] = None
    capacity: Optional[float] = None
    flood_limit_level: Optional[float] = None
    warning_level: Optional[float] = None
    capacity_curve: Optional[Dict[str, float]] = None
    downstream_safe_discharge: Optional[float] = None
    current_level: Optional[float] = None
    current_discharge: Optional[float] = None
    is_discharging: Optional[bool] = None

class ReservoirResponse(ReservoirBase):
    id: int
    current_level: float
    current_discharge: float
    is_discharging: bool
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class MonitoringStationBase(BaseModel):
    watershed_id: int
    name: str
    code: str
    station_type: str
    latitude: Optional[float] = None
    longitude: Optional[float] = None
    warning_threshold: Optional[float] = None

class MonitoringStationCreate(MonitoringStationBase):
    pass

class MonitoringStationUpdate(BaseModel):
    watershed_id: Optional[int] = None
    name: Optional[str] = None
    code: Optional[str] = None
    station_type: Optional[str] = None
    latitude: Optional[float] = None
    longitude: Optional[float] = None
    warning_threshold: Optional[float] = None
    is_active: Optional[bool] = None

class MonitoringStationResponse(MonitoringStationBase):
    id: int
    is_active: bool
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class HydrologicalDataBase(BaseModel):
    station_id: int
    value: float
    rainfall_6h: Optional[float] = None
    recorded_at: Optional[datetime] = None

class HydrologicalDataCreate(HydrologicalDataBase):
    pass

class HydrologicalDataResponse(HydrologicalDataBase):
    id: int
    created_at: datetime
    flood_event_id: Optional[int]

    class Config:
        from_attributes = True


class AlertResponse(BaseModel):
    id: int
    station_id: Optional[int]
    alert_level: str
    alert_message: str
    trigger_value: float
    threshold: Optional[float]
    triggered_at: datetime
    flood_event_id: Optional[int]
    resolved_at: Optional[datetime]
    is_active: bool

    class Config:
        from_attributes = True


class DispatchRecommendationBase(BaseModel):
    reservoir_id: int
    recommended_discharge: float
    rationale: Optional[str] = None

class DispatchRecommendationCreate(DispatchRecommendationBase):
    pass

class DispatchRecommendationResponse(DispatchRecommendationBase):
    id: int
    flood_event_id: Optional[int]
    current_capacity: Optional[float]
    downstream_safe: Optional[float]
    status: str
    created_by: str
    approved_by: Optional[str]
    created_at: datetime
    approved_at: Optional[datetime]
    executed_at: Optional[datetime]

    class Config:
        from_attributes = True


class DispatchApprove(BaseModel):
    approved_by: str = "operator"

class DispatchReject(BaseModel):
    reason: str = ""


class ReservoirOperationResponse(BaseModel):
    id: int
    reservoir_id: int
    flood_event_id: Optional[int]
    dispatch_id: Optional[int]
    previous_discharge: float
    new_discharge: float
    previous_level: float
    reason: Optional[str]
    operated_by: str
    operated_at: datetime

    class Config:
        from_attributes = True


class FloodPhaseResponse(BaseModel):
    id: int
    flood_event_id: int
    phase_type: str
    description: Optional[str]
    started_at: datetime
    ended_at: Optional[datetime]

    class Config:
        from_attributes = True


class FloodEventResponse(BaseModel):
    id: int
    watershed_id: Optional[int]
    title: str
    description: Optional[str]
    trigger_alert_id: Optional[int]
    started_at: datetime
    ended_at: Optional[datetime]
    is_active: bool
    summary: Optional[str]
    created_at: datetime

    class Config:
        from_attributes = True


class FloodEventDetail(FloodEventResponse):
    phases: List[FloodPhaseResponse] = []
    dispatches: List[DispatchRecommendationResponse] = []
    hydrological_data: List[HydrologicalDataResponse] = []
