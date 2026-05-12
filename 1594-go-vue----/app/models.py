from datetime import datetime, date, time, timedelta
from enum import Enum
from decimal import Decimal
from sqlalchemy import (
    Column,
    Integer,
    String,
    Text,
    Float,
    DateTime,
    Date,
    Time,
    Boolean,
    ForeignKey,
    Numeric,
)
from sqlalchemy.orm import relationship
from app.database import Base


class ReaderType(str, Enum):
    NORMAL = "normal"
    STUDENT = "student"
    TEACHER = "teacher"


class CopyStatus(str, Enum):
    AVAILABLE = "available"
    BORROWED = "borrowed"
    RESERVED = "reserved"
    MAINTENANCE = "maintenance"


class ReservationStatus(str, Enum):
    PENDING = "pending"
    FULFILLED = "fulfilled"
    CANCELLED = "cancelled"
    EXPIRED = "expired"


class SeatReservationStatus(str, Enum):
    RESERVED = "reserved"
    CHECKED_IN = "checked_in"
    CANCELLED = "cancelled"
    EXPIRED = "expired"


class EventRegistrationStatus(str, Enum):
    CONFIRMED = "confirmed"
    WAITLIST = "waitlist"
    CANCELLED = "cancelled"


READER_RULES = {
    ReaderType.NORMAL: {"max_books": 5, "loan_days": 30},
    ReaderType.STUDENT: {"max_books": 10, "loan_days": 60},
    ReaderType.TEACHER: {"max_books": 15, "loan_days": 90},
}


class Book(Base):
    __tablename__ = "books"

    id = Column(Integer, primary_key=True, index=True)
    isbn = Column(String(20), unique=True, index=True, nullable=False)
    title = Column(String(255), nullable=False)
    author = Column(String(255), nullable=False)
    publisher = Column(String(255))
    publish_date = Column(Date)
    category = Column(String(100))
    price = Column(Numeric(10, 2), nullable=False)
    description = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    copies = relationship("BookCopy", back_populates="book", cascade="all, delete-orphan")
    reservations = relationship("BookReservation", back_populates="book", cascade="all, delete-orphan")


class BookCopy(Base):
    __tablename__ = "book_copies"

    id = Column(Integer, primary_key=True, index=True)
    book_id = Column(Integer, ForeignKey("books.id"), nullable=False)
    copy_number = Column(String(50), unique=True, index=True, nullable=False)
    status = Column(String(20), default=CopyStatus.AVAILABLE.value, nullable=False)
    location = Column(String(100))
    condition = Column(String(50))
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    book = relationship("Book", back_populates="copies")
    borrowings = relationship("Borrowing", back_populates="copy")
    reservations = relationship("BookReservation", back_populates="copy")


