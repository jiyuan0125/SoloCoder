from datetime import datetime
from sqlalchemy import Column, Integer, String, Float, DateTime, ForeignKey, Enum as SQLEnum
from sqlalchemy.orm import relationship
import enum

from app.database import Base


class PerformanceStatus(str, enum.Enum):
    PLANNED = "planned"
    ON_SALE = "on_sale"
    SOLD_OUT = "sold_out"
    PERFORMING = "performing"
    FINISHED = "finished"


class TicketStatus(str, enum.Enum):
    SOLD = "sold"
    ISSUED = "issued"
    USED = "used"
    REFUNDED = "refunded"


class Performance(Base):
    __tablename__ = "performances"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(200), nullable=False)
    description = Column(String(500), nullable=True)
    start_time = Column(DateTime, nullable=False)
    end_time = Column(DateTime, nullable=False)
    venue_id = Column(Integer, ForeignKey("venues.id"), nullable=False)
    total_tickets = Column(Integer, nullable=False)
    price = Column(Integer, nullable=False)
    status = Column(SQLEnum(PerformanceStatus), default=PerformanceStatus.PLANNED)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    venue = relationship("Venue", back_populates="performances")
    tickets = relationship("Ticket", back_populates="performance")


class Ticket(Base):
    __tablename__ = "tickets"

    id = Column(Integer, primary_key=True, index=True)
    performance_id = Column(Integer, ForeignKey("performances.id"), nullable=False)
    customer_name = Column(String(100), nullable=False)
    customer_phone = Column(String(20), nullable=True)
    price = Column(Integer, nullable=False)
    seat_number = Column(String(20), nullable=True)
    status = Column(SQLEnum(TicketStatus), default=TicketStatus.SOLD)
    sold_at = Column(DateTime, default=datetime.utcnow)
    issued_at = Column(DateTime, nullable=True)
    used_at = Column(DateTime, nullable=True)
    refunded_at = Column(DateTime, nullable=True)
    refund_amount = Column(Integer, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    performance = relationship("Performance", back_populates="tickets")
