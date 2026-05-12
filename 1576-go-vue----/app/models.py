from sqlalchemy import Column, Integer, String, DateTime, Float, Enum, ForeignKey, Text
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func
from app.database import Base
import enum


class MaintenanceType(str, enum.Enum):
    DAILY_INSPECTION = "日常巡视"
    PERIODIC_MAINTENANCE = "周期检修"
    FAULT_REPAIR = "故障抢修"


class WorkOrderStatus(str, enum.Enum):
    CREATED = "创建"
    PENDING_EXECUTION = "待执行"
    IN_PROGRESS = "执行中"
    COMPLETED = "已完成"


class AlarmSeverity(str, enum.Enum):
    LOW = "低"
    MEDIUM = "中"
    HIGH = "高"


class WorkOrder(Base):
    __tablename__ = "work_orders"

    id = Column(Integer, primary_key=True, index=True)
    section = Column(String(100), nullable=False, index=True)
    maintenance_type = Column(Enum(MaintenanceType), nullable=False)
    status = Column(Enum(WorkOrderStatus), default=WorkOrderStatus.CREATED, nullable=False)
    plan_start_time = Column(DateTime, nullable=True)
    plan_end_time = Column(DateTime, nullable=True)
    actual_start_time = Column(DateTime, nullable=True)
    actual_end_time = Column(DateTime, nullable=True)
    person_in_charge = Column(String(50), nullable=False)
    description = Column(Text, nullable=True)
    created_at = Column(DateTime, server_default=func.now())
    updated_at = Column(DateTime, server_default=func.now(), onupdate=func.now())

    alarm_work_order = relationship("AlarmWorkOrder", back_populates="original_work_order", uselist=False)


class PowerSupplyData(Base):
    __tablename__ = "power_supply_data"

    id = Column(Integer, primary_key=True, index=True)
    section = Column(String(100), nullable=False, index=True)
    current = Column(Float, nullable=False)
    voltage = Column(Float, nullable=False)
    rated_current = Column(Float, default=1000.0, nullable=False)
    is_power_outage = Column(Integer, default=0)
    is_alarm = Column(Integer, default=0)
    recorded_at = Column(DateTime, server_default=func.now(), index=True)


class SectionStatistics(Base):
    __tablename__ = "section_statistics"

    id = Column(Integer, primary_key=True, index=True)
    section = Column(String(100), nullable=False, unique=True, index=True)
    total_inspection_count = Column(Integer, default=0)
    total_power_outage_duration_minutes = Column(Integer, default=0)
    total_alarm_count = Column(Integer, default=0)
    last_updated_at = Column(DateTime, server_default=func.now(), onupdate=func.now())


class AlarmWorkOrder(Base):
    __tablename__ = "alarm_work_orders"

    id = Column(Integer, primary_key=True, index=True)
    section = Column(String(100), nullable=False, index=True)
    severity = Column(Enum(AlarmSeverity), nullable=False)
    description = Column(Text, nullable=True)
    original_work_order_id = Column(Integer, ForeignKey("work_orders.id"), nullable=True)
    status = Column(Enum(WorkOrderStatus), default=WorkOrderStatus.CREATED, nullable=False)
    created_at = Column(DateTime, server_default=func.now())

    original_work_order = relationship("WorkOrder", back_populates="alarm_work_order")
