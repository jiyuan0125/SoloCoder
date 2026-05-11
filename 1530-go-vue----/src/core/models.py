from datetime import datetime
from enum import Enum
from sqlalchemy import Column, Integer, String, Float, DateTime, ForeignKey, Text, Boolean, Date
from sqlalchemy.orm import relationship
from .database import Base


class RiskLevel(str, Enum):
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"


class PipelineStatus(str, Enum):
    ACTIVE = "active"
    INACTIVE = "inactive"
    MAINTENANCE = "maintenance"


class SegmentStatus(str, Enum):
    OPERATIONAL = "operational"
    WARNING = "warning"
    CRITICAL = "critical"
    MAINTENANCE = "maintenance"


class AlarmStatus(str, Enum):
    NEW = "new"
    IN_PROCESS = "in_process"
    URGENT = "urgent"
    RESOLVED = "resolved"
    FALSE_ALARM = "false_alarm"


class AlarmSeverity(str, Enum):
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    URGENT = "urgent"


class Pipeline(Base):
    __tablename__ = "pipelines"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), nullable=False, index=True)
    code = Column(String(50), unique=True, nullable=False, index=True)
    description = Column(Text)
    total_length = Column(Float, nullable=False)
    design_wall_thickness = Column(Float, nullable=False)
    operating_pressure = Column(Float, nullable=False)
    material = Column(String(100))
    diameter = Column(Float)
    installation_date = Column(Date)
    risk_level = Column(String(20), default=RiskLevel.MEDIUM)
    status = Column(String(20), default=PipelineStatus.ACTIVE)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    segments = relationship("Segment", back_populates="pipeline", cascade="all, delete-orphan")
    patrol_plans = relationship("PatrolPlan", back_populates="pipeline")
    integrity_records = relationship("IntegrityRecord", back_populates="pipeline")
    pressure_points = relationship("PressurePoint", back_populates="pipeline")


class Segment(Base):
    __tablename__ = "segments"

    id = Column(Integer, primary_key=True, index=True)
    pipeline_id = Column(Integer, ForeignKey("pipelines.id"), nullable=False)
    name = Column(String(100), nullable=False)
    start_position = Column(Float, nullable=False)
    end_position = Column(Float, nullable=False)
    length = Column(Float, nullable=False)
    description = Column(Text)
    risk_level = Column(String(20), default=RiskLevel.MEDIUM)
    patrol_frequency_days = Column(Integer, default=7)
    status = Column(String(20), default=SegmentStatus.OPERATIONAL)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    pipeline = relationship("Pipeline", back_populates="segments")
    patrol_records = relationship("PatrolRecord", back_populates="segment")


class PressurePoint(Base):
    __tablename__ = "pressure_points"

    id = Column(Integer, primary_key=True, index=True)
    pipeline_id = Column(Integer, ForeignKey("pipelines.id"), nullable=False)
    name = Column(String(100), nullable=False)
    position = Column(Float, nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)

    pipeline = relationship("Pipeline", back_populates="pressure_points")
    readings = relationship("PressureReading", back_populates="pressure_point")


class PressureReading(Base):
    __tablename__ = "pressure_readings"

    id = Column(Integer, primary_key=True, index=True)
    pressure_point_id = Column(Integer, ForeignKey("pressure_points.id"), nullable=False)
    pressure = Column(Float, nullable=False)
    timestamp = Column(DateTime, default=datetime.utcnow, index=True)
    is_aggregated = Column(Boolean, default=False)

    pressure_point = relationship("PressurePoint", back_populates="readings")


class PressureDiff(Base):
    __tablename__ = "pressure_diffs"

    id = Column(Integer, primary_key=True, index=True)
    pipeline_id = Column(Integer, ForeignKey("pipelines.id"), nullable=False)
    start_point_id = Column(Integer, ForeignKey("pressure_points.id"), nullable=False)
    end_point_id = Column(Integer, ForeignKey("pressure_points.id"), nullable=False)
    pressure_diff = Column(Float, nullable=False)
    timestamp = Column(DateTime, default=datetime.utcnow, index=True)


