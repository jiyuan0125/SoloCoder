from sqlalchemy import Column, Integer, String, DateTime, Float, ForeignKey, Boolean, Text, UniqueConstraint
from sqlalchemy.orm import relationship
from server.database import Base
from datetime import datetime

class Route(Base):
    __tablename__ = "routes"
    
    id = Column(Integer, primary_key=True, index=True)
    route_code = Column(String(50), unique=True, index=True, nullable=False)
    origin_city = Column(String(100), nullable=False)
    destination_city = Column(String(100), nullable=False)
    status = Column(String(50), nullable=False, default="planning")
    aircraft_type = Column(String(50), nullable=True)
    seating_capacity = Column(Integer, nullable=True)
    weekly_frequency = Column(Integer, nullable=True)
    base_fuel_consumption = Column(Integer, nullable=True)
    planning_date = Column(DateTime, default=datetime.utcnow)
    trial_start_date = Column(DateTime, nullable=True)
    formal_start_date = Column(DateTime, nullable=True)
    pause_date = Column(DateTime, nullable=True)
    termination_date = Column(DateTime, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    
    flights = relationship("Flight", back_populates="route")
    revenue_data = relationship("RevenueData", back_populates="route")
    suggestions = relationship("OptimizationSuggestion", back_populates="route")
    
    __table_args__ = (
        UniqueConstraint('origin_city', 'destination_city', name='unique_city_pair_active'),
    )

class Flight(Base):
    __tablename__ = "flights"
    
    id = Column(Integer, primary_key=True, index=True)
    route_id = Column(Integer, ForeignKey("routes.id"), nullable=False)
    flight_number = Column(String(20), nullable=False)
    departure_datetime = Column(DateTime, nullable=False)
    arrival_datetime = Column(DateTime, nullable=False)
    status = Column(String(50), default="scheduled")
    total_seats = Column(Integer, nullable=False)
    booked_seats = Column(Integer, default=0)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    
    route = relationship("Route", back_populates="flights")

class RevenueData(Base):
    __tablename__ = "revenue_data"
    
    id = Column(Integer, primary_key=True, index=True)
    route_id = Column(Integer, ForeignKey("routes.id"), nullable=False)
    year = Column(Integer, nullable=False)
    month = Column(Integer, nullable=False)
    total_flights = Column(Integer, default=0)
    total_booked_seats = Column(Integer, default=0)
    total_available_seats = Column(Integer, default=0)
    passenger_revenue_cents = Column(Integer, default=0)
    fuel_cost_cents = Column(Integer, default=0)
    other_costs_cents = Column(Integer, default=0)
    route_status_at_month = Column(String(50), nullable=False)
    total_revenue_cents = Column(Integer, default=0)
    total_cost_cents = Column(Integer, default=0)
    net_profit_cents = Column(Integer, default=0)
    load_factor = Column(Float, default=0.0)
    average_ticket_price = Column(Float, default=0.0)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    
    route = relationship("Route", back_populates="revenue_data")

class FuelPrice(Base):
    __tablename__ = "fuel_prices"
    
    id = Column(Integer, primary_key=True, index=True)
    year = Column(Integer, nullable=False)
    month = Column(Integer, nullable=False)
    price_per_liter_cents = Column(Integer, nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)

class OptimizationSuggestion(Base):
    __tablename__ = "optimization_suggestions"
    
    id = Column(Integer, primary_key=True, index=True)
    route_id = Column(Integer, ForeignKey("routes.id"), nullable=False)
    suggestion_type = Column(String(50), nullable=False)
    reason = Column(Text, nullable=False)
    generated_at = Column(DateTime, default=datetime.utcnow)
    is_implemented = Column(Boolean, default=False)
    implemented_at = Column(DateTime, nullable=True)
    
    route = relationship("Route", back_populates="suggestions")
