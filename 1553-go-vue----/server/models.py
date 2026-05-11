from datetime import datetime, date
from enum import Enum as PyEnum
from typing import Optional, List

from sqlalchemy import (
    Column, Integer, String, Float, DateTime, Date, Boolean, ForeignKey, Enum, Text
)
from sqlalchemy.orm import relationship, Mapped

from server.database import Base


class BeaconStatus(str, PyEnum):
    NORMAL = "normal"
    FAULT = "fault"
    MISSING = "missing"


class DredgingStatus(str, PyEnum):
    PLANNED = "planned"
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"
    CANCELLED = "cancelled"


class CheckResult(str, PyEnum):
    PENDING = "pending"
    PASSED = "passed"
    FAILED = "failed"


class TodoPriority(str, PyEnum):
    URGENT = "urgent"
    HIGH = "high"
    MEDIUM = "medium"
    LOW = "low"


class TodoStatus(str, PyEnum):
    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"


class WarningType(str, PyEnum):
    DEPT_DECLINE = "depth_decline"
    BEACON_ANOMALY = "beacon_anomaly"
    RESTRICTED = "restricted"


class WarningStatus(str, PyEnum):
    ACTIVE = "active"
    ACKNOWLEDGED = "acknowledged"


class Section(Base):
    __tablename__ = "sections"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(200), nullable=False, index=True)
    design_depth = Column(Float, nullable=False)
    current_depth = Column(Float, nullable=True)
    is_restricted = Column(Boolean, default=False)
    beacon_anomaly = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    beacons: Mapped[List["Beacon"]] = relationship("Beacon", back_populates="section", cascade="all, delete-orphan")
    depth_records: Mapped[List["DepthRecord"]] = relationship("DepthRecord", back_populates="section", cascade="all, delete-orphan")
    dredging_plans: Mapped[List["DredgingPlan"]] = relationship("DredgingPlan", back_populates="section", cascade="all, delete-orphan")
    draft_declarations: Mapped[List["DraftDeclaration"]] = relationship("DraftDeclaration", back_populates="section", cascade="all, delete-orphan")
    warnings: Mapped[List["Warning"]] = relationship("Warning", back_populates="section", cascade="all, delete-orphan")


class Beacon(Base):
    __tablename__ = "beacons"

    id = Column(Integer, primary_key=True, index=True)
    section_id = Column(Integer, ForeignKey("sections.id"), nullable=False, index=True)
    identifier = Column(String(100), nullable=False)
    name = Column(String(200), nullable=True)
    status = Column(Enum(BeaconStatus), default=BeaconStatus.NORMAL, nullable=False)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    section: Mapped["Section"] = relationship("Section", back_populates="beacons")


class DepthRecord(Base):
    __tablename__ = "depth_records"

    id = Column(Integer, primary_key=True, index=True)
    section_id = Column(Integer, ForeignKey("sections.id"), nullable=False, index=True)
    measured_depth = Column(Float, nullable=False)
    record_date = Column(Date, nullable=False, index=True)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)

    section: Mapped["Section"] = relationship("Section", back_populates="depth_records")


class DredgingPlan(Base):
    __tablename__ = "dredging_plans"

    id = Column(Integer, primary_key=True, index=True)
    section_id = Column(Integer, ForeignKey("sections.id"), nullable=False, index=True)
    target_depth = Column(Float, nullable=False)
    status = Column(Enum(DredgingStatus), default=DredgingStatus.PLANNED, nullable=False)
    planned_start_date = Column(Date, nullable=True)
    planned_end_date = Column(Date, nullable=True)
    actual_start_date = Column(Date, nullable=True)
    actual_end_date = Column(Date, nullable=True)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    section: Mapped["Section"] = relationship("Section", back_populates="dredging_plans")


class DraftDeclaration(Base):
    __tablename__ = "draft_declarations"

    id = Column(Integer, primary_key=True, index=True)
    section_id = Column(Integer, ForeignKey("sections.id"), nullable=False, index=True)
    ship_name = Column(String(200), nullable=False)
    imo_number = Column(String(50), nullable=True)
    declared_draft = Column(Float, nullable=False)
    check_result = Column(Enum(CheckResult), default=CheckResult.PENDING, nullable=False)
    check_message = Column(Text, nullable=True)
    checked_at = Column(DateTime, nullable=True)
    is_completed = Column(Boolean, default=False)
    completed_at = Column(DateTime, nullable=True)
    effective_depth_at_check = Column(Float, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)

    section: Mapped["Section"] = relationship("Section", back_populates="draft_declarations")


class Todo(Base):
    __tablename__ = "todos"

    id = Column(Integer, primary_key=True, index=True)
    title = Column(String(500), nullable=False)
    description = Column(Text, nullable=True)
    priority = Column(Enum(TodoPriority), default=TodoPriority.MEDIUM, nullable=False)
    status = Column(Enum(TodoStatus), default=TodoStatus.PENDING, nullable=False)
    deadline = Column(DateTime, nullable=True)
    section_id = Column(Integer, ForeignKey("sections.id"), nullable=True)
    related_warning_id = Column(Integer, ForeignKey("warnings.id"), nullable=True)
    completed_at = Column(DateTime, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)


class Warning(Base):
    __tablename__ = "warnings"

    id = Column(Integer, primary_key=True, index=True)
    section_id = Column(Integer, ForeignKey("sections.id"), nullable=False, index=True)
    warning_type = Column(Enum(WarningType), nullable=False)
    message = Column(Text, nullable=False)
    status = Column(Enum(WarningStatus), default=WarningStatus.ACTIVE, nullable=False)
    acknowledged_at = Column(DateTime, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)

    section: Mapped["Section"] = relationship("Section", back_populates="warnings")
