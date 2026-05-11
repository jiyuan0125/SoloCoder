from sqlalchemy import (
    Column, Integer, String, Date, Float, Text, ForeignKey,
    DateTime, Boolean, Enum as SQLEnum
)
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func
from enum import Enum
from .database import Base


class FishermanStatus(str, Enum):
    ACTIVE = "active"
    SUSPENDED = "suspended"
    REVOKED = "revoked"


class LicenseStatus(str, Enum):
    PENDING = "pending"
    APPROVED = "approved"
    REJECTED = "rejected"
    SUSPENDED = "suspended"
    REVOKED = "revoked"
    EXPIRED = "expired"


class LicenseType(str, Enum):
    FISHING = "fishing"
    AQUACULTURE = "aquaculture"


class ViolationType(str, Enum):
    ILLEGAL_FISHING = "illegal_fishing"
    CLOSED_SEASON = "closed_season"
    OVERSIZE_FISH = "oversize_fish"
    ILLEGAL_EQUIPMENT = "illegal_equipment"
    NO_LICENSE = "no_license"
    OTHER = "other"


class Severity(str, Enum):
    MILD = "mild"
    MODERATE = "moderate"
    SEVERE = "severe"
    CRITICAL = "critical"


class PenaltyType(str, Enum):
    WARNING = "warning"
    FINE = "fine"
    SUSPEND_LICENSE = "suspend_license"
    REVOKE_LICENSE = "revoke_license"
    CRIMINAL_REFERRAL = "criminal_referral"


class CaseStatus(str, Enum):
    FILED = "filed"
    INVESTIGATING = "investigating"
    DECIDED = "decided"
    UNDER_APPEAL = "under_appeal"
    EXECUTING = "executing"
    CLOSED = "closed"


class TodoType(str, Enum):
    CLOSED_SEASON_NOTICE = "closed_season_notice"
    SUSPECTED_VIOLATION = "suspected_violation"


class TodoStatus(str, Enum):
    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"


class Fisherman(Base):
    __tablename__ = "fishermen"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), nullable=False)
    id_card = Column(String(50), unique=True, index=True, nullable=False)
    phone = Column(String(20))
    address = Column(String(255))
    vessel_name = Column(String(100))
    vessel_registration = Column(String(50), unique=True, index=True)
    status = Column(SQLEnum(FishermanStatus), default=FishermanStatus.ACTIVE)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())

    licenses = relationship("FishingLicense", back_populates="fisherman")
    violations = relationship("ViolationCase", back_populates="fisherman")
    port_records = relationship("PortRecord", back_populates="fisherman")


class FishingLicense(Base):
    __tablename__ = "fishing_licenses"

    id = Column(Integer, primary_key=True, index=True)
    fisherman_id = Column(Integer, ForeignKey("fishermen.id"), nullable=False)
    license_number = Column(String(50), unique=True, index=True)
    license_type = Column(SQLEnum(LicenseType), nullable=False)
    valid_from = Column(Date, nullable=False)
    valid_to = Column(Date, nullable=False)
    fishing_area = Column(String(255))
    allowed_gear = Column(String(255))
    status = Column(SQLEnum(LicenseStatus), default=LicenseStatus.PENDING)
    application_date = Column(Date, server_default=func.current_date())
    approval_date = Column(Date)
    rejection_reason = Column(Text)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())

    fisherman = relationship("Fisherman", back_populates="licenses")
    violations = relationship("ViolationCase", back_populates="license")


