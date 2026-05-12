from datetime import datetime, timedelta
from sqlalchemy import Column, Integer, String, DateTime, ForeignKey, Enum, Boolean, Float, Text
from sqlalchemy.orm import relationship
from .database import Base
import enum


class MaintenanceGrade(str, enum.Enum):
    LEVEL_1 = "level_1"
    LEVEL_2 = "level_2"
    LEVEL_3 = "level_3"


class FacilityType(str, enum.Enum):
    PLAYGROUND = "playground"
    RESTROOM = "restroom"
    BENCH = "bench"
    LIGHTING = "lighting"
    FOUNTAIN = "fountain"
    OTHER = "other"


class FacilityStatus(str, enum.Enum):
    NORMAL = "normal"
    DAMAGED = "damaged"
    REPAIRING = "repairing"
    DISABLED = "disabled"


class ActivityStatus(str, enum.Enum):
    PENDING = "pending"
    APPROVED = "approved"
    REJECTED = "rejected"
    COMPLETED = "completed"


class TaskStatus(str, enum.Enum):
    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"
    OVERDUE = "overdue"


class Season(str, enum.Enum):
    SPRING = "spring"
    SUMMER = "summer"
    AUTUMN = "autumn"
    WINTER = "winter"


def get_maintenance_frequency(grade: MaintenanceGrade) -> int:
    frequency_map = {
        MaintenanceGrade.LEVEL_1: 7,
        MaintenanceGrade.LEVEL_2: 14,
        MaintenanceGrade.LEVEL_3: 30,
    }
    return frequency_map.get(grade, 30)


class Park(Base):
    __tablename__ = "parks"

    id = Column(Integer, primary_key=True, index=True)
    park_code = Column(String(50), unique=True, index=True, nullable=False)
    name = Column(String(200), nullable=False)
    location = Column(String(500))
    area = Column(Float)
    created_at = Column(DateTime, default=datetime.utcnow)

    green_zones = relationship("GreenZone", back_populates="park", cascade="all, delete-orphan")
    facilities = relationship("Facility", back_populates="park", cascade="all, delete-orphan")
    activities = relationship("Activity", back_populates="park", cascade="all, delete-orphan")


class GreenZone(Base):
    __tablename__ = "green_zones"

    id = Column(Integer, primary_key=True, index=True)
    zone_code = Column(String(50), unique=True, index=True, nullable=False)
    name = Column(String(200), nullable=False)
    park_id = Column(Integer, ForeignKey("parks.id"), nullable=False)
    area = Column(Float)
    current_grade = Column(Enum(MaintenanceGrade), default=MaintenanceGrade.LEVEL_2)
    plant_types = Column(String(500))
    location_description = Column(String(500))
    consecutive_missed_tasks = Column(Integer, default=0)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    park = relationship("Park", back_populates="green_zones")
    maintenance_tasks = relationship("MaintenanceTask", back_populates="green_zone", cascade="all, delete-orphan")
    grade_history = relationship("GradeAdjustment", back_populates="green_zone", cascade="all, delete-orphan")


class GradeAdjustment(Base):
    __tablename__ = "grade_adjustments"

    id = Column(Integer, primary_key=True, index=True)
    green_zone_id = Column(Integer, ForeignKey("green_zones.id"), nullable=False)
    old_grade = Column(Enum(MaintenanceGrade), nullable=False)
    new_grade = Column(Enum(MaintenanceGrade), nullable=False)
    reason = Column(String(500))
    season = Column(Enum(Season))
    adjusted_by = Column(String(100))
    adjusted_at = Column(DateTime, default=datetime.utcnow)

    green_zone = relationship("GreenZone", back_populates="grade_history")


class MaintenanceTask(Base):
    __tablename__ = "maintenance_tasks"

    id = Column(Integer, primary_key=True, index=True)
    task_code = Column(String(50), unique=True, index=True, nullable=False)
    green_zone_id = Column(Integer, ForeignKey("green_zones.id"), nullable=False)
    maintenance_grade = Column(Enum(MaintenanceGrade), nullable=False)
    scheduled_date = Column(DateTime, nullable=False)
    status = Column(Enum(TaskStatus), default=TaskStatus.PENDING)
    actual_completion_date = Column(DateTime)
    executed_by = Column(String(100))
    notes = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    green_zone = relationship("GreenZone", back_populates="maintenance_tasks")


class Facility(Base):
    __tablename__ = "facilities"

    id = Column(Integer, primary_key=True, index=True)
    facility_code = Column(String(50), unique=True, index=True, nullable=False)
    name = Column(String(200), nullable=False)
    park_id = Column(Integer, ForeignKey("parks.id"), nullable=False)
    facility_type = Column(Enum(FacilityType), nullable=False)
    is_safety_related = Column(Boolean, default=False)
    status = Column(Enum(FacilityStatus), default=FacilityStatus.NORMAL)
    location_description = Column(String(500))
    installation_date = Column(DateTime)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    park = relationship("Park", back_populates="facilities")
    reports = relationship("FacilityReport", back_populates="facility", cascade="all, delete-orphan")


class FacilityReport(Base):
    __tablename__ = "facility_reports"

    id = Column(Integer, primary_key=True, index=True)
    report_code = Column(String(50), unique=True, index=True, nullable=False)
    facility_id = Column(Integer, ForeignKey("facilities.id"), nullable=False)
    reporter_name = Column(String(200))
    reporter_contact = Column(String(100))
    damage_description = Column(Text, nullable=False)
    reported_at = Column(DateTime, default=datetime.utcnow)
    deadline = Column(DateTime)
    handled_by = Column(String(100))
    handled_at = Column(DateTime)
    handling_notes = Column(Text)
    is_handled = Column(Boolean, default=False)

    facility = relationship("Facility", back_populates="reports")


class Activity(Base):
    __tablename__ = "activities"

    id = Column(Integer, primary_key=True, index=True)
    activity_code = Column(String(50), unique=True, index=True, nullable=False)
    park_id = Column(Integer, ForeignKey("parks.id"), nullable=False)
    name = Column(String(200), nullable=False)
    organizer = Column(String(200))
    organizer_contact = Column(String(100))
    area = Column(Float, nullable=False)
    start_time = Column(DateTime, nullable=False)
    end_time = Column(DateTime, nullable=False)
    expected_participants = Column(Integer)
    requires_security = Column(Boolean, default=False)
    status = Column(Enum(ActivityStatus), default=ActivityStatus.PENDING)
    approved_by = Column(String(100))
    approved_at = Column(DateTime)
    rejection_reason = Column(String(500))
    completion_notes = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)

    park = relationship("Park", back_populates="activities")
