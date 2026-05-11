from datetime import datetime
from typing import Optional, List
from pydantic import BaseModel, Field
from .models import StationType, RedtideStatus, WaveAlertLevel


class StationBase(BaseModel):
    id: str
    name: str
    station_type: StationType
    location: Optional[str] = None
    sea_area: Optional[str] = None


class StationCreate(StationBase):
    pass


class Station(StationBase):
    created_at: datetime

    class Config:
        from_attributes = True


class TideMeasurementBase(BaseModel):
    measured_at: datetime
    tide_level: float = Field(..., description="潮位（米）")


class TideMeasurementCreate(TideMeasurementBase):
    pass


class TideMeasurement(TideMeasurementBase):
    id: int
    station_id: str

    class Config:
        from_attributes = True


class TideDailyStats(BaseModel):
    date: str
    max_tide: float
    min_tide: float
    avg_tide: float


class WaveMeasurementBase(BaseModel):
    measured_at: datetime
    significant_wave_height: float = Field(..., description="有效波高（米）")
    wave_period: Optional[float] = None
    wave_direction: Optional[float] = None


class WaveMeasurementCreate(WaveMeasurementBase):
    pass


class WaveMeasurement(WaveMeasurementBase):
    id: int
    station_id: str
    alert_level: WaveAlertLevel = WaveAlertLevel.NONE

    class Config:
        from_attributes = True


class WaterTempMeasurementBase(BaseModel):
    measured_at: datetime
    temperature: float = Field(..., description="水温（摄氏度）")


class WaterTempMeasurementCreate(WaterTempMeasurementBase):
    pass


class WaterTempMeasurement(WaterTempMeasurementBase):
    id: int
    station_id: str

    class Config:
        from_attributes = True


class SalinityMeasurementBase(BaseModel):
    measured_at: datetime
    salinity: float = Field(..., description="盐度（PSU）")


class SalinityMeasurementCreate(SalinityMeasurementBase):
    pass


class SalinityMeasurement(SalinityMeasurementBase):
    id: int
    station_id: str

    class Config:
        from_attributes = True


class DissolvedOxygenMeasurementBase(BaseModel):
    measured_at: datetime
    dissolved_oxygen: float = Field(..., description="溶解氧（mg/L）")


class DissolvedOxygenMeasurementCreate(DissolvedOxygenMeasurementBase):
    pass


class DissolvedOxygenMeasurement(DissolvedOxygenMeasurementBase):
    id: int
    station_id: str

    class Config:
        from_attributes = True


class ChlorophyllMeasurementBase(BaseModel):
    measured_at: datetime
    chlorophyll: float = Field(..., description="叶绿素浓度（μg/L）")


class ChlorophyllMeasurementCreate(ChlorophyllMeasurementBase):
    pass


class ChlorophyllMeasurement(ChlorophyllMeasurementBase):
    id: int
    station_id: str

    class Config:
        from_attributes = True


class RedtideEventBase(BaseModel):
    description: Optional[str] = None


class RedtideEventCreate(RedtideEventBase):
    pass


class RedtideEvent(RedtideEventBase):
    id: int
    station_id: str
    status: RedtideStatus
    suspected_at: datetime
    confirmed_at: Optional[datetime] = None
    published_at: Optional[datetime] = None
    resolved_at: Optional[datetime] = None
    severity_level: Optional[str] = None

    class Config:
        from_attributes = True


class RedtideEventUpdate(BaseModel):
    status: Optional[RedtideStatus] = None
    description: Optional[str] = None


class FarmerBase(BaseModel):
    name: str
    phone: Optional[str] = None
    email: Optional[str] = None
    sea_area: Optional[str] = None
    farm_name: Optional[str] = None


class FarmerCreate(FarmerBase):
    pass


class Farmer(FarmerBase):
    id: int
    is_active: bool
    created_at: datetime

    class Config:
        from_attributes = True


class Notification(BaseModel):
    id: int
    redtide_event_id: int
    farmer_id: Optional[int] = None
    message: str
    sent_at: datetime
    is_read: bool

    class Config:
        from_attributes = True


class StationDataInput(BaseModel):
    measured_at: datetime
    tide: Optional[float] = None
    wave_height: Optional[float] = None
    wave_period: Optional[float] = None
    wave_direction: Optional[float] = None
    water_temp: Optional[float] = None
    salinity: Optional[float] = None
    dissolved_oxygen: Optional[float] = None
    chlorophyll: Optional[float] = None
