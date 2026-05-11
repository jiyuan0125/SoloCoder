from sqlalchemy import (
    Column, Integer, String, Text, DateTime, ForeignKey,
    Date, Float, Boolean, Enum
)
from sqlalchemy.orm import relationship
from datetime import datetime
from enum import Enum as PyEnum
from .database import Base


class LicenseStatus(PyEnum):
    DRAFT = "draft"
    SUBMITTED = "submitted"
    ACCEPTED = "accepted"
    INSPECTED = "inspected"
    PUBLISHED = "published"
    ISSUED = "issued"
    SUSPENDED = "suspended"
    REVOKED = "revoked"
    EXPIRED = "expired"


class PenaltyType(PyEnum):
    WARNING = "warning"
    FINE = "fine"
    SUSPEND = "suspend"
    REVOKE = "revoke"


class PenaltyStatus(PyEnum):
    PENDING = "pending"
    EXECUTED = "executed"


class RiverImportance(PyEnum):
    IMPORTANT = "important"
    NORMAL = "normal"


class AuditLog(Base):
    __tablename__ = "audit_logs"

    id = Column(Integer, primary_key=True, index=True)
    action = Column(String(100), nullable=False)
    entity_type = Column(String(50), nullable=False)
    entity_id = Column(Integer, nullable=False)
    details = Column(Text)
    operator = Column(String(100), nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)


class License(Base):
    __tablename__ = "licenses"

    id = Column(Integer, primary_key=True, index=True)
    license_number = Column(String(50), unique=True, index=True)
    company_name = Column(String(200), nullable=False)
    company_identifier = Column(String(100), nullable=False)
    river_section = Column(String(200), nullable=False)
    mining_location = Column(Text)
    annual_quota = Column(Float, nullable=False)
    start_date = Column(Date, nullable=False)
    end_date = Column(Date, nullable=False)
    status = Column(Enum(LicenseStatus), default=LicenseStatus.DRAFT)
    submit_time = Column(DateTime)
    accept_time = Column(DateTime)
    inspect_time = Column(DateTime)
    inspect_result = Column(Text)
    publish_time = Column(DateTime)
    publish_end_time = Column(DateTime)
    issue_time = Column(DateTime)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    mining_reports = relationship("MiningReport", back_populates="license")
    weighbridge_records = relationship("WeighbridgeRecord", back_populates="license")
    law_enforcements = relationship("LawEnforcement", back_populates="license")


class MiningReport(Base):
    __tablename__ = "mining_reports"

    id = Column(Integer, primary_key=True, index=True)
    license_id = Column(Integer, ForeignKey("licenses.id"), nullable=False)
    report_date = Column(Date, nullable=False)
    start_time = Column(DateTime, nullable=False)
    end_time = Column(DateTime, nullable=False)
    mining_location = Column(Text)
    expected_volume = Column(Float)
    operator_name = Column(String(100))
    status = Column(String(20), default="submitted")
    created_at = Column(DateTime, default=datetime.utcnow)

    license = relationship("License", back_populates="mining_reports")


class WeighbridgeRecord(Base):
    __tablename__ = "weighbridge_records"

    id = Column(Integer, primary_key=True, index=True)
    license_id = Column(Integer, ForeignKey("licenses.id"), nullable=False)
    record_time = Column(DateTime, nullable=False)
    vehicle_plate = Column(String(50))
    gross_weight = Column(Float, nullable=False)
    tare_weight = Column(Float, nullable=False)
    net_weight = Column(Float, nullable=False)
    material_type = Column(String(50))
    created_at = Column(DateTime, default=datetime.utcnow)

    license = relationship("License", back_populates="weighbridge_records")


class MiningStatistics(Base):
    __tablename__ = "mining_statistics"

    id = Column(Integer, primary_key=True, index=True)
    license_id = Column(Integer, ForeignKey("licenses.id"), nullable=False)
    year = Column(Integer, nullable=False)
    month = Column(Integer, nullable=False)
    total_volume = Column(Float, default=0.0)
    warning_sent = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)


class LawEnforcement(Base):
    __tablename__ = "law_enforcements"

    id = Column(Integer, primary_key=True, index=True)
    license_id = Column(Integer, ForeignKey("licenses.id"))
    company_name = Column(String(200), nullable=False)
    violation_time = Column(DateTime, nullable=False)
    violation_location = Column(String(200), nullable=False)
    violation_description = Column(Text, nullable=False)
    penalty_type = Column(Enum(PenaltyType), nullable=False)
    penalty_amount = Column(Float)
    penalty_status = Column(Enum(PenaltyStatus), default=PenaltyStatus.PENDING)
    penalty_time = Column(DateTime)
    created_at = Column(DateTime, default=datetime.utcnow)

    license = relationship("License", back_populates="law_enforcements")


class RiverSection(Base):
    __tablename__ = "river_sections"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(200), nullable=False, unique=True)
    importance = Column(Enum(RiverImportance), default=RiverImportance.NORMAL)
    description = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)


class PatrolTask(Base):
    __tablename__ = "patrol_tasks"

    id = Column(Integer, primary_key=True, index=True)
    river_section_id = Column(Integer, ForeignKey("river_sections.id"), nullable=False)
    task_date = Column(Date, nullable=False)
    assigned_to = Column(String(100))
    status = Column(String(20), default="pending")
    result = Column(Text)
    completed_at = Column(DateTime)
    created_at = Column(DateTime, default=datetime.utcnow)


class MonthlyReport(Base):
    __tablename__ = "monthly_reports"

    id = Column(Integer, primary_key=True, index=True)
    report_year = Column(Integer, nullable=False)
    report_month = Column(Integer, nullable=False)
    content = Column(Text, nullable=False)
    status = Column(String(20), default="pending")
    reviewer = Column(String(100))
    review_time = Column(DateTime)
    review_comment = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)
