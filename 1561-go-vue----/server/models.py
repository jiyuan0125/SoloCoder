import enum
from datetime import datetime
from sqlalchemy import Column, Integer, String, DateTime, ForeignKey, Boolean, Enum, Text
from sqlalchemy.orm import relationship
from .database import Base


class FlightStatus(str, enum.Enum):
    SCHEDULED = "scheduled"
    CHECK_IN = "check_in"
    BOARDING = "boarding"
    DEPARTED = "departed"
    COMPLETED = "completed"


class AircraftType(str, enum.Enum):
    NARROW_BODY = "narrow_body"
    WIDE_BODY = "wide_body"


class VehicleStatus(str, enum.Enum):
    AVAILABLE = "available"
    IN_USE = "in_use"
    MAINTENANCE = "maintenance"


class VehicleType(str, enum.Enum):
    PASSENGER_BUS = "passenger_bus"
    LUGGAGE_TOW = "luggage_tow"
    FUEL_TRUCK = "fuel_truck"
    CATERING = "catering"
    MAINTENANCE_TRUCK = "maintenance_truck"


class GateAdjacency(Base):
    __tablename__ = "gate_adjacencies"
    id = Column(Integer, primary_key=True, index=True)
    gate_id_1 = Column(Integer, ForeignKey("gates.id"), nullable=False)
    gate_id_2 = Column(Integer, ForeignKey("gates.id"), nullable=False)


class Flight(Base):
    __tablename__ = "flights"
    id = Column(Integer, primary_key=True, index=True)
    flight_number = Column(String(20), unique=True, index=True, nullable=False)
    aircraft_type = Column(Enum(AircraftType), nullable=False)
    status = Column(Enum(FlightStatus), default=FlightStatus.SCHEDULED, nullable=False)
    
    scheduled_departure = Column(DateTime, nullable=False)
    scheduled_arrival = Column(DateTime)
    check_in_start = Column(DateTime)
    check_in_end = Column(DateTime)
    boarding_start = Column(DateTime)
    boarding_end = Column(DateTime)
    actual_departure = Column(DateTime)
    actual_arrival = Column(DateTime)
    
    stand_id = Column(Integer, ForeignKey("stands.id"))
    gate_id = Column(Integer, ForeignKey("gates.id"))
    
    stand = relationship("Stand", back_populates="flights")
    gate = relationship("Gate", back_populates="flights")
    vehicle_dispatches = relationship("VehicleDispatch", back_populates="flight")


class Stand(Base):
    __tablename__ = "stands"
    id = Column(Integer, primary_key=True, index=True)
    code = Column(String(10), unique=True, index=True, nullable=False)
    is_bridge = Column(Boolean, default=False, nullable=False)
    supports_wide_body = Column(Boolean, default=False, nullable=False)
    is_occupied = Column(Boolean, default=False, nullable=False)
    
    flights = relationship("Flight", back_populates="stand")


class Gate(Base):
    __tablename__ = "gates"
    id = Column(Integer, primary_key=True, index=True)
    code = Column(String(10), unique=True, index=True, nullable=False)
    is_available = Column(Boolean, default=True, nullable=False)
    
    flights = relationship("Flight", back_populates="gate")


class Vehicle(Base):
    __tablename__ = "vehicles"
    id = Column(Integer, primary_key=True, index=True)
    vehicle_code = Column(String(20), unique=True, index=True, nullable=False)
    vehicle_type = Column(Enum(VehicleType), nullable=False)
    status = Column(Enum(VehicleStatus), default=VehicleStatus.AVAILABLE, nullable=False)
    
    dispatches = relationship("VehicleDispatch", back_populates="vehicle")
    maintenances = relationship("VehicleMaintenance", back_populates="vehicle")


class VehicleDispatch(Base):
    __tablename__ = "vehicle_dispatches"
    id = Column(Integer, primary_key=True, index=True)
    vehicle_id = Column(Integer, ForeignKey("vehicles.id"))
    flight_id = Column(Integer, ForeignKey("flights.id"), nullable=False)
    vehicle_type = Column(Enum(VehicleType), nullable=False)
    priority = Column(Integer, default=0, nullable=False)
    is_waiting = Column(Boolean, default=False, nullable=False)
    assigned_at = Column(DateTime, default=datetime.utcnow)
    started_at = Column(DateTime)
    completed_at = Column(DateTime)
    
    vehicle = relationship("Vehicle", back_populates="dispatches")
    flight = relationship("Flight", back_populates="vehicle_dispatches")


class VehicleMaintenance(Base):
    __tablename__ = "vehicle_maintenances"
    id = Column(Integer, primary_key=True, index=True)
    vehicle_id = Column(Integer, ForeignKey("vehicles.id"), nullable=False)
    description = Column(Text, nullable=False)
    scheduled_at = Column(DateTime, default=datetime.utcnow)
    started_at = Column(DateTime)
    completed_at = Column(DateTime)
    is_completed = Column(Boolean, default=False, nullable=False)
    
    vehicle = relationship("Vehicle", back_populates="maintenances")
