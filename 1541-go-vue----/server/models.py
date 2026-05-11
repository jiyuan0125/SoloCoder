from datetime import datetime, date
from enum import Enum as PyEnum
from sqlalchemy import (
    Column, Integer, String, Float, DateTime, Date, Boolean, ForeignKey, Enum, Text
)
from sqlalchemy.orm import relationship
from server.database import Base


class FacilityType(str, PyEnum):
    SPILLWAY_GATE = "spillway_gate"
    OVERFLOW_CHUTE = "overflow_chute"
    DRAINAGE_TUNNEL = "drainage_tunnel"


class FacilityStatus(str, PyEnum):
    IDLE = "idle"
    RUNNING = "running"
    MAINTENANCE = "maintenance"


class AlertLevel(str, PyEnum):
    NORMAL = "normal"
    YELLOW = "yellow"
    ORANGE = "orange"
    RED = "red"


class CommandStatus(str, PyEnum):
    PENDING_CONFIRM = "pending_confirm"
    PENDING_AUTHORIZATION = "pending_authorization"
    APPROVED = "approved"
    EXECUTED = "executed"
    REJECTED = "rejected"


class FloodModeStatus(str, PyEnum):
    NORMAL = "normal"
    FLOOD_CONTROL = "flood_control"


class Reservoir(Base):
    __tablename__ = "reservoirs"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), unique=True, nullable=False, index=True)
    normal_storage_level = Column(Float, nullable=False)
    design_flood_level = Column(Float, nullable=False)
    main_flood_season_start = Column(Integer, nullable=False, default=6)
    main_flood_season_end = Column(Integer, nullable=False, default=9)
    current_flood_mode = Column(String(20), Enum(FloodModeStatus), default=FloodModeStatus.NORMAL)
    flood_mode_start_time = Column(DateTime, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    facilities = relationship("FloodFacility", back_populates="reservoir", cascade="all, delete-orphan")
    water_levels = relationship("WaterLevelRecord", back_populates="reservoir", cascade="all, delete-orphan")
    inflows = relationship("InflowRecord", back_populates="reservoir", cascade="all, delete-orphan")
    commands = relationship("DispatchCommand", back_populates="reservoir", cascade="all, delete-orphan")
    flood_records = relationship("FloodOperationRecord", back_populates="reservoir", cascade="all, delete-orphan")
    alerts = relationship("Alert", back_populates="reservoir", cascade="all, delete-orphan")


class FloodFacility(Base):
    __tablename__ = "flood_facilities"

    id = Column(Integer, primary_key=True, index=True)
    reservoir_id = Column(Integer, ForeignKey("reservoirs.id"), nullable=False, index=True)
    name = Column(String(100), nullable=False)
    facility_type = Column(String(20), Enum(FacilityType), nullable=False)
    priority = Column(Integer, nullable=False, default=0)
    status = Column(String(20), Enum(FacilityStatus), default=FacilityStatus.IDLE)
    total_run_hours = Column(Float, default=0.0)
    run_start_time = Column(DateTime, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    reservoir = relationship("Reservoir", back_populates="facilities")
    gates = relationship("Gate", back_populates="facility", cascade="all, delete-orphan")


class Gate(Base):
    __tablename__ = "gates"

    id = Column(Integer, primary_key=True, index=True)
    facility_id = Column(Integer, ForeignKey("flood_facilities.id"), nullable=False, index=True)
    name = Column(String(100), nullable=False)
    gate_number = Column(Integer, nullable=False)
    current_open_percent = Column(Float, default=0.0)
    last_operation_time = Column(DateTime, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    facility = relationship("FloodFacility", back_populates="gates")
    commands = relationship("DispatchCommand", back_populates="gate")


class WaterLevelRecord(Base):
    __tablename__ = "water_level_records"

    id = Column(Integer, primary_key=True, index=True)
    reservoir_id = Column(Integer, ForeignKey("reservoirs.id"), nullable=False, index=True)
    level = Column(Float, nullable=False)
    record_time = Column(DateTime, nullable=False, index=True)
    is_aggregated = Column(Boolean, default=False)
    aggregate_hour = Column(Integer, nullable=True)
    aggregate_date = Column(Date, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)

    reservoir = relationship("Reservoir", back_populates="water_levels")


class InflowRecord(Base):
    __tablename__ = "inflow_records"

    id = Column(Integer, primary_key=True, index=True)
    reservoir_id = Column(Integer, ForeignKey("reservoirs.id"), nullable=False, index=True)
    inflow_rate = Column(Float, nullable=False)
    record_time = Column(DateTime, nullable=False, index=True)
    created_at = Column(DateTime, default=datetime.utcnow)

    reservoir = relationship("Reservoir", back_populates="inflows")


class DispatchCommand(Base):
    __tablename__ = "dispatch_commands"

    id = Column(Integer, primary_key=True, index=True)
    reservoir_id = Column(Integer, ForeignKey("reservoirs.id"), nullable=False, index=True)
    gate_id = Column(Integer, ForeignKey("gates.id"), nullable=False, index=True)
    command_type = Column(String(50), nullable=False)
    target_open_percent = Column(Float, nullable=False)
    operator_id = Column(String(50), nullable=False)
    operator_name = Column(String(100), nullable=False)
    confirmer_id = Column(String(50), nullable=True)
    confirmer_name = Column(String(100), nullable=True)
    authorize_id = Column(String(50), nullable=True)
    authorize_name = Column(String(100), nullable=True)
    status = Column(String(30), Enum(CommandStatus), default=CommandStatus.PENDING_CONFIRM)
    need_dual_confirm = Column(Boolean, default=True)
    need_headquarters_auth = Column(Boolean, default=False)
    executed_time = Column(DateTime, nullable=True)
    remark = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    reservoir = relationship("Reservoir", back_populates="commands")
    gate = relationship("Gate", back_populates="commands")


class FloodOperationRecord(Base):
    __tablename__ = "flood_operation_records"

    id = Column(Integer, primary_key=True, index=True)
    reservoir_id = Column(Integer, ForeignKey("reservoirs.id"), nullable=False, index=True)
    start_time = Column(DateTime, nullable=False)
    end_time = Column(DateTime, nullable=True)
    max_water_level = Column(Float, nullable=True)
    max_water_level_time = Column(DateTime, nullable=True)
    summary = Column(Text, nullable=True)
    commands_count = Column(Integer, default=0)
    created_at = Column(DateTime, default=datetime.utcnow)

    reservoir = relationship("Reservoir", back_populates="flood_records")


class Alert(Base):
    __tablename__ = "alerts"

    id = Column(Integer, primary_key=True, index=True)
    reservoir_id = Column(Integer, ForeignKey("reservoirs.id"), nullable=False, index=True)
    alert_level = Column(String(20), Enum(AlertLevel), nullable=False)
    alert_type = Column(String(50), nullable=False)
    message = Column(Text, nullable=False)
    trigger_value = Column(Float, nullable=True)
    threshold_value = Column(Float, nullable=True)
    is_acknowledged = Column(Boolean, default=False)
    acknowledged_time = Column(DateTime, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)

    reservoir = relationship("Reservoir", back_populates="alerts")
