from sqlalchemy import Column, Integer, String, DateTime, ForeignKey, Text, Enum as SQLAlchemyEnum, Float, Date
from sqlalchemy.orm import relationship
from app.database import Base
from datetime import datetime, date
import enum

class DeviceType(str, enum.Enum):
    LINE = "线路"
    BRIDGE = "桥梁"
    TUNNEL = "隧道"

class DeviceStatus(str, enum.Enum):
    GOOD = "正常"
    WARNING = "警告"
    CRITICAL = "严重"

class BridgeGrade(str, enum.Enum):
    SMALL = "小桥"
    MEDIUM = "中桥"
    LARGE = "大桥"
    EXTRA_LARGE = "特大桥"

class InspectionType(str, enum.Enum):
    DAILY = "日常检查"
    SPECIAL = "专项检查"
    FLOOD = "防洪检查"

class InspectionStatus(str, enum.Enum):
    PENDING = "待开始"
    IN_PROGRESS = "进行中"
    COMPLETED = "已完成"
    OVERDUE_COMPLETED = "逾期完成"
    CANCELLED = "已取消"

class ProblemLevel(str, enum.Enum):
    MINOR = "轻微"
    MAJOR = "重要"
    SERIOUS = "严重"

class Device(Base):
    __tablename__ = "devices"

    id = Column(Integer, primary_key=True, index=True)
    device_code = Column(String(50), unique=True, index=True, nullable=False)
    device_type = Column(SQLAlchemyEnum(DeviceType), nullable=False)
    name = Column(String(100), nullable=False)
    mileage = Column(String(50), nullable=False)
    status = Column(SQLAlchemyEnum(DeviceStatus), default=DeviceStatus.GOOD, nullable=False)
    
    bridge_length = Column(Float, nullable=True)
    bridge_grade = Column(SQLAlchemyEnum(BridgeGrade), nullable=True)
    
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow, nullable=False)

    status_history = relationship("DeviceStatusHistory", back_populates="device", cascade="all, delete-orphan")
    inspections = relationship("Inspection", back_populates="device", cascade="all, delete-orphan")
    inspection_plans = relationship("InspectionPlan", back_populates="device", cascade="all, delete-orphan")

class DeviceStatusHistory(Base):
    __tablename__ = "device_status_history"

    id = Column(Integer, primary_key=True, index=True)
    device_id = Column(Integer, ForeignKey("devices.id"), nullable=False)
    old_status = Column(SQLAlchemyEnum(DeviceStatus), nullable=False)
    new_status = Column(SQLAlchemyEnum(DeviceStatus), nullable=False)
    change_time = Column(DateTime, default=datetime.utcnow, nullable=False)
    remark = Column(String(255), nullable=True)

    device = relationship("Device", back_populates="status_history")

class Inspector(Base):
    __tablename__ = "inspectors"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(50), nullable=False)
    contact = Column(String(50), nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)

    inspections = relationship("Inspection", back_populates="inspector")

class InspectionPlan(Base):
    __tablename__ = "inspection_plans"

    id = Column(Integer, primary_key=True, index=True)
    device_id = Column(Integer, ForeignKey("devices.id"), nullable=False)
    inspection_type = Column(SQLAlchemyEnum(InspectionType), nullable=False)
    frequency_per_month = Column(Integer, nullable=False)
    effective_month = Column(Date, nullable=False)
    is_active = Column(Integer, default=1, nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)

    device = relationship("Device", back_populates="inspection_plans")
    inspections = relationship("Inspection", back_populates="plan")

class Inspection(Base):
    __tablename__ = "inspections"

    id = Column(Integer, primary_key=True, index=True)
    inspection_code = Column(String(50), unique=True, index=True, nullable=False)
    device_id = Column(Integer, ForeignKey("devices.id"), nullable=False)
    plan_id = Column(Integer, ForeignKey("inspection_plans.id"), nullable=True)
    inspector_id = Column(Integer, ForeignKey("inspectors.id"), nullable=False)
    inspection_type = Column(SQLAlchemyEnum(InspectionType), nullable=False)
    
    planned_start_time = Column(DateTime, nullable=False)
    planned_end_time = Column(DateTime, nullable=False)
    
    actual_start_time = Column(DateTime, nullable=True)
    actual_end_time = Column(DateTime, nullable=True)
    
    status = Column(SQLAlchemyEnum(InspectionStatus), default=InspectionStatus.PENDING, nullable=False)
    
    problem_level = Column(SQLAlchemyEnum(ProblemLevel), nullable=True)
    handling_suggestion = Column(Text, nullable=True)
    
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow, nullable=False)

    device = relationship("Device", back_populates="inspections")
    inspector = relationship("Inspector", back_populates="inspections")
    plan = relationship("InspectionPlan", back_populates="inspections")
