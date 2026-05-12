from sqlalchemy import (
    Column, Integer, String, DateTime, Float, Boolean, ForeignKey, Enum, Text
)
from sqlalchemy.orm import relationship, DeclarativeBase
import enum
from datetime import datetime


class Base(DeclarativeBase):
    pass


class PilotLevel(str, enum.Enum):
    LEVEL1 = "level1"
    LEVEL2 = "level2"
    LEVEL3 = "level3"


class ShipType(str, enum.Enum):
    GENERAL_CARGO = "general_cargo"
    PASSENGER = "passenger"
    OIL_TANKER = "oil_tanker"
    LARGE_OIL_TANKER = "large_oil_tanker"


class TaskStatus(str, enum.Enum):
    PENDING = "pending"
    DISPATCHED = "dispatched"
    IN_PROGRESS = "in_progress"
    SUSPENDED = "suspended"
    COMPLETED = "completed"


class WeatherCondition(str, enum.Enum):
    GOOD = "good"
    BAD = "bad"


class Pilot(Base):
    __tablename__ = "pilots"
    
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), nullable=False)
    level = Column(Enum(PilotLevel), nullable=False)
    is_on_duty = Column(Boolean, default=False)
    total_work_hours = Column(Float, default=0.0)
    monthly_work_hours = Column(Float, default=0.0)
    monthly_task_count = Column(Integer, default=0)
    created_at = Column(DateTime, default=datetime.utcnow)
    
    tasks = relationship("PilotTask", back_populates="pilot")


class PilotApplication(Base):
    __tablename__ = "pilot_applications"
    
    id = Column(Integer, primary_key=True, index=True)
    ship_name = Column(String(100), nullable=False)
    ship_type = Column(Enum(ShipType), nullable=False)
    from_location = Column(String(100), nullable=False)
    to_location = Column(String(100), nullable=False)
    requested_start_time = Column(DateTime, nullable=False)
    estimated_duration_hours = Column(Float, nullable=False)
    remarks = Column(Text, nullable=True)
    status = Column(String(20), default="pending")
    created_at = Column(DateTime, default=datetime.utcnow)
    
    task = relationship("PilotTask", back_populates="application", uselist=False)


class PilotTask(Base):
    __tablename__ = "pilot_tasks"
    
    id = Column(Integer, primary_key=True, index=True)
    application_id = Column(Integer, ForeignKey("pilot_applications.id"), nullable=False)
    pilot_id = Column(Integer, ForeignKey("pilots.id"), nullable=False)
    status = Column(Enum(TaskStatus), default=TaskStatus.DISPATCHED)
    assigned_at = Column(DateTime, default=datetime.utcnow)
    started_at = Column(DateTime, nullable=True)
    completed_at = Column(DateTime, nullable=True)
    suspended_at = Column(DateTime, nullable=True)
    actual_duration_hours = Column(Float, nullable=True)
    is_abnormal = Column(Boolean, default=False)
    suspension_reason = Column(Text, nullable=True)
    progress_percent = Column(Integer, default=0)
    
    application = relationship("PilotApplication", back_populates="task")
    pilot = relationship("Pilot", back_populates="tasks")


class WeatherStatus(Base):
    __tablename__ = "weather_status"
    
    id = Column(Integer, primary_key=True, index=True)
    condition = Column(Enum(WeatherCondition), default=WeatherCondition.GOOD)
    description = Column(String(200), nullable=True)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)


class TimeWindowConfig(Base):
    __tablename__ = "time_window_configs"
    
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(50), nullable=False, unique=True)
    start_time = Column(String(5), nullable=False)
    end_time = Column(String(5), nullable=False)
    description = Column(String(200), nullable=True)
    is_active = Column(Boolean, default=True)
    rule_type = Column(String(20), nullable=False)


class MonthlyStats(Base):
    __tablename__ = "monthly_stats"
    
    id = Column(Integer, primary_key=True, index=True)
    pilot_id = Column(Integer, ForeignKey("pilots.id"), nullable=False)
    year = Column(Integer, nullable=False)
    month = Column(Integer, nullable=False)
    total_hours = Column(Float, default=0.0)
    task_count = Column(Integer, default=0)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    
    __table_args__ = (
        {'sqlite_autoincrement': True},
    )
