from datetime import date
from typing import List, Optional
from enum import Enum
from sqlalchemy import Column, Integer, String, Date, Boolean, Enum as SAEnum, ForeignKey, Text
from sqlalchemy.orm import relationship
from sqlalchemy.ext.declarative import declarative_base

Base = declarative_base()


class CertificateType(str, Enum):
    MINING_LICENSE = "mining_license"
    SAFETY_LICENSE = "safety_license"
    ENV_APPROVAL = "env_approval"
    OTHER = "other"


class TodoType(str, Enum):
    CERTIFICATE_EXPIRY = "certificate_expiry"
    ANNUAL_INSPECTION = "annual_inspection"
    COMPLIANCE_RECTIFICATION = "compliance_rectification"


class TodoStatus(str, Enum):
    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"


class CertificateStatus(str, Enum):
    ACTIVE = "active"
    CANCELLED = "cancelled"
    EXPIRED = "expired"
    ABOUT_TO_EXPIRE = "about_to_expire"


class Certificate(Base):
    __tablename__ = "certificates"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(200), nullable=False)
    type = Column(SAEnum(CertificateType), nullable=False)
    number = Column(String(100), unique=True, nullable=False)
    issuing_date = Column(Date, nullable=False)
    expiry_date = Column(Date, nullable=False)
    is_cancelled = Column(Boolean, default=False, nullable=False)
    remarks = Column(Text, nullable=True)

    inspections = relationship("AnnualInspection", back_populates="certificate", cascade="all, delete-orphan")
    compliance_checks = relationship("ComplianceCheck", back_populates="certificate", cascade="all, delete-orphan")
    todos = relationship("Todo", back_populates="certificate", cascade="all, delete-orphan")

    @property
    def status(self) -> CertificateStatus:
        if self.is_cancelled:
            return CertificateStatus.CANCELLED
        today = date.today()
        delta = self.expiry_date - today
        if delta.days < 0:
            return CertificateStatus.EXPIRED
        elif delta.days <= 90:
            return CertificateStatus.ABOUT_TO_EXPIRE
        return CertificateStatus.ACTIVE

    @property
    def is_safety_related(self) -> bool:
        return self.type == CertificateType.SAFETY_LICENSE


class AnnualInspection(Base):
    __tablename__ = "annual_inspections"

    id = Column(Integer, primary_key=True, index=True)
    certificate_id = Column(Integer, ForeignKey("certificates.id"), nullable=False)
    year = Column(Integer, nullable=False)
    inspection_date = Column(Date, nullable=False)
    result = Column(Boolean, nullable=False)
    remarks = Column(Text, nullable=True)

    certificate = relationship("Certificate", back_populates="inspections")


class ComplianceCheck(Base):
    __tablename__ = "compliance_checks"

    id = Column(Integer, primary_key=True, index=True)
    certificate_id = Column(Integer, ForeignKey("certificates.id"), nullable=False)
    check_date = Column(Date, nullable=False)
    check_items = Column(Text, nullable=False)
    is_compliant = Column(Boolean, nullable=False)
    has_safety_issues = Column(Boolean, default=False, nullable=False)
    remarks = Column(Text, nullable=True)

    certificate = relationship("Certificate", back_populates="compliance_checks")
    todos = relationship("Todo", back_populates="compliance_check", cascade="all, delete-orphan")


class Todo(Base):
    __tablename__ = "todos"

    id = Column(Integer, primary_key=True, index=True)
    type = Column(SAEnum(TodoType), nullable=False)
    certificate_id = Column(Integer, ForeignKey("certificates.id"), nullable=True)
    compliance_check_id = Column(Integer, ForeignKey("compliance_checks.id"), nullable=True)
    due_date = Column(Date, nullable=False)
    status = Column(SAEnum(TodoStatus), default=TodoStatus.PENDING, nullable=False)
    description = Column(Text, nullable=False)
    created_at = Column(Date, default=date.today, nullable=False)

    certificate = relationship("Certificate", back_populates="todos")
    compliance_check = relationship("ComplianceCheck", back_populates="todos")
