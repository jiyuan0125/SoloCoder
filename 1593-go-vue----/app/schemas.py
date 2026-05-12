from datetime import datetime, date, time
from typing import Optional, List
from pydantic import BaseModel, Field


class CollectionBase(BaseModel):
    name: str
    description: Optional[str] = None
    level: str = "一般"
    location: Optional[str] = None
    operator: Optional[str] = None


class CollectionCreate(CollectionBase):
    pass


class CollectionUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    level: Optional[str] = None
    location: Optional[str] = None
    operator: Optional[str] = None


class CollectionResponse(BaseModel):
    id: int
    name: str
    description: Optional[str] = None
    level: str
    status: str
    location: Optional[str] = None
    operator: Optional[str] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class InventoryLogBase(BaseModel):
    operation_type: str
    operator: str
    notes: Optional[str] = None


class InventoryLogResponse(BaseModel):
    id: int
    collection_id: int
    operation_type: str
    operator: str
    notes: Optional[str] = None
    created_at: datetime

    class Config:
        from_attributes = True


class CollectionDetailResponse(CollectionResponse):
    inventory_logs: List[InventoryLogResponse] = []


class ExhibitionItemResponse(BaseModel):
    id: int
    collection_id: int
    collection_name: str
    added_at: datetime

    class Config:
        from_attributes = True


class ExhibitionBase(BaseModel):
    name: str
    description: Optional[str] = None
    location: Optional[str] = None
    start_date: Optional[date] = None
    end_date: Optional[date] = None


class ExhibitionCreate(ExhibitionBase):
    pass


class ExhibitionUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    location: Optional[str] = None
    start_date: Optional[date] = None
    end_date: Optional[date] = None


class ExhibitionResponse(BaseModel):
    id: int
    name: str
    description: Optional[str] = None
    location: Optional[str] = None
    status: str
    start_date: Optional[date] = None
    end_date: Optional[date] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class ExhibitionDetailResponse(ExhibitionResponse):
    items: List[ExhibitionItemResponse] = []


class AddCollectionsToExhibition(BaseModel):
    collection_ids: List[int]
    operator: str


class RemoveCollectionFromExhibition(BaseModel):
    collection_id: int
    operator: str


class VisitReservationBase(BaseModel):
    visitor_name: str
    id_number: str
    phone: Optional[str] = None
    visit_date: date
    time_slot: str
    visitor_count: int = 1


class VisitReservationCreate(VisitReservationBase):
    pass


class VisitReservationResponse(BaseModel):
    id: int
    visitor_name: str
    id_number: str
    phone: Optional[str] = None
    visit_date: date
    time_slot: str
    visitor_count: int
    status: str
    created_at: datetime
    cancelled_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class CancelReservation(BaseModel):
    id_number: str


class TimeSlotConfigBase(BaseModel):
    time_slot: str
    max_visitors: int


class TimeSlotConfigCreate(TimeSlotConfigBase):
    pass


class TimeSlotConfigResponse(BaseModel):
    id: int
    time_slot: str
    max_visitors: int
    is_active: bool

    class Config:
        from_attributes = True


class MeetingReservationBase(BaseModel):
    title: str
    organizer: str
    meeting_date: date
    start_time: time
    end_time: time
    room: str
    attendees: int = 1
    notes: Optional[str] = None


class MeetingReservationCreate(MeetingReservationBase):
    pass


class MeetingReservationResponse(BaseModel):
    id: int
    title: str
    organizer: str
    meeting_date: date
    start_time: time
    end_time: time
    room: str
    attendees: int
    notes: Optional[str] = None
    status: str
    created_at: datetime

    class Config:
        from_attributes = True


class BorrowingBase(BaseModel):
    collection_id: int
    borrower: str
    contact: Optional[str] = None
    borrow_date: date
    expected_return_date: date
    purpose: Optional[str] = None
    operator: str


class BorrowingCreate(BorrowingBase):
    pass


class ReturnBorrowing(BaseModel):
    actual_return_date: Optional[date] = None
    operator: str


class BorrowingResponse(BaseModel):
    id: int
    collection_id: int
    collection_name: Optional[str] = None
    borrower: str
    contact: Optional[str] = None
    borrow_date: date
    expected_return_date: date
    actual_return_date: Optional[date] = None
    purpose: Optional[str] = None
    status: str
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class ReminderTodoResponse(BaseModel):
    id: int
    borrowing_id: int
    borrower: Optional[str] = None
    collection_name: Optional[str] = None
    overdue_days: int
    message: str
    is_completed: bool
    created_at: datetime
    completed_at: Optional[datetime] = None

    class Config:
        from_attributes = True
