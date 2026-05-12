from sqlalchemy import Column, Integer, String, Float, Boolean, DateTime, ForeignKey, Text
from sqlalchemy.orm import relationship
from datetime import datetime
from app.database import Base


class Park(Base):
    __tablename__ = "parks"
    
    id = Column(Integer, primary_key=True, index=True)
    code = Column(String(20), unique=True, index=True, nullable=False)
    name = Column(String(100), nullable=False)
    max_capacity = Column(Integer, nullable=False, default=1000)
    current_visitors = Column(Integer, default=0)
    created_at = Column(DateTime, default=datetime.utcnow)
    
    tickets = relationship("TicketType", back_populates="park")
    guides = relationship("Guide", back_populates="park")
    routes = relationship("Route", back_populates="park")


class TicketType(Base):
    __tablename__ = "ticket_types"
    
    id = Column(Integer, primary_key=True, index=True)
    park_id = Column(Integer, ForeignKey("parks.id"), nullable=False)
    name = Column(String(50), nullable=False)
    base_price = Column(Integer, nullable=False)
    description = Column(Text)
    is_active = Column(Boolean, default=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    
    park = relationship("Park", back_populates="tickets")
    tickets = relationship("Ticket", back_populates="ticket_type")


class Ticket(Base):
    __tablename__ = "tickets"
    
    id = Column(Integer, primary_key=True, index=True)
    ticket_type_id = Column(Integer, ForeignKey("ticket_types.id"), nullable=False)
    park_code = Column(String(20), nullable=False)
    ticket_type_name = Column(String(50), nullable=False)
    final_price = Column(Integer, nullable=False)
    visitor_name = Column(String(100))
    visitor_age = Column(Integer)
    visitor_height = Column(Float)
    is_student = Column(Boolean, default=False)
    is_group = Column(Boolean, default=False)
    group_size = Column(Integer, default=1)
    route_id = Column(Integer, ForeignKey("routes.id"))
    insurance_fee = Column(Integer, default=0)
    status = Column(String(20), default="valid")
    purchase_time = Column(DateTime, default=datetime.utcnow)
    refund_time = Column(DateTime)
    
    ticket_type = relationship("TicketType", back_populates="tickets")
    route = relationship("Route")


class Guide(Base):
    __tablename__ = "guides"
    
    id = Column(Integer, primary_key=True, index=True)
    guide_id = Column(String(20), unique=True, index=True, nullable=False)
    park_id = Column(Integer, ForeignKey("parks.id"), nullable=False)
    name = Column(String(100), nullable=False)
    level = Column(String(20), nullable=False)
    is_available = Column(Boolean, default=True)
    total_assignments = Column(Integer, default=0)
    total_rating = Column(Float, default=0.0)
    rating_count = Column(Integer, default=0)
    created_at = Column(DateTime, default=datetime.utcnow)
    
    park = relationship("Park", back_populates="guides")
    assignments = relationship("GuideAssignment", back_populates="guide")


class GuideAssignment(Base):
    __tablename__ = "guide_assignments"
    
    id = Column(Integer, primary_key=True, index=True)
    guide_id = Column(Integer, ForeignKey("guides.id"), nullable=False)
    park_code = Column(String(20), nullable=False)
    route_id = Column(Integer, ForeignKey("routes.id"))
    ticket_ids = Column(Text)
    duration_hours = Column(Float, nullable=False)
    total_fee = Column(Integer, nullable=False)
    status = Column(String(20), default="active")
    start_time = Column(DateTime, default=datetime.utcnow)
    end_time = Column(DateTime)
    rating = Column(Float)
    rating_comment = Column(Text)
    
    guide = relationship("Guide", back_populates="assignments")
    route = relationship("Route")


class Route(Base):
    __tablename__ = "routes"
    
    id = Column(Integer, primary_key=True, index=True)
    route_id = Column(String(20), unique=True, index=True, nullable=False)
    park_id = Column(Integer, ForeignKey("parks.id"), nullable=False)
    name = Column(String(100), nullable=False)
    difficulty = Column(String(20), default="normal")
    capacity = Column(Integer, default=100)
    current_visitors = Column(Integer, default=0)
    duration_hours = Column(Float, default=2.0)
    base_insurance_fee = Column(Integer, default=0)
    description = Column(Text)
    is_active = Column(Boolean, default=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    
    park = relationship("Park", back_populates="routes")
