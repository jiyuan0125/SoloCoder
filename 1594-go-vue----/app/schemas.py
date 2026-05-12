from datetime import datetime, date, time, timedelta
from typing import Optional, List
from pydantic import BaseModel, Field
from decimal import Decimal
from app.models import (
    ReaderType,
    CopyStatus,
    ReservationStatus,
    SeatReservationStatus,
    EventRegistrationStatus,
)


class BookBase(BaseModel):
    isbn: str
    title: str
    author: str
    publisher: Optional[str] = None
    publish_date: Optional[date] = None
    category: Optional[str] = None
    price: Decimal
    description: Optional[str] = None


class BookCreate(BookBase):
    pass


class BookUpdate(BaseModel):
    isbn: Optional[str] = None
    title: Optional[str] = None
    author: Optional[str] = None
    publisher: Optional[str] = None
    publish_date: Optional[date] = None
    category: Optional[str] = None
    price: Optional[Decimal] = None
    description: Optional[str] = None


class BookCopyBase(BaseModel):
    copy_number: str
    location: Optional[str] = None
    condition: Optional[str] = None


class BookCopyCreate(BookCopyBase):
    book_id: int


class BookCopyUpdate(BaseModel):
    copy_number: Optional[str] = None
    status: Optional[CopyStatus] = None
    location: Optional[str] = None
    condition: Optional[str] = None


class ReaderBase(BaseModel):
    card_number: str
    name: str
    reader_type: ReaderType
    phone: Optional[str] = None
    email: Optional[str] = None
    address: Optional[str] = None


class ReaderCreate(ReaderBase):
    pass


class ReaderUpdate(BaseModel):
    card_number: Optional[str] = None
    name: Optional[str] = None
    reader_type: Optional[ReaderType] = None
    phone: Optional[str] = None
    email: Optional[str] = None
    address: Optional[str] = None
    is_active: Optional[bool] = None


class BorrowingCreate(BaseModel):
    reader_id: int
    copy_id: int


class BorrowingReturn(BaseModel):
    borrowing_id: int


class BorrowingRenew(BaseModel):
    borrowing_id: int


class BookReservationCreate(BaseModel):
    reader_id: int
    book_id: int


class SeatBase(BaseModel):
    seat_number: str
    floor: Optional[str] = None
    area: Optional[str] = None
    has_power: bool = False
    notes: Optional[str] = None


class SeatCreate(SeatBase):
    pass


class SeatUpdate(BaseModel):
    seat_number: Optional[str] = None
    floor: Optional[str] = None
    area: Optional[str] = None
    is_active: Optional[bool] = None
    has_power: Optional[bool] = None
    notes: Optional[str] = None


class SeatReservationCreate(BaseModel):
    reader_id: int
    seat_id: int
    reservation_date: date
    start_time: time
    end_time: time


class SeatCheckIn(BaseModel):
    reservation_id: int


class EventBase(BaseModel):
    title: str
    description: Optional[str] = None
    event_date: date
    start_time: time
    end_time: time
    location: Optional[str] = None
    max_participants: int
    registration_start: Optional[datetime] = None
    registration_end: Optional[datetime] = None


class EventCreate(EventBase):
    pass


class EventUpdate(BaseModel):
    title: Optional[str] = None
    description: Optional[str] = None
    event_date: Optional[date] = None
    start_time: Optional[time] = None
    end_time: Optional[time] = None
    location: Optional[str] = None
    max_participants: Optional[int] = None
    registration_start: Optional[datetime] = None
    registration_end: Optional[datetime] = None
    is_active: Optional[bool] = None


class EventRegistrationCreate(BaseModel):
    event_id: int
    reader_id: int


class EventRegistrationCancel(BaseModel):
    registration_id: int


class BookCopy(BookCopyBase):
    id: int
    book_id: int
    status: CopyStatus
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class Book(BookBase):
    id: int
    created_at: datetime
    updated_at: datetime
    copies: List[BookCopy] = []

    class Config:
        from_attributes = True


class Reader(ReaderBase):
    id: int
    is_active: bool
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class Borrowing(BaseModel):
    id: int
    reader_id: int
    copy_id: int
    borrow_date: date
    due_date: date
    return_date: Optional[date] = None
    renew_count: int
    max_renew_count: int
    late_fee: Decimal
    is_returned: bool
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class BookReservation(BaseModel):
    id: int
    reader_id: int
    book_id: int
    copy_id: Optional[int] = None
    status: ReservationStatus
    queue_position: int
    reserved_at: datetime
    fulfilled_at: Optional[datetime] = None
    expires_at: Optional[datetime] = None
    notification_sent: bool
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class Seat(SeatBase):
    id: int
    is_active: bool
    created_at: datetime

    class Config:
        from_attributes = True


class SeatReservation(BaseModel):
    id: int
    reader_id: int
    seat_id: int
    reservation_date: date
    start_time: time
    end_time: time
    status: SeatReservationStatus
    check_in_time: Optional[datetime] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class Event(EventBase):
    id: int
    current_participants: int
    is_active: bool
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class EventRegistration(BaseModel):
    id: int
    event_id: int
    reader_id: int
    status: EventRegistrationStatus
    waitlist_position: Optional[int] = None
    registered_at: datetime
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class Notification(BaseModel):
    id: int
    reader_id: int
    title: str
    message: str
    is_read: bool
    notification_type: Optional[str] = None
    created_at: datetime

    class Config:
        from_attributes = True


class LateFeeInfo(BaseModel):
    overdue_days: int
    daily_rate: Decimal = Decimal("0.10")
    max_fee: Decimal
    calculated_fee: Decimal
    final_fee: Decimal
