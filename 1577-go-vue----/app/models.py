from sqlalchemy import Column, Integer, String, DateTime, ForeignKey, Enum, Text, Float, Boolean
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func
from app.database import Base
import enum


class MaintenanceType(str, enum.Enum):
    SECTION = "段修"
    FACTORY = "厂修"


class MaintenanceStatus(str, enum.Enum):
    CREATED = "创建"
    PENDING_REVIEW = "待审核"
    APPROVED = "已审核"
    IN_PROGRESS = "执行中"
    COMPLETED = "完成"
    CANCELLED = "取消"


class Locomotive(Base):
    __tablename__ = "locomotives"

    id = Column(Integer, primary_key=True, index=True)
    locomotive_number = Column(String(50), unique=True, index=True, nullable=False)
    current_km = Column(Float, default=0.0)
    last_section_maintenance_km = Column(Float, default=0.0)
    last_factory_maintenance_km = Column(Float, default=0.0)
    section_cycle_km = Column(Float, default=300000.0)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())

    maintenance_plans = relationship("MaintenancePlan", back_populates="locomotive")


class MaintenancePlan(Base):
    __tablename__ = "maintenance_plans"

    id = Column(Integer, primary_key=True, index=True)
    plan_number = Column(String(50), unique=True, index=True, nullable=False)
    locomotive_id = Column(Integer, ForeignKey("locomotives.id"), nullable=False)
    maintenance_type = Column(Enum(MaintenanceType), nullable=False)
    planned_km = Column(Float, nullable=False)
    status = Column(Enum(MaintenanceStatus), default=MaintenanceStatus.CREATED)
    is_cancelled = Column(Boolean, default=False)
    maintenance_content = Column(Text, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())

    locomotive = relationship("Locomotive", back_populates="maintenance_plans")
    replacement_parts = relationship("ReplacementPart", back_populates="maintenance_plan")


class Part(Base):
    __tablename__ = "parts"

    id = Column(Integer, primary_key=True, index=True)
    part_code = Column(String(50), unique=True, index=True, nullable=False)
    part_name = Column(String(100), nullable=False)
    stock = Column(Integer, default=0)
    warning_threshold = Column(Integer, default=10)
    unit = Column(String(20), default="个")
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())

    replacement_parts = relationship("ReplacementPart", back_populates="part")


class ReplacementPart(Base):
    __tablename__ = "replacement_parts"

    id = Column(Integer, primary_key=True, index=True)
    maintenance_plan_id = Column(Integer, ForeignKey("maintenance_plans.id"), nullable=False)
    part_id = Column(Integer, ForeignKey("parts.id"), nullable=False)
    quantity = Column(Integer, nullable=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now())

    maintenance_plan = relationship("MaintenancePlan", back_populates="replacement_parts")
    part = relationship("Part", back_populates="replacement_parts")


class TechnicalManual(Base):
    __tablename__ = "technical_manuals"

    id = Column(Integer, primary_key=True, index=True)
    manual_code = Column(String(50), unique=True, index=True, nullable=False)
    title = Column(String(200), nullable=False)
    author = Column(String(100), nullable=True)
    version = Column(String(50), nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())

    borrow_records = relationship("BorrowRecord", back_populates="manual")


class BorrowRecord(Base):
    __tablename__ = "borrow_records"

    id = Column(Integer, primary_key=True, index=True)
    manual_id = Column(Integer, ForeignKey("technical_manuals.id"), nullable=False)
    borrower = Column(String(100), nullable=False)
    borrow_date = Column(DateTime(timezone=True), server_default=func.now())
    due_date = Column(DateTime(timezone=True), nullable=False)
    return_date = Column(DateTime(timezone=True), nullable=True)
    is_overdue = Column(Boolean, default=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now())

    manual = relationship("TechnicalManual", back_populates="borrow_records")


class PurchaseAlert(Base):
    __tablename__ = "purchase_alerts"

    id = Column(Integer, primary_key=True, index=True)
    part_id = Column(Integer, ForeignKey("parts.id"), nullable=False)
    current_stock = Column(Integer, nullable=False)
    threshold = Column(Integer, nullable=False)
    is_active = Column(Boolean, default=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())

    part = relationship("Part")
