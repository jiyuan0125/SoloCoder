from datetime import datetime
from typing import Optional, List
from pydantic import BaseModel, Field


class BerthBase(BaseModel):
    name: str
    berth_type: str
    capacity: float
    is_under_maintenance: Optional[bool] = False


class BerthCreate(BerthBase):
    pass


class BerthUpdate(BaseModel):
    name: Optional[str] = None
    berth_type: Optional[str] = None
    capacity: Optional[float] = None
    is_under_maintenance: Optional[bool] = None


class BerthResponse(BerthBase):
    id: int
    created_at: datetime
    is_available: Optional[bool] = None

    class Config:
        from_attributes = True


class ShipBase(BaseModel):
    name: str
    imo_number: Optional[str] = None
    ship_type: str
    draught: float
    expected_arrival: datetime


class ShipCreate(ShipBase):
    pass


class ShipUpdate(BaseModel):
    name: Optional[str] = None
    imo_number: Optional[str] = None
    ship_type: Optional[str] = None
    draught: Optional[float] = None
    status: Optional[str] = None
    expected_arrival: Optional[datetime] = None


class ShipResponse(ShipBase):
    id: int
    status: str
    current_berth_id: Optional[int] = None
    created_at: datetime

    class Config:
        from_attributes = True


class HandlingOperationBase(BaseModel):
    ship_id: int
    priority: int = 1
    cargo_volume: float
    description: Optional[str] = None


class HandlingOperationCreate(HandlingOperationBase):
    berth_id: Optional[int] = None
    yard_area_id: Optional[int] = None


class HandlingOperationUpdate(BaseModel):
    status: Optional[str] = None
    berth_id: Optional[int] = None
    yard_area_id: Optional[int] = None
    priority: Optional[int] = None


class HandlingOperationResponse(HandlingOperationBase):
    id: int
    berth_id: Optional[int]
    yard_area_id: Optional[int]
    status: str
    started_at: Optional[datetime] = None
    completed_at: Optional[datetime] = None
    created_at: datetime

    class Config:
        from_attributes = True


class YardAreaBase(BaseModel):
    name: str
    total_capacity: float
    is_available: Optional[bool] = True


class YardAreaCreate(YardAreaBase):
    pass


class YardAreaUpdate(BaseModel):
    name: Optional[str] = None
    total_capacity: Optional[float] = None
    is_available: Optional[bool] = None


class YardAreaResponse(YardAreaBase):
    id: int
    created_at: datetime
    current_usage: Optional[float] = None
    remaining_capacity: Optional[float] = None

    class Config:
        from_attributes = True


class SchedulingRecommendation(BaseModel):
    recommended_berths: List[BerthResponse]
    waiting_queue: List[HandlingOperationResponse]