class ViolationCase(Base):
    __tablename__ = "violation_cases"

    id = Column(Integer, primary_key=True, index=True)
    case_number = Column(String(50), unique=True, index=True, nullable=False)
    fisherman_id = Column(Integer, ForeignKey("fishermen.id"), nullable=False)
    license_id = Column(Integer, ForeignKey("fishing_licenses.id"))
    violation_type = Column(SQLEnum(ViolationType), nullable=False)
    severity = Column(SQLEnum(Severity), nullable=False)
    description = Column(Text)
    location = Column(String(255))
    violation_date = Column(Date, nullable=False)
    status = Column(SQLEnum(CaseStatus), default=CaseStatus.FILED)
    penalty_type = Column(SQLEnum(PenaltyType))
    fine_amount = Column(Float)
    suspension_days = Column(Integer)
    decision_reason = Column(Text)
    decision_date = Column(Date)
    appeal_deadline = Column(Date)
    is_closed = Column(Boolean, default=False)
    closed_date = Column(Date)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())

    fisherman = relationship("Fisherman", back_populates="violations")
    license = relationship("FishingLicense", back_populates="violations")
    evidences = relationship("Evidence", back_populates="case", cascade="all, delete-orphan")
    appeals = relationship("Appeal", back_populates="case", cascade="all, delete-orphan")


class Evidence(Base):
    __tablename__ = "evidences"

    id = Column(Integer, primary_key=True, index=True)
    case_id = Column(Integer, ForeignKey("violation_cases.id"), nullable=False)
    evidence_type = Column(String(100))
    description = Column(Text)
    file_path = Column(String(255))
    submitted_at = Column(DateTime(timezone=True), server_default=func.now())

    case = relationship("ViolationCase", back_populates="evidences")


class Appeal(Base):
    __tablename__ = "appeals"

    id = Column(Integer, primary_key=True, index=True)
    case_id = Column(Integer, ForeignKey("violation_cases.id"), nullable=False)
    appeal_reason = Column(Text)
    appeal_date = Column(Date, server_default=func.current_date())
    decision = Column(String(100))
    decision_date = Column(Date)
    decision_reason = Column(Text)

    case = relationship("ViolationCase", back_populates="appeals")


class Todo(Base):
    __tablename__ = "todos"

    id = Column(Integer, primary_key=True, index=True)
    title = Column(String(255), nullable=False)
    description = Column(Text)
    todo_type = Column(SQLEnum(TodoType), nullable=False)
    status = Column(SQLEnum(TodoStatus), default=TodoStatus.PENDING)
    related_fisherman_id = Column(Integer, ForeignKey("fishermen.id"))
    related_case_id = Column(Integer, ForeignKey("violation_cases.id"))
    due_date = Column(Date)
    location = Column(String(255))
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    completed_at = Column(DateTime(timezone=True))


class SeedFarm(Base):
    __tablename__ = "seed_farms"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(255), nullable=False)
    address = Column(String(255))
    contact_person = Column(String(100))
    phone = Column(String(20))
    license_number = Column(String(100))


class ReleaseActivity(Base):
    __tablename__ = "release_activities"

    id = Column(Integer, primary_key=True, index=True)
    activity_code = Column(String(50), unique=True, index=True, nullable=False)
    activity_name = Column(String(255), nullable=False)
    seed_farm_id = Column(Integer, ForeignKey("seed_farms.id"))
    species = Column(String(100), nullable=False)
    quantity = Column(Integer, nullable=False)
    unit = Column(String(50), default="tail")
    release_date = Column(Date, nullable=False)
    release_location = Column(String(255), nullable=False)
    water_area = Column(String(255))
    coordinator = Column(String(100))
    remarks = Column(Text)
    created_at = Column(DateTime(timezone=True), server_default=func.now())

    seed_farm = relationship("SeedFarm")


class Port(Base):
    __tablename__ = "ports"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(255), nullable=False)
    location = Column(String(255))
    berth_count = Column(Integer, nullable=False)
    max_tonnage = Column(Float, nullable=False)
    manager = Column(String(100))
    phone = Column(String(20))


class PortRecord(Base):
    __tablename__ = "port_records"

    id = Column(Integer, primary_key=True, index=True)
    port_id = Column(Integer, ForeignKey("ports.id"), nullable=False)
    fisherman_id = Column(Integer, ForeignKey("fishermen.id"), nullable=False)
    record_type = Column(String(20), nullable=False)
    vessel_name = Column(String(100))
    vessel_registration = Column(String(50))
    record_time = Column(DateTime(timezone=True), server_default=func.now())
    destination = Column(String(255))
    cargo = Column(String(255))
    remarks = Column(Text)

    fisherman = relationship("Fisherman", back_populates="port_records")
    port = relationship("Port")
