from sqlalchemy import Column, Integer, String, DateTime, Float, Boolean, ForeignKey, Text
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func
from .database import Base


class Flight(Base):
    __tablename__ = "flights"

    id = Column(String, primary_key=True, index=True)
    flight_number = Column(String, unique=True, nullable=False)
    departure = Column(String, nullable=False)
    destination = Column(String, nullable=False)
    scheduled_departure = Column(DateTime, nullable=True)
    scheduled_arrival = Column(DateTime, nullable=True)
    max_load_weight = Column(Float, nullable=False, default=2000.0)
    actual_load_weight = Column(Float, default=0.0)

    baggage = relationship("Baggage", back_populates="flight")
    sorting_records = relationship("SortingRecord", back_populates="flight")
    loading_records = relationship("LoadingRecord", back_populates="flight")
    alerts = relationship("Alert", back_populates="flight")


class Baggage(Base):
    __tablename__ = "baggage"

    id = Column(String, primary_key=True, index=True)
    tag_number = Column(String, unique=True, nullable=False, index=True)
    weight = Column(Float, nullable=False)
    is_valuable = Column(Boolean, default=False)
    passenger_name = Column(String, nullable=False)
    flight_id = Column(String, ForeignKey("flights.id"), nullable=True)
    current_status = Column(String, nullable=False, default="CHECKED_IN")
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())

    flight = relationship("Flight", back_populates="baggage")
    events = relationship("BaggageEvent", back_populates="baggage")
    sorting_records = relationship("SortingRecord", back_populates="baggage")
    loading_records = relationship("LoadingRecord", back_populates="baggage")
    claims = relationship("Claim", back_populates="baggage")
    alerts = relationship("Alert", back_populates="baggage")


class BaggageEvent(Base):
    __tablename__ = "baggage_events"

    id = Column(Integer, primary_key=True, autoincrement=True)
    baggage_id = Column(String, ForeignKey("baggage.id"), nullable=False)
    event_type = Column(String, nullable=False)
    event_time = Column(DateTime(timezone=True), server_default=func.now())
    location = Column(String, nullable=True)
    notes = Column(Text, nullable=True)
    success = Column(Boolean, default=True)

    baggage = relationship("Baggage", back_populates="events")


class SortingRecord(Base):
    __tablename__ = "sorting_records"

    id = Column(Integer, primary_key=True, autoincrement=True)
    baggage_id = Column(String, ForeignKey("baggage.id"), nullable=False)
    flight_id = Column(String, ForeignKey("flights.id"), nullable=True)
    sorting_time = Column(DateTime(timezone=True), server_default=func.now())
    success = Column(Boolean, default=True)
    failure_reason = Column(String, nullable=True)
    lane = Column(String, nullable=True)

    baggage = relationship("Baggage", back_populates="sorting_records")
    flight = relationship("Flight", back_populates="sorting_records")


class LoadingRecord(Base):
    __tablename__ = "loading_records"

    id = Column(Integer, primary_key=True, autoincrement=True)
    baggage_id = Column(String, ForeignKey("baggage.id"), nullable=False)
    flight_id = Column(String, ForeignKey("flights.id"), nullable=False)
    loading_time = Column(DateTime(timezone=True), server_default=func.now())
    position = Column(String, nullable=False)
    is_special_position = Column(Boolean, default=False)

    baggage = relationship("Baggage", back_populates="loading_records")
    flight = relationship("Flight", back_populates="loading_records")


class ConveyorBeltRecord(Base):
    __tablename__ = "conveyor_belt_records"

    id = Column(Integer, primary_key=True, autoincrement=True)
    baggage_id = Column(String, ForeignKey("baggage.id"), nullable=False)
    conveyor_belt_number = Column(String, nullable=False)
    placed_at = Column(DateTime(timezone=True), server_default=func.now())
    picked_at = Column(DateTime(timezone=True), nullable=True)
    is_delayed = Column(Boolean, default=False)
    is_stuck = Column(Boolean, default=False)


class Alert(Base):
    __tablename__ = "alerts"

    id = Column(Integer, primary_key=True, autoincrement=True)
    alert_type = Column(String, nullable=False)
    flight_id = Column(String, ForeignKey("flights.id"), nullable=True)
    baggage_id = Column(String, ForeignKey("baggage.id"), nullable=True)
    alert_time = Column(DateTime(timezone=True), server_default=func.now())
    message = Column(Text, nullable=False)
    resolved = Column(Boolean, default=False)
    resolved_at = Column(DateTime(timezone=True), nullable=True)

    flight = relationship("Flight", back_populates="alerts")
    baggage = relationship("Baggage", back_populates="alerts")


class Claim(Base):
    __tablename__ = "claims"

    id = Column(Integer, primary_key=True, autoincrement=True)
    baggage_id = Column(String, ForeignKey("baggage.id"), nullable=False)
    claim_amount = Column(Float, nullable=False)
    max_allowed = Column(Float, nullable=False)
    status = Column(String, nullable=False, default="PENDING")
    filed_at = Column(DateTime(timezone=True), server_default=func.now())
    resolved_at = Column(DateTime(timezone=True), nullable=True)
    notes = Column(Text, nullable=True)

    baggage = relationship("Baggage", back_populates="claims")
