from datetime import date, datetime
from sqlalchemy import (
    Column,
    Integer,
    String,
    Text,
    Date,
    DateTime,
    ForeignKey,
    Boolean,
    Float,
    Enum,
)
from sqlalchemy.orm import relationship, DeclarativeBase
import enum


class Base(DeclarativeBase):
    pass


class ApprovalLevel(enum.Enum):
    A = "A"
    B = "B"
    C = "C"


class ApprovalStatus(enum.Enum):
    PENDING_LEVEL1 = "pending_level1"
    APPROVED_LEVEL1 = "approved_level1"
    PENDING_LEVEL2 = "pending_level2"
    APPROVED_LEVEL2 = "approved_level2"
    PENDING_LEVEL3 = "pending_level3"
    APPROVED = "approved"
    REJECTED = "rejected"


class Enterprise(Base):
    __tablename__ = "enterprises"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(255), nullable=False, index=True)
    address = Column(String(500))
    contact_person = Column(String(100))
    contact_phone = Column(String(50))
    license_number = Column(String(100))
    created_at = Column(DateTime, default=datetime.now)
    updated_at = Column(DateTime, default=datetime.now, onupdate=datetime.now)

    chemicals = relationship("Chemical", back_populates="enterprise")
    approvals = relationship("Approval", back_populates="enterprise")
    ledgers = relationship("Ledger", back_populates="enterprise")
    plans = relationship("EmergencyPlan", back_populates="enterprise")
    accidents = relationship("AccidentRecord", back_populates="enterprise")
    drills = relationship("Drill", back_populates="enterprise")


class Chemical(Base):
    __tablename__ = "chemicals"

    id = Column(Integer, primary_key=True, index=True)
    enterprise_id = Column(Integer, ForeignKey("enterprises.id"), nullable=False)
    name = Column(String(255), nullable=False, index=True)
    category = Column(String(100), index=True)
    cas_number = Column(String(50))
    hazard_level = Column(String(50))
    description = Column(Text)
    created_at = Column(DateTime, default=datetime.now)
    updated_at = Column(DateTime, default=datetime.now, onupdate=datetime.now)

    enterprise = relationship("Enterprise", back_populates="chemicals")
    approvals = relationship("Approval", back_populates="chemical")
    ledgers = relationship("Ledger", back_populates="chemical")


class Approval(Base):
    __tablename__ = "approvals"

    id = Column(Integer, primary_key=True, index=True)
    enterprise_id = Column(Integer, ForeignKey("enterprises.id"), nullable=False)
    chemical_id = Column(Integer, ForeignKey("chemicals.id"), nullable=False)
    approval_level = Column(Enum(ApprovalLevel), nullable=False)
    status = Column(Enum(ApprovalStatus), default=ApprovalStatus.PENDING_LEVEL1)
    max_quantity = Column(Float, nullable=False)
    issue_date = Column(Date, nullable=True)
    expiry_date = Column(Date, nullable=True)
    level1_approver = Column(String(100))
    level1_comment = Column(Text)
    level1_approved_at = Column(DateTime)
    level2_approver = Column(String(100))
    level2_comment = Column(Text)
    level2_approved_at = Column(DateTime)
    level3_approver = Column(String(100))
    level3_comment = Column(Text)
    level3_approved_at = Column(DateTime)
    created_at = Column(DateTime, default=datetime.now)
    updated_at = Column(DateTime, default=datetime.now, onupdate=datetime.now)

    enterprise = relationship("Enterprise", back_populates="approvals")
    chemical = relationship("Chemical", back_populates="approvals")


class EmergencyPlan(Base):
    __tablename__ = "emergency_plans"

    id = Column(Integer, primary_key=True, index=True)
    enterprise_id = Column(Integer, ForeignKey("enterprises.id"), nullable=False)
    title = Column(String(255), nullable=False)
    version = Column(String(50))
    content = Column(Text, nullable=False)
    created_by = Column(String(100))
    created_at = Column(DateTime, default=datetime.now)
    updated_at = Column(DateTime, default=datetime.now, onupdate=datetime.now)

    enterprise = relationship("Enterprise", back_populates="plans")
    drills = relationship("Drill", back_populates="plan")


