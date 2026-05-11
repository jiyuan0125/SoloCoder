from datetime import datetime
from enum import Enum
from typing import Optional
from pydantic import BaseModel, Field


class ShiftType(str, Enum):
    MORNING = "morning"
    AFTERNOON = "afternoon"
    NIGHT = "night"


class AlarmLevel(int, Enum):
    LEVEL1 = 1
    LEVEL2 = 2
    LEVEL3 = 3


class MetricType(str, Enum):
    METHANE = "methane"
    CO = "co"
    WIND_SPEED = "wind_speed"
    TEMPERATURE = "temperature"
    DUST = "dust"


class Mine(BaseModel):
    id: Optional[int] = None
    name: str
    description: Optional[str] = None
    created_at: Optional[datetime] = None


class MonitorZone(BaseModel):
    id: Optional[int] = None
    mine_id: int
    name: str
    description: Optional[str] = None
    created_at: Optional[datetime] = None


class ThresholdConfig(BaseModel):
    id: Optional[int] = None
    zone_id: int
    metric_type: MetricType
    level1_threshold: float
    level2_threshold: float
    level3_threshold: float
    created_at: Optional[datetime] = None


class SensorReading(BaseModel):
    id: Optional[int] = None
    zone_id: int
    timestamp: datetime
    methane: float = Field(ge=0)
    co: float = Field(ge=0)
    wind_speed: float = Field(ge=0)
    temperature: float
    dust: float = Field(ge=0)
    shift_type: Optional[ShiftType] = None


class AlarmEvent(BaseModel):
    id: Optional[int] = None
    zone_id: int
    metric_type: MetricType
    alarm_level: AlarmLevel
    value: float
    threshold: float
    start_time: datetime
    end_time: Optional[datetime] = None
    acknowledged: bool = False
    acknowledged_at: Optional[datetime] = None
    acknowledged_by: Optional[str] = None


class ShiftStats(BaseModel):
    id: Optional[int] = None
    zone_id: int
    shift_date: str
    shift_type: ShiftType
    methane_avg: Optional[float] = None
    methane_max: Optional[float] = None
    methane_over_count: int = 0
    co_avg: Optional[float] = None
    co_max: Optional[float] = None
    co_over_count: int = 0
    wind_speed_avg: Optional[float] = None
    wind_speed_max: Optional[float] = None
    wind_speed_over_count: int = 0
    temperature_avg: Optional[float] = None
    temperature_max: Optional[float] = None
    temperature_over_count: int = 0
    dust_avg: Optional[float] = None
    dust_max: Optional[float] = None
    dust_over_count: int = 0
    reading_count: int = 0
    created_at: Optional[datetime] = None


class DailyReport(BaseModel):
    id: Optional[int] = None
    zone_id: int
    report_date: str
    methane_avg: Optional[float] = None
    methane_max: Optional[float] = None
    methane_over_count: int = 0
    co_avg: Optional[float] = None
    co_max: Optional[float] = None
    co_over_count: int = 0
    wind_speed_avg: Optional[float] = None
    wind_speed_max: Optional[float] = None
    wind_speed_over_count: int = 0
    temperature_avg: Optional[float] = None
    temperature_max: Optional[float] = None
    temperature_over_count: int = 0
    dust_avg: Optional[float] = None
    dust_max: Optional[float] = None
    dust_over_count: int = 0
    alarm_count: int = 0
    reading_count: int = 0
    created_at: Optional[datetime] = None
