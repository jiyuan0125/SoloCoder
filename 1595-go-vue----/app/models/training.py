from datetime import datetime
from sqlalchemy import Column, Integer, String, DateTime, ForeignKey, Enum as SQLEnum, Text
from sqlalchemy.orm import relationship
import enum

from app.database import Base


class TrainingStatus(str, enum.Enum):
    REGISTRATION = "registration"
    IN_PROGRESS = "in_progress"
    FINISHED = "finished"


class RegistrationStatus(str, enum.Enum):
    WAITLIST = "waitlist"
    CONFIRMED = "confirmed"
    CANCELLED = "cancelled"


class Training(Base):
    __tablename__ = "trainings"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(200), nullable=False)
    description = Column(Text, nullable=True)
    instructor = Column(String(100), nullable=True)
    max_participants = Column(Integer, nullable=False)
    venue_id = Column(Integer, ForeignKey("venues.id"), nullable=False)
    start_date = Column(DateTime, nullable=False)
    end_date = Column(DateTime, nullable=False)
    schedule = Column(String(200), nullable=True)
    status = Column(SQLEnum(TrainingStatus), default=TrainingStatus.REGISTRATION)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    venue = relationship("Venue", back_populates="trainings")
    registrations = relationship("Registration", back_populates="training")


class Registration(Base):
    __tablename__ = "registrations"

    id = Column(Integer, primary_key=True, index=True)
    training_id = Column(Integer, ForeignKey("trainings.id"), nullable=False)
    student_name = Column(String(100), nullable=False)
    student_phone = Column(String(20), nullable=True)
    student_email = Column(String(100), nullable=True)
    status = Column(SQLEnum(RegistrationStatus), default=RegistrationStatus.WAITLIST)
    waitlist_order = Column(Integer, nullable=True)
    consecutive_absences = Column(Integer, default=0)
    registered_at = Column(DateTime, default=datetime.utcnow)
    confirmed_at = Column(DateTime, nullable=True)
    cancelled_at = Column(DateTime, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    training = relationship("Training", back_populates="registrations")
    attendances = relationship("Attendance", back_populates="registration")


class Attendance(Base):
    __tablename__ = "attendances"

    id = Column(Integer, primary_key=True, index=True)
    registration_id = Column(Integer, ForeignKey("registrations.id"), nullable=False)
    session_date = Column(DateTime, nullable=False)
    is_present = Column(Integer, default=0)
    created_at = Column(DateTime, default=datetime.utcnow)

    registration = relationship("Registration", back_populates="attendances")
