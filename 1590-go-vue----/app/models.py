from datetime import datetime, date, time
from sqlalchemy import (
    Column, Integer, String, Float, DateTime, Date, Time,
    Boolean, Enum, ForeignKey, Text
)
from sqlalchemy.orm import relationship
from app.database import Base
import enum


class CabinStatus(enum.Enum):
    IDLE = "idle"
    UPWARD = "upward"
    DOWNWARD = "downward"
    MAINTENANCE = "maintenance"
    DISABLED = "disabled"


class MaintenanceType(enum.Enum):
    DAILY = "daily"
    WEEKLY = "weekly"
    MONTHLY = "monthly"
    YEARLY = "yearly"


class MaintenanceStatus(enum.Enum):
    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"
    OVERDUE = "overdue"


class Cabin(Base):
    __tablename__ = "cabins"

    id = Column(Integer, primary_key=True, index=True)
    cabin_number = Column(String(50), unique=True, index=True, nullable=False)
    max_weight = Column(Float, nullable=False, default=1000.0)
    current_status = Column(Enum(CabinStatus), default=CabinStatus.IDLE)
    current_weight = Column(Float, default=0.0)
    passenger_count = Column(Integer, default=0)
    last_dispatch_time = Column(DateTime, nullable=True)
    last_arrive_time = Column(DateTime, nullable=True)
    current_direction = Column(String(10), nullable=True)
    sensor_status_ok = Column(Boolean, default=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)


class QueueItem(Base):
    __tablename__ = "queue_items"

    id = Column(Integer, primary_key=True, index=True)
    passenger_count = Column(Integer, nullable=False, default=1)
    average_weight = Column(Float, nullable=False, default=70.0)
    join_time = Column(DateTime, default=datetime.utcnow)
    assigned_cabin_id = Column(Integer, ForeignKey("cabins.id"), nullable=True)
    is_served = Column(Boolean, default=False)
    served_time = Column(DateTime, nullable=True)


class TripRecord(Base):
    __tablename__ = "trip_records"

    id = Column(Integer, primary_key=True, index=True)
    cabin_id = Column(Integer, ForeignKey("cabins.id"), nullable=False)
    dispatch_time = Column(DateTime, nullable=False)
    arrive_time = Column(DateTime, nullable=True)
    direction = Column(String(10), nullable=False)
    weight = Column(Float, nullable=False)
    weight_is_estimated = Column(Boolean, default=False)
    passenger_count = Column(Integer, nullable=True)
    is_first_trip = Column(Boolean, default=False)
    has_passengers = Column(Boolean, default=True)
    created_at = Column(DateTime, default=datetime.utcnow)

    cabin = relationship("Cabin")


class MaintenanceTask(Base):
    __tablename__ = "maintenance_tasks"

    id = Column(Integer, primary_key=True, index=True)
    task_type = Column(Enum(MaintenanceType), nullable=False)
    cabin_id = Column(Integer, ForeignKey("cabins.id"), nullable=True)
    scheduled_date = Column(Date, nullable=False)
    scheduled_time = Column(Time, nullable=True)
    status = Column(Enum(MaintenanceStatus), default=MaintenanceStatus.PENDING)
    start_time = Column(DateTime, nullable=True)
    complete_time = Column(DateTime, nullable=True)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    cabin = relationship("Cabin")


class ShiftRecord(Base):
    __tablename__ = "shift_records"

    id = Column(Integer, primary_key=True, index=True)
    shift_date = Column(Date, nullable=False)
    shift_number = Column(Integer, nullable=False)
    start_time = Column(DateTime, nullable=False)
    end_time = Column(DateTime, nullable=True)
    total_passengers = Column(Integer, nullable=True)
    total_weight = Column(Float, nullable=True)
    total_trips = Column(Integer, nullable=True)
    total_revenue = Column(Float, nullable=True)
    average_wait_time = Column(Float, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
