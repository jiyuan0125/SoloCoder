from datetime import datetime
from sqlalchemy import Column, Integer, String, Float, DateTime, ForeignKey, Boolean
from sqlalchemy.orm import relationship, declarative_base

Base = declarative_base()

class Berth(Base):
    __tablename__ = "berths"
    
    id = Column(Integer, primary_key=True, autoincrement=True)
    name = Column(String(100), unique=True, nullable=False)
    type = Column(String(50), nullable=False)
    capacity = Column(Float, nullable=False)
    is_under_maintenance = Column(Boolean, default=False)
    current_ship_id = Column(Integer, ForeignKey("ships.id"), nullable=True)
    
    ship = relationship("Ship", back_populates="current_berth", foreign_keys=[current_ship_id])
    operations = relationship("Operation", back_populates="berth")

class Ship(Base):
    __tablename__ = "ships"
    
    id = Column(Integer, primary_key=True, autoincrement=True)
    name = Column(String(100), unique=True, nullable=False)
    type = Column(String(50), nullable=False)
    draft = Column(Float, nullable=False)
    status = Column(String(50), default="arriving")
    eta = Column(DateTime, nullable=False)
    
    current_berth = relationship("Berth", back_populates="ship", foreign_keys="Berth.current_ship_id", uselist=False)
    operations = relationship("Operation", back_populates="ship")

class YardZone(Base):
    __tablename__ = "yard_zones"
    
    id = Column(Integer, primary_key=True, autoincrement=True)
    name = Column(String(100), unique=True, nullable=False)
    total_capacity = Column(Float, nullable=False)
    used_capacity = Column(Float, default=0)
    is_available = Column(Boolean, default=True)
    
    items = relationship("YardItem", back_populates="zone", cascade="all, delete-orphan")

class YardItem(Base):
    __tablename__ = "yard_items"
    
    id = Column(Integer, primary_key=True, autoincrement=True)
    zone_id = Column(Integer, ForeignKey("yard_zones.id"), nullable=False)
    operation_id = Column(Integer, ForeignKey("operations.id"), nullable=True)
    quantity = Column(Float, nullable=False)
    status = Column(String(50), default="in_queue")
    created_at = Column(DateTime, default=datetime.utcnow)
    
    zone = relationship("YardZone", back_populates="items")
    operation = relationship("Operation", back_populates="yard_items")

class Operation(Base):
    __tablename__ = "operations"
    
    id = Column(Integer, primary_key=True, autoincrement=True)
    ship_id = Column(Integer, ForeignKey("ships.id"), nullable=False)
    berth_id = Column(Integer, ForeignKey("berths.id"), nullable=False)
    priority = Column(Integer, default=5)
    yard_zone_id = Column(Integer, ForeignKey("yard_zones.id"), nullable=True)
    quantity = Column(Float, nullable=False)
    status = Column(String(50), default="queued")
    created_at = Column(DateTime, default=datetime.utcnow)
    started_at = Column(DateTime, nullable=True)
    completed_at = Column(DateTime, nullable=True)
    
    ship = relationship("Ship", back_populates="operations")
    berth = relationship("Berth", back_populates="operations")
    yard_zone = relationship("YardZone")
    yard_items = relationship("YardItem", back_populates="operation", cascade="all, delete-orphan")
