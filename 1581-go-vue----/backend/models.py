from sqlalchemy import (
    Column,
    Integer,
    String,
    DateTime,
    Boolean,
    ForeignKey,
    Enum,
    Float,
    Text,
)
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func
import enum

from .database import Base


class TrainStatus(str, enum.Enum):
    IDLE = "idle"
    RUNNING = "running"
    ARRIVED = "arrived"
    DELAYED = "delayed"
    STOPPED = "stopped"


class SignalStatus(str, enum.Enum):
    NORMAL = "normal"
    FAULT = "fault"
    RECOVERED = "recovered"


class PowerStatus(str, enum.Enum):
    ACTIVE = "active"
    OUTAGE = "outage"


class LogLevel(str, enum.Enum):
    INFO = "info"
    WARNING = "warning"
    ERROR = "error"


class Line(Base):
    __tablename__ = "lines"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), unique=True, index=True, nullable=False)
    description = Column(String(255))
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())

    stations = relationship("Station", back_populates="line")
    power_sections = relationship("PowerSection", back_populates="line")
    schedules = relationship("Schedule", back_populates="line")


class Station(Base):
    __tablename__ = "stations"

    id = Column(Integer, primary_key=True, index=True)
    line_id = Column(Integer, ForeignKey("lines.id"), nullable=False)
    name = Column(String(100), nullable=False)
    sequence = Column(Integer, nullable=False)
    is_terminal = Column(Boolean, default=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now())

    line = relationship("Line", back_populates="stations")
    signals = relationship("Signal", back_populates="station")


class Train(Base):
    __tablename__ = "trains"

    id = Column(Integer, primary_key=True, index=True)
    train_number = Column(String(50), unique=True, index=True, nullable=False)
    status = Column(Enum(TrainStatus), default=TrainStatus.IDLE)
    current_station_id = Column(Integer, ForeignKey("stations.id"), nullable=True)
    current_section_id = Column(Integer, ForeignKey("power_sections.id"), nullable=True)
    speed_limit = Column(Float, default=60.0)
    capacity = Column(Integer, default=1000)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())

    current_station = relationship("Station")
    current_section = relationship("PowerSection")
    schedule_trains = relationship("ScheduleTrain", back_populates="train")


class Signal(Base):
    __tablename__ = "signals"

    id = Column(Integer, primary_key=True, index=True)
    signal_code = Column(String(50), unique=True, index=True, nullable=False)
    station_id = Column(Integer, ForeignKey("stations.id"), nullable=False)
    section_id = Column(Integer, ForeignKey("power_sections.id"), nullable=True)
    status = Column(Enum(SignalStatus), default=SignalStatus.NORMAL)
    fault_time = Column(DateTime(timezone=True), nullable=True)
    recovery_time = Column(DateTime(timezone=True), nullable=True)
    fault_description = Column(Text, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())

    station = relationship("Station", back_populates="signals")
    section = relationship("PowerSection")
    fault_reports = relationship("FaultReport", back_populates="signal")


class PowerSection(Base):
    __tablename__ = "power_sections"

    id = Column(Integer, primary_key=True, index=True)
    line_id = Column(Integer, ForeignKey("lines.id"), nullable=False)
    section_code = Column(String(50), unique=True, index=True, nullable=False)
    name = Column(String(100), nullable=False)
    start_station_id = Column(Integer, ForeignKey("stations.id"))
    end_station_id = Column(Integer, ForeignKey("stations.id"))
    status = Column(Enum(PowerStatus), default=PowerStatus.ACTIVE)
    outage_time = Column(DateTime(timezone=True), nullable=True)
    recovery_time = Column(DateTime(timezone=True), nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())

    line = relationship("Line", back_populates="power_sections")
    start_station = relationship("Station", foreign_keys=[start_station_id])
    end_station = relationship("Station", foreign_keys=[end_station_id])
    signals = relationship("Signal", back_populates="section")


class FaultReport(Base):
    __tablename__ = "fault_reports"

    id = Column(Integer, primary_key=True, index=True)
    signal_id = Column(Integer, ForeignKey("signals.id"), nullable=False)
    report_time = Column(DateTime(timezone=True), server_default=func.now())
    fault_description = Column(Text, nullable=True)
    resolved_by = Column(String(100), nullable=True)
    resolved_at = Column(DateTime(timezone=True), nullable=True)

    signal = relationship("Signal", back_populates="fault_reports")


class Schedule(Base):
    __tablename__ = "schedules"

    id = Column(Integer, primary_key=True, index=True)
    line_id = Column(Integer, ForeignKey("lines.id"), nullable=False)
    name = Column(String(100), nullable=False)
    first_departure = Column(DateTime(timezone=True), nullable=False)
    last_departure = Column(DateTime(timezone=True), nullable=False)
    interval_seconds = Column(Integer, default=300)
    is_active = Column(Boolean, default=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())

    line = relationship("Line", back_populates="schedules")
    schedule_trains = relationship("ScheduleTrain", back_populates="schedule")


class ScheduleTrain(Base):
    __tablename__ = "schedule_trains"

    id = Column(Integer, primary_key=True, index=True)
    schedule_id = Column(Integer, ForeignKey("schedules.id"), nullable=False)
    train_id = Column(Integer, ForeignKey("trains.id"), nullable=True)
    sequence = Column(Integer, nullable=False)
    scheduled_departure = Column(DateTime(timezone=True), nullable=False)
    actual_departure = Column(DateTime(timezone=True), nullable=True)
    scheduled_arrival = Column(DateTime(timezone=True), nullable=True)
    actual_arrival = Column(DateTime(timezone=True), nullable=True)
    turnaround_time = Column(Integer, nullable=True)
    is_completed = Column(Boolean, default=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())

    schedule = relationship("Schedule", back_populates="schedule_trains")
    train = relationship("Train", back_populates="schedule_trains")


class PassengerData(Base):
    __tablename__ = "passenger_data"

    id = Column(Integer, primary_key=True, index=True)
    line_id = Column(Integer, ForeignKey("lines.id"), nullable=False)
    station_id = Column(Integer, ForeignKey("stations.id"), nullable=True)
    timestamp = Column(DateTime(timezone=True), nullable=False)
    passenger_count = Column(Integer, default=0)
    created_at = Column(DateTime(timezone=True), server_default=func.now())

    line = relationship("Line")
    station = relationship("Station")


class DispatchLog(Base):
    __tablename__ = "dispatch_logs"

    id = Column(Integer, primary_key=True, index=True)
    level = Column(Enum(LogLevel), default=LogLevel.INFO)
    message = Column(Text, nullable=False)
    entity_type = Column(String(50), nullable=True)
    entity_id = Column(Integer, nullable=True)
    operator = Column(String(100), nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
