from sqlalchemy import Column, Integer, String, Boolean, DateTime, ForeignKey
from sqlalchemy.orm import relationship
from datetime import datetime
from app.database import Base


class Switch(Base):
    __tablename__ = "switches"
    
    id = Column(Integer, primary_key=True, index=True)
    device_id = Column(String, unique=True, index=True)
    name = Column(String)
    position = Column(String, default="normal")
    locked = Column(Boolean, default=False)
    has_indication = Column(Boolean, default=True)
    occupied = Column(Boolean, default=False)
    status = Column(String, default="normal")
    last_updated = Column(DateTime, default=datetime.utcnow)


class Interlocking(Base):
    __tablename__ = "interlockings"
    
    id = Column(Integer, primary_key=True, index=True)
    device_id = Column(String, unique=True, index=True)
    name = Column(String)
    status = Column(String, default="normal")
    current_route = Column(String, nullable=True)
    route_status = Column(String, default="idle")
    last_updated = Column(DateTime, default=datetime.utcnow)


class BlockSection(Base):
    __tablename__ = "block_sections"
    
    id = Column(Integer, primary_key=True, index=True)
    device_id = Column(String, unique=True, index=True)
    name = Column(String)
    occupied = Column(Boolean, default=False)
    mode = Column(String, default="automatic")
    circuit_ok = Column(Boolean, default=True)
    last_updated = Column(DateTime, default=datetime.utcnow)


class Semaphore(Base):
    __tablename__ = "semaphores"
    
    id = Column(Integer, primary_key=True, index=True)
    device_id = Column(String, unique=True, index=True)
    name = Column(String)
    status = Column(String, default="red")
    fault = Column(Boolean, default=False)
    last_updated = Column(DateTime, default=datetime.utcnow)


class SwitchLog(Base):
    __tablename__ = "switch_logs"
    
    id = Column(Integer, primary_key=True, index=True)
    switch_id = Column(String, ForeignKey("switches.device_id"))
    action = Column(String)
    timestamp = Column(DateTime, default=datetime.utcnow)
    is_fault = Column(Boolean, default=False)


class BlockLog(Base):
    __tablename__ = "block_logs"
    
    id = Column(Integer, primary_key=True, index=True)
    section_id = Column(String, ForeignKey("block_sections.device_id"))
    action = Column(String)
    timestamp = Column(DateTime, default=datetime.utcnow)
    occupied = Column(Boolean, default=False)


class Alert(Base):
    __tablename__ = "alerts"
    
    id = Column(Integer, primary_key=True, index=True)
    device_type = Column(String)
    device_id = Column(String)
    message = Column(String)
    timestamp = Column(DateTime, default=datetime.utcnow)
    resolved = Column(Boolean, default=False)
