from datetime import datetime, date
from typing import Optional
from enum import Enum
from pydantic import Field
from .base import BaseModelWithID


class TimePeriod(str, Enum):
    DAYTIME = "daytime"
    NIGHTTIME = "nighttime"


class NoiseData(BaseModelWithID):
    point_id: int
    timestamp: datetime
    value: float = Field(..., ge=0, le=200)
    is_valid: bool = True
    is_exceeded: bool = False
    time_period: TimePeriod


class AlertType(str, Enum):
    EXCEED = "exceed"
    CALIBRATION = "calibration"
    DEVICE_FAULT = "device_fault"


class AlertLevel(str, Enum):
    INFO = "info"
    WARNING = "warning"
    CRITICAL = "critical"


class AlertEvent(BaseModelWithID):
    point_id: int
    alert_type: AlertType
    level: AlertLevel
    message: str
    is_resolved: bool = False
    resolved_at: Optional[datetime] = None
    start_time: datetime
    end_time: Optional[datetime] = None


class ConstructionPermit(BaseModelWithID):
    point_ids: list[int]
    permit_number: str
    start_date: date
    end_date: date
    description: Optional[str] = None
    is_active: bool = True


class DailyStatistics(BaseModelWithID):
    point_id: int
    date: date
    daytime_avg: Optional[float] = None
    nighttime_avg: Optional[float] = None
    overall_avg: Optional[float] = None
    daytime_compliance_rate: Optional[float] = None
    nighttime_compliance_rate: Optional[float] = None
    overall_compliance_rate: Optional[float] = None
    data_missing_hours: float = 0.0
    is_valid_for_compliance: bool = True
