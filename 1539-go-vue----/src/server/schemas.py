from datetime import datetime, date
from typing import Optional
from pydantic import BaseModel, Field
from src.core.models import (
    ZoneType,
    TimePeriod,
    AlertType,
    AlertLevel,
)


class ZoneLimitCreate(BaseModel):
    zone_type: ZoneType
    daytime_limit: float = Field(..., gt=0, lt=150)
    nighttime_limit: float = Field(..., gt=0, lt=150)


class ZoneLimitUpdate(BaseModel):
    zone_type: Optional[ZoneType] = None
    daytime_limit: Optional[float] = Field(None, gt=0, lt=150)
    nighttime_limit: Optional[float] = Field(None, gt=0, lt=150)


class ZoneLimitResponse(BaseModel):
    id: int
    zone_type: ZoneType
    daytime_limit: float
    nighttime_limit: float
    created_at: datetime
    updated_at: datetime
    
    class Config:
        from_attributes = True


class MonitoringPointCreate(BaseModel):
    name: str
    code: str
    zone_type: ZoneType
    address: Optional[str] = None
    latitude: Optional[float] = None
    longitude: Optional[float] = None
    is_active: bool = True
    last_calibration_date: date
    next_calibration_date: date


class MonitoringPointUpdate(BaseModel):
    name: Optional[str] = None
    code: Optional[str] = None
    zone_type: Optional[ZoneType] = None
    address: Optional[str] = None
    latitude: Optional[float] = None
    longitude: Optional[float] = None
    is_active: Optional[bool] = None
    last_calibration_date: Optional[date] = None
    next_calibration_date: Optional[date] = None
    calibration_due: Optional[bool] = None


class MonitoringPointResponse(BaseModel):
    id: int
    name: str
    code: str
    zone_type: ZoneType
    address: Optional[str] = None
    latitude: Optional[float] = None
    longitude: Optional[float] = None
    is_active: bool
    last_calibration_date: date
    next_calibration_date: date
    calibration_due: bool
    created_at: datetime
    updated_at: datetime
    
    class Config:
        from_attributes = True


class NoiseDataCreate(BaseModel):
    point_id: int
    value: float = Field(..., ge=0, le=200)
    timestamp: Optional[datetime] = None


class NoiseDataResponse(BaseModel):
    id: int
    point_id: int
    timestamp: datetime
    value: float
    is_valid: bool
    is_exceeded: bool
    time_period: TimePeriod
    created_at: datetime
    updated_at: datetime
    
    class Config:
        from_attributes = True


class AlertEventResponse(BaseModel):
    id: int
    point_id: int
    alert_type: AlertType
    level: AlertLevel
    message: str
    is_resolved: bool
    resolved_at: Optional[datetime]
    start_time: datetime
    end_time: Optional[datetime]
    created_at: datetime
    updated_at: datetime
    
    class Config:
        from_attributes = True


class ConstructionPermitCreate(BaseModel):
    point_ids: list[int]
    permit_number: str
    start_date: date
    end_date: date
    description: Optional[str] = None
    is_active: bool = True


class ConstructionPermitUpdate(BaseModel):
    point_ids: Optional[list[int]] = None
    permit_number: Optional[str] = None
    start_date: Optional[date] = None
    end_date: Optional[date] = None
    description: Optional[str] = None
    is_active: Optional[bool] = None


class ConstructionPermitResponse(BaseModel):
    id: int
    point_ids: list[int]
    permit_number: str
    start_date: date
    end_date: date
    description: Optional[str]
    is_active: bool
    created_at: datetime
    updated_at: datetime
    
    class Config:
        from_attributes = True


class DailyStatisticsResponse(BaseModel):
    id: int
    point_id: int
    date: date
    daytime_avg: Optional[float]
    nighttime_avg: Optional[float]
    overall_avg: Optional[float]
    daytime_compliance_rate: Optional[float]
    nighttime_compliance_rate: Optional[float]
    overall_compliance_rate: Optional[float]
    data_missing_hours: float
    is_valid_for_compliance: bool
    created_at: datetime
    updated_at: datetime
    
    class Config:
        from_attributes = True


class MessageResponse(BaseModel):
    message: str
