from sqlalchemy import Column, Integer, String, Float, DateTime, ForeignKey, Text, Boolean, Enum as SQLEnum
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func
from .database import Base
from enum import Enum


class CapacityStatus(str, Enum):
    IDLE = "idle"
    BUSY = "busy"
    SATURATED = "saturated"
    RESTRICTED = "restricted"


class FlightType(str, Enum):
    DOMESTIC = "domestic"
    INTERNATIONAL = "international"


class Waypoint(Base):
    __tablename__ = "waypoints"

    id = Column(Integer, primary_key=True, index=True)
    code = Column(String(20), unique=True, index=True, nullable=False)
    name = Column(String(100))
    latitude = Column(Float, nullable=False)
    longitude = Column(Float, nullable=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())


class Route(Base):
    __tablename__ = "routes"

    id = Column(Integer, primary_key=True, index=True)
    code = Column(String(50), unique=True, index=True, nullable=False)
    name = Column(String(100))
    capacity = Column(Integer, default=10)
    busy_threshold = Column(Float, default=0.8)
    saturated_threshold = Column(Float, default=1.0)
    min_interval_minutes = Column(Integer, default=10)
    status = Column(SQLEnum(CapacityStatus), default=CapacityStatus.IDLE)
    average_speed = Column(Float, default=800.0)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())

    waypoints = relationship("RouteWaypoint", back_populates="route", cascade="all, delete-orphan")
    flights = relationship("Flight", back_populates="route")
    conflicts = relationship("Conflict", back_populates="route")
    flow_history = relationship("FlowRecord", back_populates="route")
    status_history = relationship("RouteStatusHistory", back_populates="route")


class RouteWaypoint(Base):
    __tablename__ = "route_waypoints"

    id = Column(Integer, primary_key=True, index=True)
    route_id = Column(Integer, ForeignKey("routes.id"), nullable=False)
    waypoint_id = Column(Integer, ForeignKey("waypoints.id"), nullable=False)
    sequence = Column(Integer, nullable=False)
    estimated_time_minutes = Column(Integer, default=0)

    route = relationship("Route", back_populates="waypoints")
    waypoint = relationship("Waypoint")


class Flight(Base):
    __tablename__ = "flights"

    id = Column(Integer, primary_key=True, index=True)
    flight_number = Column(String(20), unique=True, index=True, nullable=False)
    route_id = Column(Integer, ForeignKey("routes.id"), nullable=False)
    direction = Column(String(10), nullable=False)
    altitude = Column(Float, nullable=False)
    flight_type = Column(SQLEnum(FlightType), default=FlightType.DOMESTIC)
    estimated_entry_time = Column(DateTime(timezone=True), nullable=False)
    actual_entry_time = Column(DateTime(timezone=True))
    exit_time = Column(DateTime(timezone=True))
    is_waiting = Column(Boolean, default=False)
    is_completed = Column(Boolean, default=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now())

    route = relationship("Route", back_populates="flights")
    conflicts1 = relationship("Conflict", foreign_keys="Conflict.flight1_id", back_populates="flight1")
    conflicts2 = relationship("Conflict", foreign_keys="Conflict.flight2_id", back_populates="flight2")


class Conflict(Base):
    __tablename__ = "conflicts"

    id = Column(Integer, primary_key=True, index=True)
    route_id = Column(Integer, ForeignKey("routes.id"), nullable=False)
    flight1_id = Column(Integer, ForeignKey("flights.id"), nullable=False)
    flight2_id = Column(Integer, ForeignKey("flights.id"), nullable=False)
    conflict_type = Column(String(50), nullable=False)
    description = Column(Text)
    detected_at = Column(DateTime(timezone=True), server_default=func.now())
    resolved = Column(Boolean, default=False)
    resolution_suggestion = Column(Text)

    route = relationship("Route", back_populates="conflicts")
    flight1 = relationship("Flight", foreign_keys=[flight1_id], back_populates="conflicts1")
    flight2 = relationship("Flight", foreign_keys=[flight2_id], back_populates="conflicts2")


class FlowRecord(Base):
    __tablename__ = "flow_records"

    id = Column(Integer, primary_key=True, index=True)
    route_id = Column(Integer, ForeignKey("routes.id"), nullable=False)
    timestamp = Column(DateTime(timezone=True), server_default=func.now())
    active_flights = Column(Integer, default=0)
    waiting_flights = Column(Integer, default=0)
    capacity = Column(Integer)
    utilization = Column(Float)
    status = Column(SQLEnum(CapacityStatus))

    route = relationship("Route", back_populates="flow_history")


class RouteStatusHistory(Base):
    __tablename__ = "route_status_history"

    id = Column(Integer, primary_key=True, index=True)
    route_id = Column(Integer, ForeignKey("routes.id"), nullable=False)
    old_status = Column(SQLEnum(CapacityStatus))
    new_status = Column(SQLEnum(CapacityStatus), nullable=False)
    changed_at = Column(DateTime(timezone=True), server_default=func.now())
    reason = Column(String(200))

    route = relationship("Route", back_populates="status_history")