class Reader(Base):
    __tablename__ = "readers"

    id = Column(Integer, primary_key=True, index=True)
    card_number = Column(String(50), unique=True, index=True, nullable=False)
    name = Column(String(100), nullable=False)
    reader_type = Column(String(20), nullable=False)
    phone = Column(String(20))
    email = Column(String(100))
    address = Column(Text)
    is_active = Column(Boolean, default=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    borrowings = relationship("Borrowing", back_populates="reader")
    reservations = relationship("BookReservation", back_populates="reader")
    seat_reservations = relationship("SeatReservation", back_populates="reader")
    event_registrations = relationship("EventRegistration", back_populates="reader")


class Borrowing(Base):
    __tablename__ = "borrowings"

    id = Column(Integer, primary_key=True, index=True)
    reader_id = Column(Integer, ForeignKey("readers.id"), nullable=False)
    copy_id = Column(Integer, ForeignKey("book_copies.id"), nullable=False)
    borrow_date = Column(Date, nullable=False)
    due_date = Column(Date, nullable=False)
    return_date = Column(Date)
    renew_count = Column(Integer, default=0)
    max_renew_count = Column(Integer, default=1)
    late_fee = Column(Numeric(10, 2), default=Decimal("0.00"))
    is_returned = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    reader = relationship("Reader", back_populates="borrowings")
    copy = relationship("BookCopy", back_populates="borrowings")


class BookReservation(Base):
    __tablename__ = "book_reservations"

    id = Column(Integer, primary_key=True, index=True)
    reader_id = Column(Integer, ForeignKey("readers.id"), nullable=False)
    book_id = Column(Integer, ForeignKey("books.id"), nullable=False)
    copy_id = Column(Integer, ForeignKey("book_copies.id"))
    status = Column(String(20), default=ReservationStatus.PENDING.value, nullable=False)
    queue_position = Column(Integer, nullable=False)
    reserved_at = Column(DateTime, default=datetime.utcnow)
    fulfilled_at = Column(DateTime)
    expires_at = Column(DateTime)
    notification_sent = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    reader = relationship("Reader", back_populates="reservations")
    book = relationship("Book", back_populates="reservations")
    copy = relationship("BookCopy", back_populates="reservations")


class Seat(Base):
    __tablename__ = "seats"

    id = Column(Integer, primary_key=True, index=True)
    seat_number = Column(String(20), unique=True, index=True, nullable=False)
    floor = Column(String(20))
    area = Column(String(50))
    is_active = Column(Boolean, default=True)
    has_power = Column(Boolean, default=False)
    notes = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)

    reservations = relationship("SeatReservation", back_populates="seat")


class SeatReservation(Base):
    __tablename__ = "seat_reservations"

    id = Column(Integer, primary_key=True, index=True)
    reader_id = Column(Integer, ForeignKey("readers.id"), nullable=False)
    seat_id = Column(Integer, ForeignKey("seats.id"), nullable=False)
    reservation_date = Column(Date, nullable=False)
    start_time = Column(Time, nullable=False)
    end_time = Column(Time, nullable=False)
    status = Column(String(20), default=SeatReservationStatus.RESERVED.value, nullable=False)
    check_in_time = Column(DateTime)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    reader = relationship("Reader", back_populates="seat_reservations")
    seat = relationship("Seat", back_populates="reservations")


class Event(Base):
    __tablename__ = "events"

    id = Column(Integer, primary_key=True, index=True)
    title = Column(String(255), nullable=False)
    description = Column(Text)
    event_date = Column(Date, nullable=False)
    start_time = Column(Time, nullable=False)
    end_time = Column(Time, nullable=False)
    location = Column(String(100))
    max_participants = Column(Integer, nullable=False)
    current_participants = Column(Integer, default=0)
    registration_start = Column(DateTime)
    registration_end = Column(DateTime)
    is_active = Column(Boolean, default=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    registrations = relationship("EventRegistration", back_populates="event", cascade="all, delete-orphan")


class EventRegistration(Base):
    __tablename__ = "event_registrations"

    id = Column(Integer, primary_key=True, index=True)
    event_id = Column(Integer, ForeignKey("events.id"), nullable=False)
    reader_id = Column(Integer, ForeignKey("readers.id"), nullable=False)
    status = Column(String(20), default=EventRegistrationStatus.WAITLIST.value, nullable=False)
    waitlist_position = Column(Integer)
    registered_at = Column(DateTime, default=datetime.utcnow)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    event = relationship("Event", back_populates="registrations")
    reader = relationship("Reader", back_populates="event_registrations")


class Notification(Base):
    __tablename__ = "notifications"

    id = Column(Integer, primary_key=True, index=True)
    reader_id = Column(Integer, ForeignKey("readers.id"), nullable=False)
    title = Column(String(255), nullable=False)
    message = Column(Text, nullable=False)
    is_read = Column(Boolean, default=False)
    notification_type = Column(String(50))
    created_at = Column(DateTime, default=datetime.utcnow)


SEAT_TIME_SLOTS = [
    (time(8, 0), time(10, 0)),
    (time(10, 0), time(12, 0)),
    (time(12, 0), time(14, 0)),
    (time(14, 0), time(16, 0)),
    (time(16, 0), time(18, 0)),
    (time(18, 0), time(20, 0)),
    (time(20, 0), time(21, 0)),
]
