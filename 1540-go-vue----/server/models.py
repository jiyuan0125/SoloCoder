from datetime import datetime, date
from enum import Enum as PyEnum
from sqlalchemy import Column, Integer, String, DateTime, Date, Float, ForeignKey, Text, Enum
from sqlalchemy.orm import relationship

from .database import Base


class SourceStatus(PyEnum):
    IN_STOCK = "入库"
    IN_USE = "在用"
    IN_STORAGE = "暂存"
    RETIRED = "退役"


class TodoStatus(PyEnum):
    PENDING = "待处理"
    PROCESSING = "处理中"
    RESOLVED = "已解决"


class ApprovalStatus(PyEnum):
    DRAFT = "草稿"
    SUBMITTED = "待审批"
    APPROVED = "已通过"
    REJECTED = "已驳回"


class AlertLevel(PyEnum):
    YELLOW = "黄色提醒"
    ORANGE = "橙色警告"
    RED = "红色严重警告"


class Unit(Base):
    __tablename__ = "units"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(255), nullable=False, unique=True, index=True)
    address = Column(String(500), nullable=False)
    contact_person = Column(String(100))
    contact_phone = Column(String(50))
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    sources = relationship("RadiationSource", back_populates="unit")


class RadiationSource(Base):
    __tablename__ = "radiation_sources"

    id = Column(Integer, primary_key=True, index=True)
    source_code = Column(String(12), nullable=False, unique=True, index=True)
    name = Column(String(255), nullable=False)
    source_type = Column(String(100), nullable=False)
    source_class = Column(Integer, nullable=False)
    activity = Column(Float, nullable=False)
    unit_id = Column(Integer, ForeignKey("units.id"), nullable=False, index=True)
    status = Column(Enum(SourceStatus), default=SourceStatus.IN_STOCK, nullable=False)
    manufacture_date = Column(Date, nullable=False)
    expire_date = Column(Date, nullable=False)
    storage_location = Column(String(255))
    description = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    unit = relationship("Unit", back_populates="sources")
    inspections = relationship("InspectionRecord", back_populates="source", order_by="InspectionRecord.inspection_time.desc()")
    approvals = relationship("RetirementApproval", back_populates="source")


class InspectionRecord(Base):
    __tablename__ = "inspection_records"

    id = Column(Integer, primary_key=True, index=True)
    source_id = Column(Integer, ForeignKey("radiation_sources.id"), nullable=False, index=True)
    inspection_time = Column(DateTime, default=datetime.utcnow, index=True)
    inspector = Column(String(100), nullable=False)
    radiation_dose = Column(Float, nullable=False)
    is_abnormal = Column(Integer, default=0)
    result = Column(String(50), default="正常")
    remarks = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)

    source = relationship("RadiationSource", back_populates="inspections")
    todo = relationship("EmergencyTodo", back_populates="inspection", uselist=False)


class EmergencyTodo(Base):
    __tablename__ = "emergency_todos"

    id = Column(Integer, primary_key=True, index=True)
    inspection_id = Column(Integer, ForeignKey("inspection_records.id"), nullable=False, unique=True, index=True)
    source_id = Column(Integer, ForeignKey("radiation_sources.id"), nullable=False, index=True)
    source_code = Column(String(12), nullable=False)
    dose_value = Column(Float, nullable=False)
    status = Column(Enum(TodoStatus), default=TodoStatus.PENDING, nullable=False)
    handler = Column(String(100))
    handle_time = Column(DateTime)
    handle_result = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)

    inspection = relationship("InspectionRecord", back_populates="todo")


class RetirementApproval(Base):
    __tablename__ = "retirement_approvals"

    id = Column(Integer, primary_key=True, index=True)
    source_id = Column(Integer, ForeignKey("radiation_sources.id"), nullable=False, index=True)
    source_code = Column(String(12), nullable=False)
    plan_content = Column(Text, nullable=False)
    applicant = Column(String(100), nullable=False)
    apply_time = Column(DateTime, default=datetime.utcnow)
    status = Column(Enum(ApprovalStatus), default=ApprovalStatus.DRAFT, nullable=False)
    approver = Column(String(100))
    approval_time = Column(DateTime)
    approval_remarks = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    source = relationship("RadiationSource", back_populates="approvals")


class AuditLog(Base):
    __tablename__ = "audit_logs"

    id = Column(Integer, primary_key=True, index=True)
    operation_type = Column(String(100), nullable=False, index=True)
    operator = Column(String(100), nullable=False, index=True)
    target_type = Column(String(50))
    target_id = Column(Integer)
    description = Column(Text)
    operation_time = Column(DateTime, default=datetime.utcnow, index=True)
    created_at = Column(DateTime, default=datetime.utcnow)
