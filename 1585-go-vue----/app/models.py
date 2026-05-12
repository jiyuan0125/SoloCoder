from datetime import datetime
from sqlalchemy import Column, Integer, String, Float, Boolean, DateTime, ForeignKey, Text, Enum
from sqlalchemy.orm import relationship

from app.database import Base
import enum


class DeviceStatus(str, enum.Enum):
    RUNNING = "running"
    STOPPED = "stopped"
    FAULT = "fault"


class AlarmLevel(str, enum.Enum):
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"


class AlarmStatus(str, enum.Enum):
    ACTIVE = "active"
    ACKNOWLEDGED = "acknowledged"
    RESOLVED = "resolved"


class ControlMode(str, enum.Enum):
    MANUAL = "manual"
    AUTO = "auto"


class LightingGroup(str, enum.Enum):
    ENTRANCE = "entrance"
    MIDDLE = "middle"
    EXIT = "exit"


class MaintenanceTaskStatus(str, enum.Enum):
    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"


class VentilationDevice(Base):
    __tablename__ = "ventilation_devices"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String, index=True, nullable=False)
    status = Column(String, default=DeviceStatus.STOPPED.value)
    mode = Column(String, default=ControlMode.MANUAL.value)
    fault = Column(Boolean, default=False)

    co_level = Column(Float, default=0.0)
    visibility = Column(Float, default=100.0)
    co_threshold = Column(Float, default=100.0)
    visibility_threshold = Column(Float, default=10.0)

    is_running = Column(Boolean, default=False)
    start_time = Column(DateTime, nullable=True)
    shutdown_delay_active = Column(Boolean, default=False)
    shutdown_delay_start = Column(DateTime, nullable=True)

    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)


class LightingDevice(Base):
    __tablename__ = "lighting_devices"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String, index=True, nullable=False)
    status = Column(String, default=DeviceStatus.STOPPED.value)
    mode = Column(String, default=ControlMode.MANUAL.value)
    fault = Column(Boolean, default=False)

    brightness = Column(Integer, default=50)
    target_brightness = Column(Integer, default=50)
    group = Column(String, default=LightingGroup.MIDDLE.value)

    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)


class DrainageDevice(Base):
    __tablename__ = "drainage_devices"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String, index=True, nullable=False)
    status = Column(String, default=DeviceStatus.STOPPED.value)
    mode = Column(String, default=ControlMode.MANUAL.value)
    fault = Column(Boolean, default=False)

    model = Column(String, nullable=False)
    parameters = Column(Text, nullable=True)

    water_level = Column(Float, default=0.0)
    warning_level = Column(Float, default=5.0)
    safe_level = Column(Float, default=2.0)

    is_running = Column(Boolean, default=False)
    start_time = Column(DateTime, nullable=True)
    shutdown_delay_active = Column(Boolean, default=False)
    shutdown_delay_start = Column(DateTime, nullable=True)

    cumulative_runtime = Column(Float, default=0.0)
    maintenance_interval_hours = Column(Float, default=500.0)
    last_maintenance_runtime = Column(Float, default=0.0)

    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)


class Alarm(Base):
    __tablename__ = "alarms"

    id = Column(Integer, primary_key=True, index=True)
    subsystem = Column(String, index=True, nullable=False)
    device_id = Column(Integer, index=True, nullable=False)
    device_name = Column(String, nullable=False)
    level = Column(String, nullable=False)
    message = Column(Text, nullable=False)
    status = Column(String, default=AlarmStatus.ACTIVE.value)

    created_at = Column(DateTime, default=datetime.utcnow)
    acknowledged_at = Column(DateTime, nullable=True)
    resolved_at = Column(DateTime, nullable=True)


class MaintenanceTask(Base):
    __tablename__ = "maintenance_tasks"

    id = Column(Integer, primary_key=True, index=True)
    drainage_device_id = Column(Integer, ForeignKey("drainage_devices.id"))
    device_name = Column(String, nullable=False)
    device_model = Column(String, nullable=False)
    status = Column(String, default=MaintenanceTaskStatus.PENDING.value)
    content = Column(Text, nullable=False)

    created_at = Column(DateTime, default=datetime.utcnow)
    completed_at = Column(DateTime, nullable=True)

    device = relationship("DrainageDevice", backref="maintenance_tasks")
