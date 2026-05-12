from datetime import datetime, date
from enum import Enum as PyEnum
from sqlalchemy import Column, Integer, String, Date, DateTime, Float, ForeignKey, Text, Enum
from sqlalchemy.orm import relationship
from server.database import Base


class LicenseType(PyEnum):
    CABOTAGE = "cabotage"
    CARGO = "cargo"
    PASSENGER = "passenger"
    SPECIAL = "special"


class LicenseStatus(PyEnum):
    PENDING = "pending"
    APPROVED = "approved"
    REJECTED = "rejected"
    EXPIRED = "expired"


class InspectionResult(PyEnum):
    PASSED = "passed"
    FAILED = "failed"


class RectificationStatus(PyEnum):
    CREATED = "created"
    RECTIFYING = "rectifying"
    RECHECKING = "rechecking"
    CLOSED = "closed"


class PenaltyStatus(PyEnum):
    PENDING = "pending"
    PAID = "paid"
    OVERDUE = "overdue"


class ActionType(PyEnum):
    CREATE = "create"
    UPDATE = "update"
    DELETE = "delete"


class Ship(Base):
    __tablename__ = "ships"

    id = Column(Integer, primary_key=True, index=True)
    imo_number = Column(String(20), unique=True, index=True, nullable=True)
    name = Column(String(100), nullable=True)
    registration_port = Column(String(100), nullable=True)
    gross_tonnage = Column(Float, nullable=True)
    length = Column(Float, nullable=True)
    width = Column(Float, nullable=True)
    owner_name = Column(String(100), nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    licenses = relationship("License", back_populates="ship")
    inspections = relationship("Inspection", back_populates="ship")
    penalties = relationship("Penalty", back_populates="ship")
    rectifications = relationship("Rectification", back_populates="ship")


class License(Base):
    __tablename__ = "licenses"

    id = Column(Integer, primary_key=True, index=True)
    ship_id = Column(Integer, ForeignKey("ships.id"), nullable=True)
    license_type = Column(Enum(LicenseType), nullable=False)
    application_date = Column(Date, nullable=False)
    expected_effective_date = Column(Date, nullable=False)
    effective_date = Column(Date, nullable=True)
    expiration_date = Column(Date, nullable=True)
    applicant = Column(String(100), nullable=False)
    status = Column(Enum(LicenseStatus), default=LicenseStatus.PENDING)
    rejection_reason = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    ship = relationship("Ship", back_populates="licenses")


class Inspection(Base):
    __tablename__ = "inspections"

    id = Column(Integer, primary_key=True, index=True)
    ship_id = Column(Integer, ForeignKey("ships.id"), nullable=True)
    ship_temp_identifier = Column(String(100), nullable=True)
    inspection_date = Column(Date, nullable=False)
    inspector = Column(String(100), nullable=False)
    location = Column(String(200), nullable=True)
    result = Column(Enum(InspectionResult), nullable=False)
    findings = Column(Text, nullable=True)
    remarks = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    ship = relationship("Ship", back_populates="inspections")
    rectification = relationship("Rectification", back_populates="inspection", uselist=False)
    penalty = relationship("Penalty", back_populates="inspection", uselist=False)


class Rectification(Base):
    __tablename__ = "rectifications"

    id = Column(Integer, primary_key=True, index=True)
    ship_id = Column(Integer, ForeignKey("ships.id"), nullable=True)
    ship_temp_identifier = Column(String(100), nullable=True)
    inspection_id = Column(Integer, ForeignKey("inspections.id"), nullable=True)
    status = Column(Enum(RectificationStatus), default=RectificationStatus.CREATED)
    requirement = Column(Text, nullable=False)
    deadline = Column(Date, nullable=False)
    rectification_details = Column(Text, nullable=True)
    rectification_date = Column(Date, nullable=True)
    recheck_result = Column(Text, nullable=True)
    recheck_date = Column(Date, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    ship = relationship("Ship", back_populates="rectifications")
    inspection = relationship("Inspection", back_populates="rectification")


class Penalty(Base):
    __tablename__ = "penalties"

    id = Column(Integer, primary_key=True, index=True)
    ship_id = Column(Integer, ForeignKey("ships.id"), nullable=True)
    ship_temp_identifier = Column(String(100), nullable=True)
    inspection_id = Column(Integer, ForeignKey("inspections.id"), nullable=True)
    violation = Column(Text, nullable=False)
    fine_amount = Column(Float, nullable=False)
    issue_date = Column(Date, nullable=False)
    due_date = Column(Date, nullable=False)
    payment_date = Column(Date, nullable=True)
    status = Column(Enum(PenaltyStatus), default=PenaltyStatus.PENDING)
    remarks = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    ship = relationship("Ship", back_populates="penalties")
    inspection = relationship("Inspection", back_populates="penalty")


class AuditLog(Base):
    __tablename__ = "audit_logs"

    id = Column(Integer, primary_key=True, index=True)
    entity_type = Column(String(50), nullable=False)
    entity_id = Column(Integer, nullable=False)
    action = Column(Enum(ActionType), nullable=False)
    details = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
