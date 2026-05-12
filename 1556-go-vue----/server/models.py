from datetime import date, datetime
from sqlalchemy import (
    Column, Integer, String, Date, Float, ForeignKey, Enum, Boolean, Text, DateTime
)
from sqlalchemy.orm import relationship
from .config import Base
import enum


class CrewStatus(str, enum.Enum):
    AVAILABLE = "待派"
    ON_BOARD = "在船"
    OFF_BOARD = "下船"


class AttendanceType(str, enum.Enum):
    NORMAL = "正常"
    OVERTIME = "加班"
    SICK_LEAVE = "病假"
    PERSONAL_LEAVE = "事假"


class CertificateStatus(str, enum.Enum):
    VALID = "有效"
    WARNING_YELLOW = "黄色警告"
    WARNING_ORANGE = "橙色警告"
    EXPIRED = "过期"


class Crew(Base):
    __tablename__ = "crew"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), nullable=False)
    id_number = Column(String(50), unique=True, nullable=False)
    phone = Column(String(20))
    email = Column(String(100))
    is_intern = Column(Boolean, default=False)
    status = Column(String(50), default=CrewStatus.AVAILABLE.value)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    certificates = relationship("Certificate", back_populates="crew", cascade="all, delete-orphan")
    training_records = relationship("TrainingRecord", back_populates="crew", cascade="all, delete-orphan")
    assignments = relationship("Assignment", back_populates="crew")
    attendance_records = relationship("AttendanceRecord", back_populates="crew", cascade="all, delete-orphan")
    salary_records = relationship("SalaryRecord", back_populates="crew", cascade="all, delete-orphan")


class Certificate(Base):
    __tablename__ = "certificates"

    id = Column(Integer, primary_key=True, index=True)
    crew_id = Column(Integer, ForeignKey("crew.id"), nullable=False)
    certificate_type = Column(String(100), nullable=False)
    certificate_number = Column(String(100), unique=True, nullable=False)
    issue_date = Column(Date, nullable=False)
    expiry_date = Column(Date, nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    crew = relationship("Crew", back_populates="certificates")


class TrainingRecord(Base):
    __tablename__ = "training_records"

    id = Column(Integer, primary_key=True, index=True)
    crew_id = Column(Integer, ForeignKey("crew.id"), nullable=False)
    training_type = Column(String(100), nullable=False)
    training_date = Column(Date, nullable=False)
    description = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)

    crew = relationship("Crew", back_populates="training_records")


class Ship(Base):
    __tablename__ = "ships"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), unique=True, nullable=False)
    imo_number = Column(String(50), unique=True)
    capacity = Column(Integer, nullable=False)
    description = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    assignments = relationship("Assignment", back_populates="ship")


class Assignment(Base):
    __tablename__ = "assignments"

    id = Column(Integer, primary_key=True, index=True)
    crew_id = Column(Integer, ForeignKey("crew.id"), nullable=False)
    ship_id = Column(Integer, ForeignKey("ships.id"), nullable=False)
    position = Column(String(100), nullable=False)
    is_watch_keeper = Column(Boolean, default=False)
    start_date = Column(Date, nullable=False)
    end_date = Column(Date, nullable=True)
    is_active = Column(Boolean, default=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    crew = relationship("Crew", back_populates="assignments")
    ship = relationship("Ship", back_populates="assignments")


class AttendanceRecord(Base):
    __tablename__ = "attendance_records"

    id = Column(Integer, primary_key=True, index=True)
    crew_id = Column(Integer, ForeignKey("crew.id"), nullable=False)
    attendance_date = Column(Date, nullable=False)
    attendance_type = Column(String(50), nullable=False)
    assignment_id = Column(Integer, ForeignKey("assignments.id"), nullable=True)
    ship_id = Column(Integer, ForeignKey("ships.id"), nullable=True)
    is_half_day = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)

    crew = relationship("Crew", back_populates="attendance_records")


class SalaryRecord(Base):
    __tablename__ = "salary_records"

    id = Column(Integer, primary_key=True, index=True)
    crew_id = Column(Integer, ForeignKey("crew.id"), nullable=False)
    year = Column(Integer, nullable=False)
    month = Column(Integer, nullable=False)
    base_salary = Column(Float, default=0.0)
    normal_days = Column(Float, default=0.0)
    overtime_days = Column(Float, default=0.0)
    sick_leave_days = Column(Float, default=0.0)
    personal_leave_days = Column(Float, default=0.0)
    total_amount = Column(Float, default=0.0)
    ship_id = Column(Integer, ForeignKey("ships.id"), nullable=True)
    assignment_id = Column(Integer, ForeignKey("assignments.id"), nullable=True)
    is_settled = Column(Boolean, default=False)
    settlement_date = Column(Date, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    crew = relationship("Crew", back_populates="salary_records")
