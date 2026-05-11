from datetime import datetime, date
from sqlalchemy import Column, Integer, String, Float, DateTime, Date, ForeignKey, Text, Enum, Boolean
from sqlalchemy.orm import relationship
from .database import Base
import enum


class AuditAction(str, enum.Enum):
    CREATE = "CREATE"
    UPDATE = "UPDATE"
    DELETE = "DELETE"


class Vessel(Base):
    __tablename__ = "vessels"
    
    id = Column(Integer, primary_key=True, index=True)
    vessel_name = Column(String(100), nullable=False)
    vessel_number = Column(String(50), unique=True, index=True, nullable=True)
    vessel_type = Column(String(50), nullable=True)
    gross_tonnage = Column(Float, nullable=True)
    length = Column(Float, nullable=True)
    owner_name = Column(String(100), nullable=True)
    owner_id_card = Column(String(50), nullable=True)
    registry_port = Column(String(100), nullable=True)
    operating_company = Column(String(100), nullable=True)
    is_focus_attention = Column(Boolean, default=False, nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow, nullable=False)
    
    permits = relationship("Permit", back_populates="vessel")
    inspections = relationship("Inspection", back_populates="vessel")
    penalties = relationship("Penalty", back_populates="vessel")
    rectifications = relationship("Rectification", back_populates="vessel")


class PermitStatus(str, enum.Enum):
    PENDING = "待审批"
    APPROVED = "已通过"
    REJECTED = "已驳回"
    EXPIRED = "已过期"


class Permit(Base):
    __tablename__ = "permits"
    
    id = Column(Integer, primary_key=True, index=True)
    vessel_id = Column(Integer, ForeignKey("vessels.id"), nullable=True)
    permit_type = Column(String(100), nullable=False)
    application_number = Column(String(50), unique=True, nullable=False)
    applicant = Column(String(100), nullable=False)
    application_date = Column(Date, nullable=False)
    expected_effective_date = Column(Date, nullable=False)
    effective_date = Column(Date, nullable=True)
    expiry_date = Column(Date, nullable=True)
    status = Column(String(50), default=PermitStatus.PENDING.value, nullable=False)
    reject_reason = Column(Text, nullable=True)
    remarks = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow, nullable=False)
    
    vessel = relationship("Vessel", back_populates="permits")


class InspectionResult(str, enum.Enum):
    QUALIFIED = "合格"
    UNQUALIFIED = "不合格"


class Inspection(Base):
    __tablename__ = "inspections"
    
    id = Column(Integer, primary_key=True, index=True)
    vessel_id = Column(Integer, ForeignKey("vessels.id"), nullable=True)
    inspection_number = Column(String(50), unique=True, nullable=False)
    inspector = Column(String(100), nullable=False)
    inspection_date = Column(Date, nullable=False)
    inspection_location = Column(String(200), nullable=True)
    inspection_items = Column(Text, nullable=True)
    result = Column(String(50), nullable=False)
    unqualified_items = Column(Text, nullable=True)
    remarks = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow, nullable=False)
    
    vessel = relationship("Vessel", back_populates="inspections")
    rectifications = relationship("Rectification", back_populates="inspection")
    penalties = relationship("Penalty", back_populates="inspection")


class PenaltyStatus(str, enum.Enum):
    PENDING = "待处理"
    PROCESSING = "处理中"
    COMPLETED = "已完成"
    OVERDUE = "逾期"


class Penalty(Base):
    __tablename__ = "penalties"
    
    id = Column(Integer, primary_key=True, index=True)
    vessel_id = Column(Integer, ForeignKey("vessels.id"), nullable=True)
    inspection_id = Column(Integer, ForeignKey("inspections.id"), nullable=True)
    penalty_number = Column(String(50), unique=True, nullable=False)
    violation_description = Column(Text, nullable=False)
    fine_amount = Column(Float, nullable=False)
    issue_date = Column(Date, nullable=False)
    deadline_date = Column(Date, nullable=False)
    payment_date = Column(Date, nullable=True)
    status = Column(String(50), default=PenaltyStatus.PENDING.value, nullable=False)
    remarks = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow, nullable=False)
    
    vessel = relationship("Vessel", back_populates="penalties")
    inspection = relationship("Inspection", back_populates="penalties")


class RectificationStatus(str, enum.Enum):
    PENDING = "待整改"
    RECTIFYING = "整改中"
    RECHECKING = "待复查"
    REJECTED = "复查不通过"
    CLOSED = "已关闭"


class Rectification(Base):
    __tablename__ = "rectifications"
    
    id = Column(Integer, primary_key=True, index=True)
    vessel_id = Column(Integer, ForeignKey("vessels.id"), nullable=True)
    inspection_id = Column(Integer, ForeignKey("inspections.id"), nullable=True)
    rectification_number = Column(String(50), unique=True, nullable=False)
    rectification_content = Column(Text, nullable=False)
    created_date = Column(Date, nullable=False)
    rectification_deadline = Column(Date, nullable=True)
    rectification_measure = Column(Text, nullable=True)
    rectification_date = Column(Date, nullable=True)
    recheck_date = Column(Date, nullable=True)
    recheck_result = Column(String(200), nullable=True)
    status = Column(String(50), default=RectificationStatus.PENDING.value, nullable=False)
    recheck_count = Column(Integer, default=0, nullable=False)
    remarks = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow, nullable=False)
    
    vessel = relationship("Vessel", back_populates="rectifications")
    inspection = relationship("Inspection", back_populates="rectifications")


class AuditLog(Base):
    __tablename__ = "audit_logs"
    
    id = Column(Integer, primary_key=True, index=True)
    action = Column(String(50), nullable=False)
    model_name = Column(String(100), nullable=False)
    record_id = Column(Integer, nullable=True)
    detail = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)
