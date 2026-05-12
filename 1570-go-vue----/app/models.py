from sqlalchemy import Column, Integer, String, Float, DateTime, Boolean, ForeignKey, Text
from sqlalchemy.orm import relationship
from datetime import datetime

from .database import Base


class Flight(Base):
    __tablename__ = "flights"
    
    id = Column(Integer, primary_key=True, index=True)
    flight_number = Column(String, unique=True, index=True, nullable=False)
    route = Column(String, nullable=False)
    scheduled_departure = Column(DateTime, nullable=False)
    scheduled_arrival = Column(DateTime, nullable=False)
    unit_price = Column(Float, nullable=False)
    min_load_rate = Column(Float, default=0.60)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    
    compartments = relationship("Compartment", back_populates="flight", cascade="all, delete-orphan")
    cargoes = relationship("Cargo", back_populates="flight")


class Compartment(Base):
    __tablename__ = "compartments"
    
    id = Column(Integer, primary_key=True, index=True)
    compartment_number = Column(String, unique=True, index=True, nullable=False)
    flight_id = Column(Integer, ForeignKey("flights.id"), nullable=False)
    position = Column(String, nullable=False)
    total_capacity_weight = Column(Float, nullable=False)
    remaining_capacity_weight = Column(Float, nullable=False)
    is_temperature_controlled = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    
    flight = relationship("Flight", back_populates="compartments")
    load_assignments = relationship("LoadAssignment", back_populates="compartment")


class Cargo(Base):
    __tablename__ = "cargoes"
    
    id = Column(Integer, primary_key=True, index=True)
    cargo_number = Column(String, unique=True, index=True, nullable=False)
    flight_id = Column(Integer, ForeignKey("flights.id"), nullable=True)
    shipper = Column(String, nullable=False)
    consignee = Column(String, nullable=False)
    cargo_type = Column(String, nullable=False)
    description = Column(Text, nullable=True)
    actual_weight = Column(Float, nullable=True)
    volume = Column(Float, nullable=True)
    volume_weight = Column(Float, nullable=True)
    chargeable_weight = Column(Float, nullable=True)
    price_per_kg = Column(Float, nullable=True)
    freight_charge = Column(Float, nullable=True)
    storage_charge = Column(Float, default=0.0)
    arrival_time = Column(DateTime, nullable=True)
    status = Column(String, default="accepted")
    declaration_number = Column(String, nullable=True)
    quarantine_certificate_number = Column(String, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    
    flight = relationship("Flight", back_populates="cargoes")
    load_assignments = relationship("LoadAssignment", back_populates="cargo")


class LoadAssignment(Base):
    __tablename__ = "load_assignments"
    
    id = Column(Integer, primary_key=True, index=True)
    cargo_id = Column(Integer, ForeignKey("cargoes.id"), nullable=False)
    compartment_id = Column(Integer, ForeignKey("compartments.id"), nullable=False)
    assigned_weight = Column(Float, nullable=False)
    is_confirmed = Column(Boolean, default=False)
    confirmed_at = Column(DateTime, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    
    cargo = relationship("Cargo", back_populates="load_assignments")
    compartment = relationship("Compartment", back_populates="load_assignments")


class Dispatcher(Base):
    __tablename__ = "dispatchers"
    
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String, nullable=False)
    contact = Column(String, nullable=True)
    is_active = Column(Boolean, default=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)


class AlertLog(Base):
    __tablename__ = "alert_logs"
    
    id = Column(Integer, primary_key=True, index=True)
    flight_id = Column(Integer, ForeignKey("flights.id"), nullable=False)
    message = Column(Text, nullable=False)
    alert_type = Column(String, nullable=False)
    sent_to_dispatcher = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)
