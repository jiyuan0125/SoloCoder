from sqlalchemy import Column, Integer, String, Float, DateTime, Boolean, ForeignKey, Enum, Text
from sqlalchemy.orm import relationship
from datetime import datetime
import enum
from database import Base


class FlightStatus(str, enum.Enum):
    PLANNED = "计划中"
    DEPARTED = "已开航"
    ARRIVED = "已到港"
    CANCELLED = "取消"
    DELAYED = "延误"


class BerthStatus(str, enum.Enum):
    AVAILABLE = "空闲"
    OCCUPIED = "占用中"


class ShipStatus(str, enum.Enum):
    AVAILABLE = "可用"
    IN_MAINTENANCE = "年度检验中"
    IN_SERVICE = "运营中"


class TicketStatus(str, enum.Enum):
    SOLD = "已售出"
    REFUNDED = "已退票"


class Terminal(Base):
    __tablename__ = "terminals"
    
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), unique=True, nullable=False)
    location = Column(String(200))
    max_berths = Column(Integer, default=1)
    
    berths = relationship("Berth", back_populates="terminal")


class Berth(Base):
    __tablename__ = "berths"
    
    id = Column(Integer, primary_key=True, index=True)
    terminal_id = Column(Integer, ForeignKey("terminals.id"), nullable=False)
    berth_number = Column(String(20), nullable=False)
    capacity = Column(Integer, default=1)
    status = Column(String(20), default=BerthStatus.AVAILABLE)
    
    terminal = relationship("Terminal", back_populates="berths")
    flights = relationship("Flight", back_populates="berth")


class Ship(Base):
    __tablename__ = "ships"
    
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), unique=True, nullable=False)
    passenger_capacity = Column(Integer, nullable=False)
    vehicle_capacity = Column(Integer, default=0)
    status = Column(String(20), default=ShipStatus.AVAILABLE)
    inspection_start_date = Column(DateTime)
    inspection_end_date = Column(DateTime)
    
    flights = relationship("Flight", back_populates="ship")


class Flight(Base):
    __tablename__ = "flights"
    
    id = Column(Integer, primary_key=True, index=True)
    flight_number = Column(String(50), unique=True, nullable=False)
    ship_id = Column(Integer, ForeignKey("ships.id"))
    departure_terminal_id = Column(Integer, ForeignKey("terminals.id"), nullable=False)
    arrival_terminal_id = Column(Integer, ForeignKey("terminals.id"), nullable=False)
    berth_id = Column(Integer, ForeignKey("berths.id"))
    scheduled_departure = Column(DateTime, nullable=False)
    scheduled_arrival = Column(DateTime, nullable=False)
    actual_departure = Column(DateTime)
    actual_arrival = Column(DateTime)
    status = Column(String(20), default=FlightStatus.PLANNED)
    estimated_passengers = Column(Integer, default=0)
    actual_passengers = Column(Integer, default=0)
    actual_vehicles = Column(Integer, default=0)
    delay_minutes = Column(Integer, default=0)
    delay_notification_sent = Column(Boolean, default=False)
    
    ship = relationship("Ship", back_populates="flights")
    berth = relationship("Berth", back_populates="flights")
    tickets = relationship("Ticket", back_populates="flight")
    departure_terminal = relationship("Terminal", foreign_keys=[departure_terminal_id])
    arrival_terminal = relationship("Terminal", foreign_keys=[arrival_terminal_id])


class Ticket(Base):
    __tablename__ = "tickets"
    
    id = Column(Integer, primary_key=True, index=True)
    ticket_number = Column(String(50), unique=True, nullable=False)
    flight_id = Column(Integer, ForeignKey("flights.id"), nullable=False)
    passenger_name = Column(String(100))
    passenger_count = Column(Integer, default=1)
    vehicle_count = Column(Integer, default=0)
    is_on_site = Column(Boolean, default=False)
    status = Column(String(20), default=TicketStatus.SOLD)
    amount = Column(Float, default=0.0)
    created_at = Column(DateTime, default=datetime.utcnow)
    
    flight = relationship("Flight", back_populates="tickets")
    refunds = relationship("Refund", back_populates="ticket")


class Refund(Base):
    __tablename__ = "refunds"
    
    id = Column(Integer, primary_key=True, index=True)
    ticket_id = Column(Integer, ForeignKey("tickets.id"), nullable=False)
    refund_amount = Column(Float, default=0.0)
    reason = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)
    fee_waived = Column(Boolean, default=False)
    
    ticket = relationship("Ticket", back_populates="refunds")


class Weather(Base):
    __tablename__ = "weather"
    
    id = Column(Integer, primary_key=True, index=True)
    terminal_id = Column(Integer, ForeignKey("terminals.id"))
    timestamp = Column(DateTime, default=datetime.utcnow)
    wind_speed = Column(Float, default=0.0)
    visibility = Column(Integer, default=0)
    weather_condition = Column(String(100))
