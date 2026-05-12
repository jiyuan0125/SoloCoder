import enum
from datetime import datetime, date, time
from sqlalchemy import Column, Integer, String, DateTime, Date, Time, ForeignKey, Text, Boolean
from sqlalchemy.orm import relationship
from app.database import Base


class CollectionStatus(str, enum.Enum):
    IN_STOCK = "在库"
    ON_EXHIBIT = "展出"
    BORROWED = "外借"
    UNDER_REPAIR = "修复中"
    DECOMMISSIONED = "已注销"


class CollectionLevel(str, enum.Enum):
    FIRST_LEVEL = "一级"
    SECOND_LEVEL = "二级"
    THIRD_LEVEL = "三级"
    GENERAL = "一般"


class ExhibitionStatus(str, enum.Enum):
    PLANNING = "策划"
    SETUP = "布展"
    ON_DISPLAY = "展出"
    TEARDOWN = "撤展"
    COMPLETED = "已完成"


class Collection(Base):
    __tablename__ = "collections"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(200), nullable=False)
    description = Column(Text, nullable=True)
    level = Column(String(20), nullable=False, default=CollectionLevel.GENERAL.value)
    status = Column(String(20), nullable=False, default=CollectionStatus.IN_STOCK.value)
    location = Column(String(100), nullable=True)
    operator = Column(String(100), nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    inventory_logs = relationship("InventoryLog", back_populates="collection", cascade="all, delete-orphan")
    exhibition_items = relationship("ExhibitionItem", back_populates="collection")
    borrowings = relationship("Borrowing", back_populates="collection")


class InventoryLog(Base):
    __tablename__ = "inventory_logs"

    id = Column(Integer, primary_key=True, index=True)
    collection_id = Column(Integer, ForeignKey("collections.id"), nullable=False)
    operation_type = Column(String(50), nullable=False)
    operator = Column(String(100), nullable=False)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)

    collection = relationship("Collection", back_populates="inventory_logs")


class Exhibition(Base):
    __tablename__ = "exhibitions"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(200), nullable=False)
    description = Column(Text, nullable=True)
    location = Column(String(100), nullable=True)
    status = Column(String(20), nullable=False, default=ExhibitionStatus.PLANNING.value)
    start_date = Column(Date, nullable=True)
    end_date = Column(Date, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    items = relationship("ExhibitionItem", back_populates="exhibition", cascade="all, delete-orphan")


class ExhibitionItem(Base):
    __tablename__ = "exhibition_items"

    id = Column(Integer, primary_key=True, index=True)
    exhibition_id = Column(Integer, ForeignKey("exhibitions.id"), nullable=False)
    collection_id = Column(Integer, ForeignKey("collections.id"), nullable=False)
    added_at = Column(DateTime, default=datetime.utcnow)

    exhibition = relationship("Exhibition", back_populates="items")
    collection = relationship("Collection", back_populates="exhibition_items")


class VisitReservation(Base):
    __tablename__ = "visit_reservations"

    id = Column(Integer, primary_key=True, index=True)
    visitor_name = Column(String(100), nullable=False)
    id_number = Column(String(50), nullable=False, index=True)
    phone = Column(String(20), nullable=True)
    visit_date = Column(Date, nullable=False, index=True)
    time_slot = Column(String(50), nullable=False)
    visitor_count = Column(Integer, default=1)
    status = Column(String(20), nullable=False, default="已预约")
    created_at = Column(DateTime, default=datetime.utcnow)
    cancelled_at = Column(DateTime, nullable=True)


class TimeSlotConfig(Base):
    __tablename__ = "time_slot_configs"

    id = Column(Integer, primary_key=True, index=True)
    time_slot = Column(String(50), nullable=False, unique=True)
    max_visitors = Column(Integer, nullable=False)
    is_active = Column(Boolean, default=True)


class MeetingReservation(Base):
    __tablename__ = "meeting_reservations"

    id = Column(Integer, primary_key=True, index=True)
    title = Column(String(200), nullable=False)
    organizer = Column(String(100), nullable=False)
    meeting_date = Column(Date, nullable=False, index=True)
    start_time = Column(Time, nullable=False)
    end_time = Column(Time, nullable=False)
    room = Column(String(100), nullable=False)
    attendees = Column(Integer, default=1)
    notes = Column(Text, nullable=True)
    status = Column(String(20), nullable=False, default="已预约")
    created_at = Column(DateTime, default=datetime.utcnow)


class Borrowing(Base):
    __tablename__ = "borrowings"

    id = Column(Integer, primary_key=True, index=True)
    collection_id = Column(Integer, ForeignKey("collections.id"), nullable=False)
    borrower = Column(String(200), nullable=False)
    contact = Column(String(100), nullable=True)
    borrow_date = Column(Date, nullable=False)
    expected_return_date = Column(Date, nullable=False)
    actual_return_date = Column(Date, nullable=True)
    purpose = Column(Text, nullable=True)
    status = Column(String(20), nullable=False, default="外借中")
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    collection = relationship("Collection", back_populates="borrowings")
    reminders = relationship("ReminderTodo", back_populates="borrowing", cascade="all, delete-orphan")


class ReminderTodo(Base):
    __tablename__ = "reminder_todos"

    id = Column(Integer, primary_key=True, index=True)
    borrowing_id = Column(Integer, ForeignKey("borrowings.id"), nullable=False)
    overdue_days = Column(Integer, nullable=False)
    message = Column(Text, nullable=False)
    is_completed = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    completed_at = Column(DateTime, nullable=True)

    borrowing = relationship("Borrowing", back_populates="reminders")