class AccidentRecord(Base):
    __tablename__ = "accident_records"

    id = Column(Integer, primary_key=True, index=True)
    enterprise_id = Column(Integer, ForeignKey("enterprises.id"), nullable=False)
    title = Column(String(255), nullable=False)
    accident_date = Column(DateTime, nullable=False)
    location = Column(String(500))
    description = Column(Text, nullable=False)
    severity = Column(String(50))
    casualties = Column(Integer, default=0)
    financial_loss = Column(Float, default=0.0)
    disposition = Column(Text)
    status = Column(String(50), default="processing")
    created_at = Column(DateTime, default=datetime.now)
    updated_at = Column(DateTime, default=datetime.now, onupdate=datetime.now)

    enterprise = relationship("Enterprise", back_populates="accidents")


class AuditLog(Base):
    __tablename__ = "audit_logs"

    id = Column(Integer, primary_key=True, index=True)
    action = Column(String(100), nullable=False)
    entity_type = Column(String(100))
    entity_id = Column(Integer)
    description = Column(Text, nullable=False)
    user = Column(String(100))
    timestamp = Column(DateTime, default=datetime.now, nullable=False)
    details = Column(Text)

    __table_args__ = ()


class Ledger(Base):
    __tablename__ = "ledgers"

    id = Column(Integer, primary_key=True, index=True)
    enterprise_id = Column(Integer, ForeignKey("enterprises.id"), nullable=False)
    chemical_id = Column(Integer, ForeignKey("chemicals.id"), nullable=False)
    transaction_type = Column(String(50), nullable=False)
    quantity = Column(Float, nullable=False)
    balance = Column(Float, nullable=False)
    unit = Column(String(50), default="kg")
    transaction_date = Column(Date, nullable=False)
    operator = Column(String(100))
    description = Column(Text)
    created_at = Column(DateTime, default=datetime.now)

    enterprise = relationship("Enterprise", back_populates="ledgers")
    chemical = relationship("Chemical", back_populates="ledgers")


class LedgerCheck(Base):
    __tablename__ = "ledger_checks"

    id = Column(Integer, primary_key=True, index=True)
    check_date = Column(Date, nullable=False)
    enterprise_id = Column(Integer, ForeignKey("enterprises.id"))
    chemical_id = Column(Integer, ForeignKey("chemicals.id"))
    is_abnormal = Column(Boolean, default=False)
    anomaly_type = Column(String(100))
    anomaly_description = Column(Text)
    created_at = Column(DateTime, default=datetime.now)


class Drill(Base):
    __tablename__ = "drills"

    id = Column(Integer, primary_key=True, index=True)
    enterprise_id = Column(Integer, ForeignKey("enterprises.id"), nullable=False)
    plan_id = Column(Integer, ForeignKey("emergency_plans.id"))
    drill_date = Column(Date, nullable=False)
    title = Column(String(255), nullable=False)
    description = Column(Text)
    participants = Column(Integer)
    duration_hours = Column(Float)
    evaluation_result = Column(String(50))
    evaluation_details = Column(Text)
    created_at = Column(DateTime, default=datetime.now)

    enterprise = relationship("Enterprise", back_populates="drills")
    plan = relationship("EmergencyPlan", back_populates="drills")


class Reminder(Base):
    __tablename__ = "reminders"

    id = Column(Integer, primary_key=True, index=True)
    reminder_type = Column(String(100), nullable=False)
    target_id = Column(Integer, nullable=False)
    target_type = Column(String(100), nullable=False)
    message = Column(Text, nullable=False)
    due_date = Column(Date)
    is_completed = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.now)
