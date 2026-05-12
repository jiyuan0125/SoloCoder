from datetime import date, datetime
from sqlalchemy import Column, Integer, String, Date, DateTime, Float, ForeignKey, Boolean, Text
from sqlalchemy.orm import relationship

from .database import Base


class Pilot(Base):
    __tablename__ = "pilots"

    id = Column(Integer, primary_key=True, index=True)
    employee_id = Column(String(50), unique=True, index=True, nullable=False)
    name = Column(String(100), nullable=False)
    gender = Column(String(10))
    birth_date = Column(Date, nullable=False)
    english_level = Column(Integer, default=1)
    
    status = Column(String(20), default="standby", nullable=False)
    
    monthly_hours = Column(Float, default=0.0)
    yearly_hours = Column(Float, default=0.0)
    priority_score = Column(Integer, default=100)
    
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    qualifications = relationship("Qualification", back_populates="pilot", cascade="all, delete-orphan")
    medicals = relationship("Medical", back_populates="pilot", cascade="all, delete-orphan")
    schedules = relationship("FlightSchedule", back_populates="pilot", cascade="all, delete-orphan")


class Qualification(Base):
    __tablename__ = "qualifications"

    id = Column(Integer, primary_key=True, index=True)
    pilot_id = Column(Integer, ForeignKey("pilots.id"), nullable=False)
    qualification_type = Column(String(50), nullable=False)
    aircraft_type = Column(String(50))
    certificate_number = Column(String(100))
    issue_date = Column(Date, nullable=False)
    expiry_date = Column(Date, nullable=False)
    last_used_date = Column(Date)
    is_mandatory = Column(Boolean, default=False)
    is_valid = Column(Boolean, default=True)
    
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    pilot = relationship("Pilot", back_populates="qualifications")


class Medical(Base):
    __tablename__ = "medicals"

    id = Column(Integer, primary_key=True, index=True)
    pilot_id = Column(Integer, ForeignKey("pilots.id"), nullable=False)
    examination_date = Column(Date, nullable=False)
    expiry_date = Column(Date, nullable=False)
    result = Column(String(20), default="qualified", nullable=False)
    medical_type = Column(String(20), default="class1")
    notes = Column(Text)
    
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    pilot = relationship("Pilot", back_populates="medicals")


class FlightSchedule(Base):
    __tablename__ = "flight_schedules"

    id = Column(Integer, primary_key=True, index=True)
    pilot_id = Column(Integer, ForeignKey("pilots.id"), nullable=False)
    flight_number = Column(String(50), nullable=False)
    departure_airport = Column(String(10), nullable=False)
    arrival_airport = Column(String(10), nullable=False)
    aircraft_type = Column(String(50), nullable=False)
    is_international = Column(Boolean, default=False)
    
    departure_time = Column(DateTime, nullable=False)
    arrival_time = Column(DateTime, nullable=False)
    flight_duration = Column(Float, nullable=False)
    
    status = Column(String(20), default="pending", nullable=False)
    validation_result = Column(Text)
    validation_passed = Column(Boolean, default=True)
    
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    pilot = relationship("Pilot", back_populates="schedules")
