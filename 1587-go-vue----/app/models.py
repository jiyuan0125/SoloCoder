from sqlalchemy import Column, Integer, String, Float, DateTime, ForeignKey, Enum, Boolean, Text
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func
from app.database import Base
import enum

class StationStatus(str, enum.Enum):
    ACTIVE = "active"
    INACTIVE = "inactive"

class ChargerStatus(str, enum.Enum):
    IDLE = "idle"
    CHARGING = "charging"
    OCCUPIED = "occupied"
    FAULTY = "faulty"
    LOCKED = "locked"

class VehicleStatus(str, enum.Enum):
    IDLE = "idle"
    RUNNING = "running"
    CHARGING = "charging"
    LOCKED = "locked"
    MAINTENANCE = "maintenance"
    RESERVE = "reserve"

class RouteStatus(str, enum.Enum):
    ACTIVE = "active"
    INACTIVE = "inactive"

class AssignmentStatus(str, enum.Enum):
    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"
    FAILED = "failed"
    CANCELLED = "cancelled"

class AlertType(str, enum.Enum):
    LOW_BATTERY = "low_battery"
    CHARGER_FAULT = "charger_fault"
    SCHEDULE_ISSUE = "schedule_issue"
    SYSTEM_ERROR = "system_error"

class AlertStatus(str, enum.Enum):
    ACTIVE = "active"
    ACKNOWLEDGED = "acknowledged"
    RESOLVED = "resolved"

class NotificationType(str, enum.Enum):
    TERMINAL = "terminal"
    LOG_ONLY = "log_only"

class Station(Base):
    __tablename__ = "stations"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), unique=True, nullable=False)
    status = Column(Enum(StationStatus), default=StationStatus.ACTIVE)
    location = Column(String(200))
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())

    chargers = relationship("Charger", back_populates="station", cascade="all, delete-orphan")

class Charger(Base):
    __tablename__ = "chargers"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(50), unique=True, nullable=False)
    station_id = Column(Integer, ForeignKey("stations.id"), nullable=True)
    status = Column(Enum(ChargerStatus), default=ChargerStatus.IDLE)
    power_kw = Column(Float, default=50.0)
    total_charging_count = Column(Integer, default=0)
    locked_by_vehicle_id = Column(Integer, ForeignKey("vehicles.id"), nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())

    station = relationship("Station", back_populates="chargers")
    charging_sessions = relationship("ChargingSession", back_populates="charger", cascade="all, delete-orphan")
    locked_vehicle = relationship("Vehicle", foreign_keys=[locked_by_vehicle_id], backref="locked_charger")

class Vehicle(Base):
    __tablename__ = "vehicles"

    id = Column(Integer, primary_key=True, index=True)
    plate_number = Column(String(20), unique=True, nullable=False)
    status = Column(Enum(VehicleStatus), default=VehicleStatus.IDLE)
    current_battery = Column(Float, default=100.0)
    max_battery_capacity_kwh = Column(Float, default=80.0)
    current_route_id = Column(Integer, ForeignKey("routes.id"), nullable=True)
    current_station_id = Column(Integer, ForeignKey("stations.id"), nullable=True)
    latitude = Column(Float, nullable=True)
    longitude = Column(Float, nullable=True)
    last_report_time = Column(DateTime(timezone=True), nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())

    current_route = relationship("Route", back_populates="active_vehicles")
    current_station = relationship("Station")
    charging_sessions = relationship("ChargingSession", back_populates="vehicle", cascade="all, delete-orphan")
    assignments = relationship("Assignment", back_populates="vehicle", cascade="all, delete-orphan")

class Route(Base):
    __tablename__ = "routes"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), unique=True, nullable=False)
    status = Column(Enum(RouteStatus), default=RouteStatus.ACTIVE)
    total_distance_km = Column(Float, default=0.0)
    estimated_duration_min = Column(Integer, default=0)
    start_station_id = Column(Integer, ForeignKey("stations.id"), nullable=True)
    end_station_id = Column(Integer, ForeignKey("stations.id"), nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())

    active_vehicles = relationship("Vehicle", back_populates="current_route")
    schedules = relationship("Schedule", back_populates="route", cascade="all, delete-orphan")

class Schedule(Base):
    __tablename__ = "schedules"

    id = Column(Integer, primary_key=True, index=True)
    route_id = Column(Integer, ForeignKey("routes.id"), nullable=False)
    vehicle_id = Column(Integer, ForeignKey("vehicles.id"), nullable=True)
    departure_time = Column(DateTime(timezone=True), nullable=False)
    return_time = Column(DateTime(timezone=True), nullable=False)
    is_last_run = Column(Boolean, default=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now())

    route = relationship("Route", back_populates="schedules")
    vehicle = relationship("Vehicle")

class ChargingSession(Base):
    __tablename__ = "charging_sessions"

    id = Column(Integer, primary_key=True, index=True)
    vehicle_id = Column(Integer, ForeignKey("vehicles.id"), nullable=False)
    charger_id = Column(Integer, ForeignKey("chargers.id"), nullable=False)
    start_time = Column(DateTime(timezone=True), nullable=False)
    end_time = Column(DateTime(timezone=True), nullable=True)
    start_battery = Column(Float, nullable=False)
    target_battery = Column(Float, default=90.0)
    current_battery = Column(Float, nullable=True)
    estimated_end_time = Column(DateTime(timezone=True), nullable=True)
    status = Column(Enum(AssignmentStatus), default=AssignmentStatus.IN_PROGRESS)
    created_at = Column(DateTime(timezone=True), server_default=func.now())

    vehicle = relationship("Vehicle", back_populates="charging_sessions")
    charger = relationship("Charger", back_populates="charging_sessions")

class Assignment(Base):
    __tablename__ = "assignments"

    id = Column(Integer, primary_key=True, index=True)
    vehicle_id = Column(Integer, ForeignKey("vehicles.id"), nullable=False)
    route_id = Column(Integer, ForeignKey("routes.id"), nullable=True)
    schedule_id = Column(Integer, ForeignKey("schedules.id"), nullable=True)
    status = Column(Enum(AssignmentStatus), default=AssignmentStatus.PENDING)
    reason = Column(String(500), nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())

    vehicle = relationship("Vehicle", back_populates="assignments")
    route = relationship("Route")
    schedule = relationship("Schedule")

class Alert(Base):
    __tablename__ = "alerts"

    id = Column(Integer, primary_key=True, index=True)
    type = Column(Enum(AlertType), nullable=False)
    vehicle_id = Column(Integer, ForeignKey("vehicles.id"), nullable=True)
    charger_id = Column(Integer, ForeignKey("chargers.id"), nullable=True)
    message = Column(Text, nullable=False)
    status = Column(Enum(AlertStatus), default=AlertStatus.ACTIVE)
    notification_type = Column(Enum(NotificationType), default=NotificationType.TERMINAL)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    resolved_at = Column(DateTime(timezone=True), nullable=True)

    vehicle = relationship("Vehicle")
    charger = relationship("Charger")

class DispatchLog(Base):
    __tablename__ = "dispatch_logs"

    id = Column(Integer, primary_key=True, index=True)
    level = Column(String(20), default="INFO")
    message = Column(Text, nullable=False)
    related_vehicle_id = Column(Integer, ForeignKey("vehicles.id"), nullable=True)
    related_charger_id = Column(Integer, ForeignKey("chargers.id"), nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())

    vehicle = relationship("Vehicle")
    charger = relationship("Charger")