class Alarm(Base):
    __tablename__ = "alarms"

    id = Column(Integer, primary_key=True, index=True)
    alarm_type = Column(String(50), nullable=False)
    severity = Column(String(20), default=AlarmSeverity.MEDIUM)
    status = Column(String(20), default=AlarmStatus.NEW)
    pipeline_id = Column(Integer, ForeignKey("pipelines.id"), nullable=False)
    segment_id = Column(Integer, ForeignKey("segments.id"))
    pressure_diff_id = Column(Integer, ForeignKey("pressure_diffs.id"))
    title = Column(String(200), nullable=False)
    description = Column(Text)
    false_alarm_reason = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow, index=True)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    resolved_at = Column(DateTime)
    handled_by = Column(String(100))
    escalated = Column(Boolean, default=False)
    reminders = Column(Integer, default=0)

    pressure_diff = relationship("PressureDiff")


class PatrolPlan(Base):
    __tablename__ = "patrol_plans"

    id = Column(Integer, primary_key=True, index=True)
    pipeline_id = Column(Integer, ForeignKey("pipelines.id"), nullable=False)
    plan_date = Column(Date, index=True)
    status = Column(String(20), default="pending")
    created_at = Column(DateTime, default=datetime.utcnow)

    pipeline = relationship("Pipeline", back_populates="patrol_plans")
    plan_items = relationship("PatrolPlanItem", back_populates="plan", cascade="all, delete-orphan")


class PatrolPlanItem(Base):
    __tablename__ = "patrol_plan_items"

    id = Column(Integer, primary_key=True, index=True)
    plan_id = Column(Integer, ForeignKey("patrol_plans.id"), nullable=False)
    segment_id = Column(Integer, ForeignKey("segments.id"), nullable=False)
    scheduled_time = Column(DateTime)
    assigned_to = Column(String(100))
    completed = Column(Boolean, default=False)

    plan = relationship("PatrolPlan", back_populates="plan_items")
    segment = relationship("Segment")


class PatrolRecord(Base):
    __tablename__ = "patrol_records"

    id = Column(Integer, primary_key=True, index=True)
    segment_id = Column(Integer, ForeignKey("segments.id"), nullable=False)
    patrol_date = Column(Date, index=True)
    inspector = Column(String(100))
    findings = Column(Text)
    status = Column(String(50), default="normal")
    created_at = Column(DateTime, default=datetime.utcnow)

    segment = relationship("Segment", back_populates="patrol_records")


class IntegrityRecord(Base):
    __tablename__ = "integrity_records"

    id = Column(Integer, primary_key=True, index=True)
    pipeline_id = Column(Integer, ForeignKey("pipelines.id"), nullable=False)
    segment_id = Column(Integer, ForeignKey("segments.id"))
    inspection_date = Column(Date, index=True)
    wall_thickness = Column(Float, nullable=False)
    location = Column(String(200))
    notes = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)

    pipeline = relationship("Pipeline", back_populates="integrity_records")
    segment = relationship("Segment")


class MaintenanceTask(Base):
    __tablename__ = "maintenance_tasks"

    id = Column(Integer, primary_key=True, index=True)
    task_type = Column(String(50), nullable=False)
    pipeline_id = Column(Integer, ForeignKey("pipelines.id"), nullable=False)
    segment_id = Column(Integer, ForeignKey("segments.id"))
    title = Column(String(200), nullable=False)
    description = Column(Text)
    priority = Column(String(20), default="medium")
    status = Column(String(20), default="pending")
    due_date = Column(Date)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)


class DailyReport(Base):
    __tablename__ = "daily_reports"

    id = Column(Integer, primary_key=True, index=True)
    report_date = Column(Date, unique=True, index=True)
    total_pipelines = Column(Integer, default=0)
    active_pipelines = Column(Integer, default=0)
    total_alarms = Column(Integer, default=0)
    new_alarms = Column(Integer, default=0)
    resolved_alarms = Column(Integer, default=0)
    pressure_readings_count = Column(Integer, default=0)
    patrols_completed = Column(Integer, default=0)
    maintenance_tasks = Column(Integer, default=0)
    created_at = Column(DateTime, default=datetime.utcnow)
    content = Column(Text)
