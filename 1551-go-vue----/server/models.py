from datetime import datetime
from sqlalchemy import Column, Integer, String, Float, DateTime, ForeignKey, Boolean, Text
from sqlalchemy.orm import relationship, declarative_base

Base = declarative_base()


class Berth(Base):
    __tablename__ = "berths"
    
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(50), unique=True, index=True, nullable=False)
    berth_type = Column(String(50), nullable=False)
    capacity = Column(Float, nullable=False)
    is_under_maintenance = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    
    handling_operations = relationship("HandlingOperation", back_populates="berth")
    
    def check_available(self):
        if self.is_under_maintenance:
            return False
        for op in self.handling_operations:
            if op.status in ["pending", "in_progress"]:
                return False
        return True


class Ship(Base):
    __tablename__ = "ships"
    
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), unique=True, index=True, nullable=False)
    imo_number = Column(String(20), unique=True, index=True)
    ship_type = Column(String(50), nullable=False)
    draught = Column(Float, nullable=False)
    status = Column(String(20), default="expected", nullable=False)
    expected_arrival = Column(DateTime, nullable=False)
    current_berth_id = Column(Integer, ForeignKey("berths.id"), nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    
    handling_operations = relationship("HandlingOperation", back_populates="ship")
    
    def has_active_operation(self):
        for op in self.handling_operations:
            if op.status in ["pending", "in_progress", "waiting"]:
                return True
        return False


class HandlingOperation(Base):
    __tablename__ = "handling_operations"
    
    id = Column(Integer, primary_key=True, index=True)
    ship_id = Column(Integer, ForeignKey("ships.id"), nullable=False)
    berth_id = Column(Integer, ForeignKey("berths.id"), nullable=True)
    yard_area_id = Column(Integer, ForeignKey("yard_areas.id"), nullable=True)
    priority = Column(Integer, default=1, nullable=False)
    cargo_volume = Column(Float, nullable=False)
    status = Column(String(20), default="pending", nullable=False)
    started_at = Column(DateTime, nullable=True)
    completed_at = Column(DateTime, nullable=True)
    description = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    
    ship = relationship("Ship", back_populates="handling_operations")
    berth = relationship("Berth", back_populates="handling_operations")
    yard_area = relationship("YardArea", back_populates="handling_operations")


class YardArea(Base):
    __tablename__ = "yard_areas"
    
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), unique=True, index=True, nullable=False)
    total_capacity = Column(Float, nullable=False)
    is_available = Column(Boolean, default=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    
    handling_operations = relationship("HandlingOperation", back_populates="yard_area")
    
    def calculate_usage(self, db):
        total = 0
        for op in self.handling_operations:
            if op.status in ["pending", "in_progress"]:
                total += op.cargo_volume
        return total
    
    def has_capacity(self, volume, db):
        return (self.calculate_usage(db) + volume) <= self.total_capacity
