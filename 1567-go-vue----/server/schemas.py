from datetime import datetime
from typing import Optional, List
from pydantic import BaseModel, Field


class FlightBase(BaseModel):
    flight_number: str
    departure: str
    destination: str
    scheduled_departure: Optional[datetime] = None
    scheduled_arrival: Optional[datetime] = None
    max_load_weight: float = 2000.0


class FlightCreate(FlightBase):
    id: str


class FlightUpdate(BaseModel):
    flight_number: Optional[str] = None
    departure: Optional[str] = None
    destination: Optional[str] = None
    scheduled_departure: Optional[datetime] = None
    scheduled_arrival: Optional[datetime] = None
    max_load_weight: Optional[float] = None


class FlightResponse(FlightBase):
    id: str
    actual_load_weight: float = 0.0

    class Config:
        from_attributes = True


class BaggageBase(BaseModel):
    tag_number: str
    weight: float
    is_valuable: bool = False
    passenger_name: str
    flight_id: Optional[str] = None


class BaggageCreate(BaggageBase):
    id: str


class BaggageUpdate(BaseModel):
    weight: Optional[float] = None
    is_valuable: Optional[bool] = None
    passenger_name: Optional[str] = None
    flight_id: Optional[str] = None


class BaggageResponse(BaggageBase):
    id: str
    current_status: str
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class BaggageEventResponse(BaseModel):
    id: int
    baggage_id: str
    event_type: str
    event_time: datetime
    location: Optional[str] = None
    notes: Optional[str] = None
    success: bool

    class Config:
        from_attributes = True


class SortingRecordResponse(BaseModel):
    id: int
    baggage_id: str
    flight_id: Optional[str] = None
    sorting_time: datetime
    success: bool
    failure_reason: Optional[str] = None
    lane: Optional[str] = None

    class Config:
        from_attributes = True


class LoadingRecordResponse(BaseModel):
    id: int
    baggage_id: str
    flight_id: str
    loading_time: datetime
    position: str
    is_special_position: bool

    class Config:
        from_attributes = True


class ConveyorBeltRecordResponse(BaseModel):
    id: int
    baggage_id: str
    conveyor_belt_number: str
    placed_at: datetime
    picked_at: Optional[datetime] = None
    is_delayed: bool
    is_stuck: bool

    class Config:
        from_attributes = True


class AlertResponse(BaseModel):
    id: int
    alert_type: str
    flight_id: Optional[str] = None
    baggage_id: Optional[str] = None
    alert_time: datetime
    message: str
    resolved: bool
    resolved_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class ClaimResponse(BaseModel):
    id: int
    baggage_id: str
    claim_amount: float
    max_allowed: float
    status: str
    filed_at: datetime
    resolved_at: Optional[datetime] = None
    notes: Optional[str] = None

    class Config:
        from_attributes = True


class SecurityCheckRequest(BaseModel):
    success: bool
    notes: Optional[str] = None


class SortingRequest(BaseModel):
    success: bool
    lane: Optional[str] = None
    failure_reason: Optional[str] = None


class LoadingRequest(BaseModel):
    position: str


class ConveyorBeltRequest(BaseModel):
    conveyor_belt_number: str


class ClaimCreate(BaseModel):
    claim_amount: float
    notes: Optional[str] = None


class FlightBaggageListResponse(BaseModel):
    flight_id: str
    total_count: int
    baggage: List[BaggageResponse] = []


class FlightSortingResponse(BaseModel):
    flight_id: str
    total_count: int
    success_count: int
    failure_count: int
    records: List[SortingRecordResponse] = []


class FlightLoadingResponse(BaseModel):
    flight_id: str
    total_count: int
    total_weight: float
    max_load_weight: float
    special_position_count: int
    records: List[LoadingRecordResponse] = []
