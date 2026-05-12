from datetime import datetime
from typing import Optional, List
from pydantic import BaseModel
from .models import StationType, RedTideStatus, WaveAlertLevel


class StationBase(BaseModel):
    name: str
    station_type: StationType
    latitude: Optional[float] = None
    longitude: Optional[float] = None
    description: Optional[str] = None


class StationCreate(StationBase):
    pass


class Station(StationBase):
    id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class TideReadingBase(BaseModel):
    value: float
    reading_time: datetime


class TideReadingCreate(TideReadingBase):
    pass


class TideReading(TideReadingBase):
    id: int
    station_id: int
    created_at: datetime

    class Config:
        from_attributes = True


class HourlyTideReading(BaseModel):
    hour: int
    value: float
    reading_time: datetime


class DailyTideStats(BaseModel):
    date: str
    max_value: float
    min_value: float
    avg_value: float


class WaveReadingBase(BaseModel):
    significant_wave_height: float
    reading_time: datetime


class WaveReadingCreate(WaveReadingBase):
    pass


class WaveReading(WaveReadingBase):
    id: int
    station_id: int
    alert_level: WaveAlertLevel
    created_at: datetime

    class Config:
        from_attributes = True


class WaterTempReadingBase(BaseModel):
    value: float
    reading_time: datetime


class WaterTempReadingCreate(WaterTempReadingBase):
    pass


class WaterTempReading(WaterTempReadingBase):
    id: int
    station_id: int
    created_at: datetime

    class Config:
        from_attributes = True


class SalinityReadingBase(BaseModel):
    value: float
    reading_time: datetime


class SalinityReadingCreate(SalinityReadingBase):
    pass


class SalinityReading(SalinityReadingBase):
    id: int
    station_id: int
    created_at: datetime

    class Config:
        from_attributes = True


class DissolvedOxygenReadingBase(BaseModel):
    value: float
    reading_time: datetime


class DissolvedOxygenReadingCreate(DissolvedOxygenReadingBase):
    pass


class DissolvedOxygenReading(DissolvedOxygenReadingBase):
    id: int
    station_id: int
    created_at: datetime

    class Config:
        from_attributes = True


class ChlorophyllReadingBase(BaseModel):
    value: float
    reading_time: datetime


class ChlorophyllReadingCreate(ChlorophyllReadingBase):
    pass


class ChlorophyllReading(ChlorophyllReadingBase):
    id: int
    station_id: int
    is_anomaly: bool
    created_at: datetime

    class Config:
        from_attributes = True


class RedTideEventBase(BaseModel):
    notes: Optional[str] = None


class RedTideEventCreate(RedTideEventBase):
    pass


class RedTideEventConfirm(BaseModel):
    confirmed_by: str
    notes: Optional[str] = None


class RedTideEvent(RedTideEventBase):
    id: int
    station_id: int
    status: RedTideStatus
    suspected_at: datetime
    confirmed_at: Optional[datetime] = None
    published_at: Optional[datetime] = None
    resolved_at: Optional[datetime] = None
    confirmed_by: Optional[str] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class RedTideReadingBase(BaseModel):
    chlorophyll_value: float
    do_value: Optional[float] = None
    reading_time: datetime
    is_over_threshold: bool = True


class RedTideReading(RedTideReadingBase):
    id: int
    event_id: int
    created_at: datetime

    class Config:
        from_attributes = True


class FarmerBase(BaseModel):
    name: str
    contact: Optional[str] = None
    email: Optional[str] = None
    phone: Optional[str] = None
    station_ids: Optional[str] = None


class FarmerCreate(FarmerBase):
    pass


class Farmer(FarmerBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class NotificationBase(BaseModel):
    message: str
    notification_type: str


class Notification(NotificationBase):
    id: int
    farmer_id: Optional[int] = None
    station_id: Optional[int] = None
    event_id: Optional[int] = None
    sent_at: datetime
    is_read: bool

    class Config:
        from_attributes = True


class TideResponse(BaseModel):
    hourly_readings: List[HourlyTideReading]
    daily_stats: List[DailyTideStats]


class WaveResponse(BaseModel):
    readings: List[WaveReading]
    latest_alert: Optional[WaveAlertLevel]


class RedTideResponse(BaseModel):
    current_status: RedTideStatus
    current_event: Optional[RedTideEvent]
    recent_events: List[RedTideEvent]
    recent_readings: List[ChlorophyllReading]


class StationData(BaseModel):
    station: Station
    latest_water_temp: Optional[WaterTempReading]
    salinity_status: Optional[str]
    latest_salinity: Optional[SalinityReading]
    latest_do: Optional[DissolvedOxygenReading]
    latest_chlorophyll: Optional[ChlorophyllReading]
