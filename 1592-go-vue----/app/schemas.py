from pydantic import BaseModel, Field
from typing import Optional, List
from datetime import datetime, date, time


class VenueBase(BaseModel):
    name: str
    location: Optional[str] = None
    description: Optional[str] = None


class VenueCreate(VenueBase):
    pass


class VenueUpdate(BaseModel):
    name: Optional[str] = None
    location: Optional[str] = None
    description: Optional[str] = None


class VenueResponse(VenueBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class HallBase(BaseModel):
    name: str
    floor: Optional[int] = None
    total_area: Optional[float] = None
    max_exhibitors_per_slot: int = 5
    description: Optional[str] = None


class HallCreate(HallBase):
    venue_id: int


class HallUpdate(BaseModel):
    name: Optional[str] = None
    floor: Optional[int] = None
    total_area: Optional[float] = None
    max_exhibitors_per_slot: Optional[int] = None
    description: Optional[str] = None


class HallResponse(HallBase):
    id: int
    venue_id: int

    class Config:
        from_attributes = True


class BoothBase(BaseModel):
    booth_number: str
    area: float
    is_special: bool = False
    base_price: float
    position: Optional[str] = None
    description: Optional[str] = None


class BoothCreate(BoothBase):
    hall_id: int


class BoothUpdate(BaseModel):
    booth_number: Optional[str] = None
    area: Optional[float] = None
    is_special: Optional[bool] = None
    base_price: Optional[float] = None
    position: Optional[str] = None
    status: Optional[str] = None
    description: Optional[str] = None


class BoothResponse(BoothBase):
    id: int
    hall_id: int
    status: str

    class Config:
        from_attributes = True


class BoothPriceUpdate(BaseModel):
    new_price: float


class ExhibitionBase(BaseModel):
    name: str
    organizer: Optional[str] = None
    start_date: date
    end_date: date
    description: Optional[str] = None


class ExhibitionCreate(ExhibitionBase):
    hall_ids: List[int]


class ExhibitionUpdate(BaseModel):
    name: Optional[str] = None
    organizer: Optional[str] = None
    start_date: Optional[date] = None
    end_date: Optional[date] = None
    description: Optional[str] = None


class ExhibitionResponse(ExhibitionBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class BoothSelectionBase(BaseModel):
    exhibitor_name: str
    contact_person: Optional[str] = None
    contact_phone: Optional[str] = None
    notes: Optional[str] = None


class BoothSelectionCreate(BoothSelectionBase):
    booth_id: int
    exhibition_id: int


class BoothSelectionConfirm(BaseModel):
    payment_confirm: bool = True


class BoothSelectionResponse(BaseModel):
    id: int
    booth_id: int
    exhibition_id: int
    exhibitor_name: str
    contact_person: Optional[str]
    contact_phone: Optional[str]
    selected_at: datetime
    expires_at: Optional[datetime]
    status: str
    final_price: Optional[float]
    is_paid: bool
    paid_at: Optional[datetime]
    notes: Optional[str]

    class Config:
        from_attributes = True


class SetupScheduleBase(BaseModel):
    setup_date: date
    start_time: time
    end_time: time
    notes: Optional[str] = None


class SetupScheduleCreate(SetupScheduleBase):
    selection_id: int


class SetupScheduleResponse(SetupScheduleBase):
    id: int
    hall_id: int
    selection_id: int
    actual_hours: Optional[float]
    billed_hours: Optional[int]
    created_at: datetime

    class Config:
        from_attributes = True


class VisitReservationBase(BaseModel):
    visitor_name: str
    visitor_phone: Optional[str] = None
    visitor_email: Optional[str] = None
    visit_date: date
    party_size: int = 1


class VisitReservationCreate(VisitReservationBase):
    exhibition_id: int


class VisitReservationResponse(BaseModel):
    id: int
    exhibition_id: int
    visitor_name: str
    visitor_phone: Optional[str]
    visitor_email: Optional[str]
    visit_date: date
    party_size: int
    status: str
    queue_position: Optional[int]
    created_at: datetime

    class Config:
        from_attributes = True


class DailyCapacityBase(BaseModel):
    date: date
    max_visitors: int


class DailyCapacityCreate(DailyCapacityBase):
    exhibition_id: int


class DailyCapacityResponse(DailyCapacityBase):
    id: int
    exhibition_id: int
    current_confirmed: int

    class Config:
        from_attributes = True
