from sqlalchemy import Column, Integer, String, DateTime, Float, Boolean, ForeignKey, Text, UniqueConstraint
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func
from app.database import Base
from enum import Enum as PyEnum
from sqlalchemy import Enum


class Direction(PyEnum):
    UP = "up"
    DOWN = "down"


class ScheduleStatus(PyEnum):
    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"
    CANCELLED = "cancelled"


class Station(Base):
    __tablename__ = "stations"
    
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), nullable=False)
    code = Column(String(50), nullable=False, unique=True)
    latitude = Column(Float, nullable=True)
    longitude = Column(Float, nullable=True)
    is_transfer = Column(Boolean, default=False)
    created_at = Column(DateTime, server_default=func.now())
    updated_at = Column(DateTime, server_default=func.now(), onupdate=func.now())
    
    route_stations = relationship("RouteStation", back_populates="station")
    signal_priorities = relationship("SignalPriority", back_populates="station")
    arrival_records = relationship("ArrivalRecord", back_populates="station")


class Route(Base):
    __tablename__ = "routes"
    
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), nullable=False)
    code = Column(String(50), nullable=False, unique=True)
    start_station = Column(String(100), nullable=False)
    end_station = Column(String(100), nullable=False)
    created_at = Column(DateTime, server_default=func.now())
    updated_at = Column(DateTime, server_default=func.now(), onupdate=func.now())
    
    route_stations = relationship("RouteStation", back_populates="route")
    operation_params = relationship("OperationParameter", back_populates="route")
    schedules = relationship("Schedule", back_populates="route")


class RouteStation(Base):
    __tablename__ = "route_stations"
    
    id = Column(Integer, primary_key=True, index=True)
    route_id = Column(Integer, ForeignKey("routes.id"), nullable=False)
    station_id = Column(Integer, ForeignKey("stations.id"), nullable=False)
    direction = Column(Enum(Direction), nullable=False)
    sequence = Column(Integer, nullable=False)
    travel_time_from_prev = Column(Integer, nullable=False, default=2)
    stop_time = Column(Integer, nullable=False, default=1)
    
    route = relationship("Route", back_populates="route_stations")
    station = relationship("Station", back_populates="route_stations")
    
    __table_args__ = (
        UniqueConstraint('route_id', 'direction', 'sequence', name='unique_route_direction_sequence'),
    )


class Vehicle(Base):
    __tablename__ = "vehicles"
    
    id = Column(Integer, primary_key=True, index=True)
    plate_number = Column(String(20), nullable=False, unique=True)
    vehicle_code = Column(String(50), nullable=False, unique=True)
    capacity = Column(Integer, nullable=False, default=80)
    current_route_id = Column(Integer, ForeignKey("routes.id"), nullable=True)
    current_direction = Column(Enum(Direction), nullable=True)
    status = Column(String(20), default="idle")
    current_load = Column(Integer, default=0)
    created_at = Column(DateTime, server_default=func.now())
    updated_at = Column(DateTime, server_default=func.now(), onupdate=func.now())
    
    signals = relationship("SignalPriorityRequest", back_populates="vehicle")
    schedules = relationship("Schedule", back_populates="vehicle")


class SignalPriority(Base):
    __tablename__ = "signal_priorities"
    
    id = Column(Integer, primary_key=True, index=True)
    station_id = Column(Integer, ForeignKey("stations.id"), nullable=False)
    intersection_name = Column(String(100), nullable=False)
    priority_interval = Column(Integer, nullable=False)
    is_active = Column(Boolean, default=True)
    created_at = Column(DateTime, server_default=func.now())
    
    station = relationship("Station", back_populates="signal_priorities")


class OperationParameter(Base):
    __tablename__ = "operation_parameters"
    
    id = Column(Integer, primary_key=True, index=True)
    route_id = Column(Integer, ForeignKey("routes.id"), nullable=False, unique=True)
    turnaround_time = Column(Integer, nullable=False, default=10)
    default_interval = Column(Integer, nullable=False, default=10)
    start_time = Column(String(5), nullable=False, default="06:00")
    end_time = Column(String(5), nullable=False, default="22:00")
    max_signal_priority_daily = Column(Integer, nullable=False, default=20)
    created_at = Column(DateTime, server_default=func.now())
    updated_at = Column(DateTime, server_default=func.now(), onupdate=func.now())
    
    route = relationship("Route", back_populates="operation_params")


class Schedule(Base):
    __tablename__ = "schedules"
    
    id = Column(Integer, primary_key=True, index=True)
    route_id = Column(Integer, ForeignKey("routes.id"), nullable=False)
    vehicle_id = Column(Integer, ForeignKey("vehicles.id"), nullable=False)
    direction = Column(Enum(Direction), nullable=False)
    start_time = Column(DateTime, nullable=False)
    end_time = Column(DateTime, nullable=False)
    status = Column(Enum(ScheduleStatus), default=ScheduleStatus.PENDING)
    schedule_data = Column(Text, nullable=False)
    created_at = Column(DateTime, server_default=func.now())
    updated_at = Column(DateTime, server_default=func.now(), onupdate=func.now())
    
    route = relationship("Route", back_populates="schedules")
    vehicle = relationship("Vehicle", back_populates="schedules")
    arrival_records = relationship("ArrivalRecord", back_populates="schedule")


class ArrivalRecord(Base):
    __tablename__ = "arrival_records"
    
    id = Column(Integer, primary_key=True, index=True)
    schedule_id = Column(Integer, ForeignKey("schedules.id"), nullable=False)
    station_id = Column(Integer, ForeignKey("stations.id"), nullable=False)
    planned_arrival_time = Column(DateTime, nullable=False)
    actual_arrival_time = Column(DateTime, nullable=True)
    is_on_time = Column(Boolean, nullable=True)
    created_at = Column(DateTime, server_default=func.now())
    
    schedule = relationship("Schedule", back_populates="arrival_records")
    station = relationship("Station", back_populates="arrival_records")


class SignalPriorityRequest(Base):
    __tablename__ = "signal_priority_requests"
    
    id = Column(Integer, primary_key=True, index=True)
    vehicle_id = Column(Integer, ForeignKey("vehicles.id"), nullable=False)
    intersection_name = Column(String(100), nullable=False)
    request_time = Column(DateTime, server_default=func.now())
    granted = Column(Boolean, default=True)
    created_at = Column(DateTime, server_default=func.now())
    
    vehicle = relationship("Vehicle", back_populates="signals")
