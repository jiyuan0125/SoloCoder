from sqlalchemy import Column, Integer, String, Float, DateTime, Boolean, Text, ForeignKey, Enum
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func
from .database import Base
import enum


class NavigationStatus(str, enum.Enum):
    NORMAL = "normal"
    RESTRICTED = "restricted"


class BeaconStatus(str, enum.Enum):
    NORMAL = "normal"
    FAULTY = "faulty"
    MISSING = "missing"


class DredgingStatus(str, enum.Enum):
    PLANNED = "planned"
    UNDERWAY = "underway"
    COMPLETED = "completed"
    CANCELLED = "cancelled"


class DraftDeclarationStatus(str, enum.Enum):
    PENDING = "pending"
    APPROVED = "approved"
    REJECTED = "rejected"
    PASSED = "passed"


class AlertSeverity(str, enum.Enum):
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"


class TodoPriority(str, enum.Enum):
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    URGENT = "urgent"


class Section(Base):
    __tablename__ = "sections"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), unique=True, index=True, nullable=False)
    design_depth = Column(Float, nullable=False)
    current_measured_depth = Column(Float, nullable=True)
    navigation_status = Column(String(20), default=NavigationStatus.NORMAL, nullable=False)
    beacon_status_normal = Column(Boolean, default=True, nullable=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now())

    beacons = relationship("Beacon", back_populates="section", cascade="all, delete-orphan")
    depth_records = relationship("DepthRecord", back_populates="section", cascade="all, delete-orphan")
    dredging_plans = relationship("DredgingPlan", back_populates="section", cascade="all, delete-orphan")
    draft_declarations = relationship("DraftDeclaration", back_populates="section", cascade="all, delete-orphan")
    alerts = relationship("Alert", back_populates="section", cascade="all, delete-orphan")
    todos = relationship("Todo", back_populates="section", cascade="all, delete-orphan")


class Beacon(Base):
    __tablename__ = "beacons"

    id = Column(Integer, primary_key=True, index=True)
    section_id = Column(Integer, ForeignKey("sections.id"), nullable=False)
    name = Column(String(100), nullable=False)
    code = Column(String(50), unique=True, index=True, nullable=False)
    status = Column(String(20), default=BeaconStatus.NORMAL, nullable=False)
    description = Column(Text, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now())

    section = relationship("Section", back_populates="beacons")


class DepthRecord(Base):
    __tablename__ = "depth_records"

    id = Column(Integer, primary_key=True, index=True)
    section_id = Column(Integer, ForeignKey("sections.id"), nullable=False)
    measured_depth = Column(Float, nullable=False)
    recorded_at = Column(DateTime(timezone=True), nullable=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now())

    section = relationship("Section", back_populates="depth_records")


class DredgingPlan(Base):
    __tablename__ = "dredging_plans"

    id = Column(Integer, primary_key=True, index=True)
    section_id = Column(Integer, ForeignKey("sections.id"), nullable=False)
    target_depth = Column(Float, nullable=False)
    status = Column(String(20), default=DredgingStatus.PLANNED, nullable=False)
    start_date = Column(DateTime(timezone=True), nullable=True)
    end_date = Column(DateTime(timezone=True), nullable=True)
    description = Column(Text, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now())

    section = relationship("Section", back_populates="dredging_plans")
    draft_declarations = relationship("DraftDeclaration", back_populates="dredging_plan")


class DraftDeclaration(Base):
    __tablename__ = "draft_declarations"

    id = Column(Integer, primary_key=True, index=True)
    section_id = Column(Integer, ForeignKey("sections.id"), nullable=False)
    dredging_plan_id = Column(Integer, ForeignKey("dredging_plans.id"), nullable=True)
    vessel_name = Column(String(100), nullable=False)
    declared_draft = Column(Float, nullable=False)
    safety_margin = Column(Float, default=0.3, nullable=False)
    status = Column(String(20), default=DraftDeclarationStatus.PENDING, nullable=False)
    validation_result = Column(Text, nullable=True)
    approved_at = Column(DateTime(timezone=True), nullable=True)
    passed_at = Column(DateTime(timezone=True), nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now())

    section = relationship("Section", back_populates="draft_declarations")
    dredging_plan = relationship("DredgingPlan", back_populates="draft_declarations")


class Alert(Base):
    __tablename__ = "alerts"

    id = Column(Integer, primary_key=True, index=True)
    section_id = Column(Integer, ForeignKey("sections.id"), nullable=False)
    type = Column(String(50), nullable=False)
    severity = Column(String(20), default=AlertSeverity.MEDIUM, nullable=False)
    message = Column(Text, nullable=False)
    is_active = Column(Boolean, default=True, nullable=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    resolved_at = Column(DateTime(timezone=True), nullable=True)

    section = relationship("Section", back_populates="alerts")


class Todo(Base):
    __tablename__ = "todos"

    id = Column(Integer, primary_key=True, index=True)
    section_id = Column(Integer, ForeignKey("sections.id"), nullable=True)
    title = Column(String(200), nullable=False)
    description = Column(Text, nullable=True)
    priority = Column(String(20), default=TodoPriority.MEDIUM, nullable=False)
    deadline = Column(DateTime(timezone=True), nullable=True)
    is_completed = Column(Boolean, default=False, nullable=False)
    completed_at = Column(DateTime(timezone=True), nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now())

    section = relationship("Section", back_populates="todos")
