from datetime import datetime
from typing import Optional

from pydantic import BaseModel, Field


class LighthouseBase(BaseModel):
    name: str = Field(..., max_length=100)
    location: str = Field(..., max_length=255)
    rated_illuminance: float
    battery_capacity: float
    solar_power: float
    last_overhaul_date: Optional[str] = None
    status: Optional[str] = "normal"
    operator: Optional[str] = None


class LighthouseCreate(LighthouseBase):
    pass


class LighthouseUpdate(BaseModel):
    name: Optional[str] = None
    location: Optional[str] = None
    rated_illuminance: Optional[float] = None
    battery_capacity: Optional[float] = None
    solar_power: Optional[float] = None
    last_overhaul_date: Optional[str] = None
    status: Optional[str] = None
    operator: Optional[str] = None


class LighthouseResponse(LighthouseBase):
    id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class LightRecordBase(BaseModel):
    lighthouse_id: int
    timestamp: datetime
    illuminance: float
    is_on: int = 1
    status: str = "normal"


class LightRecordCreate(LightRecordBase):
    pass


class LightRecordResponse(LightRecordBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class EnergyRecordBase(BaseModel):
    lighthouse_id: int
    timestamp: datetime
    solar_generation: float
    battery_level: float
    battery_percent: float


class EnergyRecordCreate(EnergyRecordBase):
    pass


class EnergyRecordResponse(EnergyRecordBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True
