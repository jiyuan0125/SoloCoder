from datetime import datetime
from typing import Optional
from enum import Enum
from pydantic import Field
from .base import BaseModelWithID


class ZoneType(str, Enum):
    RESIDENTIAL = "residential"
    COMMERCIAL = "commercial"
    INDUSTRIAL = "industrial"
    ROADSIDE = "roadside"
    MIXED = "mixed"


class ZoneLimit(BaseModelWithID):
    zone_type: ZoneType
    daytime_limit: float = Field(..., gt=0, lt=150)
    nighttime_limit: float = Field(..., gt=0, lt=150)


class MonitoringPoint(BaseModelWithID):
    name: str
    code: str
    zone_type: ZoneType
    address: Optional[str] = None
    latitude: Optional[float] = None
    longitude: Optional[float] = None
    is_active: bool = True
    last_calibration_date: datetime
    next_calibration_date: datetime
    calibration_due: bool = False
