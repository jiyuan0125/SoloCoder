from datetime import datetime
from sqlalchemy import Column, Integer, Float, String, Boolean, DateTime, ForeignKey, Text, Enum as SQLEnum
from sqlalchemy.orm import relationship
from app.database import Base
import enum


class RopewayStatus(str, enum.Enum):
    NORMAL = "normal"
    DECELERATE = "decelerate"
    PAUSE = "pause"
    STOP = "stop"


class InspectionType(str, enum.Enum):
    DAILY = "daily"
    WEEKLY = "weekly"
    MONTHLY = "monthly"
    YEARLY = "yearly"


class InspectionStatus(str, enum.Enum):
    PENDING = "pending"
    COMPLETED = "completed"
    URGENT = "urgent"


class WeatherData(Base):
    __tablename__ = "weather_data"
    
    id = Column(Integer, primary_key=True, index=True)
    timestamp = Column(DateTime, default=datetime.utcnow, index=True)
    wind_speed = Column(Float, nullable=False)
    has_lightning = Column(Boolean, default=False)
    temperature = Column(Float, nullable=True)
    humidity = Column(Float, nullable=True)
    
    status_records = relationship("StatusRecord", back_populates="weather_data")


class StatusRecord(Base):
    __tablename__ = "status_records"
    
    id = Column(Integer, primary_key=True, index=True)
    timestamp = Column(DateTime, default=datetime.utcnow, index=True)
    from_status = Column(SQLEnum(RopewayStatus), nullable=True)
    to_status = Column(SQLEnum(RopewayStatus), nullable=False)
    reason = Column(String(255), nullable=False)
    weather_data_id = Column(Integer, ForeignKey("weather_data.id"), nullable=True)
    operator_id = Column(Integer, ForeignKey("operators.id"), nullable=True)
    
    weather_data = relationship("WeatherData", back_populates="status_records")
    operator = relationship("Operator", back_populates="status_records")


class Operator(Base):
    __tablename__ = "operators"
    
    id = Column(Integer, primary_key=True, index=True)
    username = Column(String(50), unique=True, nullable=False)
    full_name = Column(String(100), nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    
    status_records = relationship("StatusRecord", back_populates="operator")
    inspections = relationship("Inspection", back_populates="operator")


class Inspection(Base):
    __tablename__ = "inspections"
    
    id = Column(Integer, primary_key=True, index=True)
    inspection_type = Column(SQLEnum(InspectionType), nullable=False)
    scheduled_date = Column(DateTime, nullable=False, index=True)
    completed_date = Column(DateTime, nullable=True)
    status = Column(SQLEnum(InspectionStatus), default=InspectionStatus.PENDING)
    is_urgent = Column(Boolean, default=False)
    resolved = Column(Boolean, default=True)
    issues_found = Column(Text, nullable=True)
    notes = Column(Text, nullable=True)
    operator_id = Column(Integer, ForeignKey("operators.id"), nullable=True)
    
    operator = relationship("Operator", back_populates="inspections")


class GondolaCapacity(Base):
    __tablename__ = "gondola_capacity"
    
    id = Column(Integer, primary_key=True, index=True)
    total_capacity = Column(Integer, nullable=False)
    effective_date = Column(DateTime, default=datetime.utcnow)
    is_active = Column(Boolean, default=True)


class PassengerRecord(Base):
    __tablename__ = "passenger_records"
    
    id = Column(Integer, primary_key=True, index=True)
    timestamp = Column(DateTime, default=datetime.utcnow, index=True)
    gondola_id = Column(Integer, nullable=False)
    adult_count = Column(Integer, default=0)
    child_count = Column(Integer, default=0)
    total_counted = Column(Integer, default=0)
    
    @property
    def total_actual(self):
        return self.adult_count + self.child_count


class QueueData(Base):
    __tablename__ = "queue_data"
    
    id = Column(Integer, primary_key=True, index=True)
    timestamp = Column(DateTime, default=datetime.utcnow, index=True)
    queue_count = Column(Integer, nullable=False)
    gondola_capacity = Column(Integer, nullable=False)
    is_limited = Column(Boolean, default=False)


class DailyReport(Base):
    __tablename__ = "daily_reports"
    
    id = Column(Integer, primary_key=True, index=True)
    report_date = Column(DateTime, index=True, unique=True)
    total_operating_hours = Column(Float, default=0.0)
    effective_operating_hours = Column(Float, default=0.0)
    total_passengers = Column(Integer, default=0)
    total_passengers_counted = Column(Integer, default=0)
    below_min_hours = Column(Boolean, default=False)
    weather_events = Column(Text, nullable=True)
    urgent_inspections = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
