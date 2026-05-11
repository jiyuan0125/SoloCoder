from sqlalchemy import Column, Integer, String, Float, DateTime, ForeignKey, Boolean, Text
from sqlalchemy.orm import relationship
from datetime import datetime
from .database import Base


class Warehouse(Base):
    __tablename__ = "warehouses"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), nullable=False)
    location = Column(String(200))
    created_at = Column(DateTime, default=datetime.utcnow)

    storage_areas = relationship("StorageArea", back_populates="warehouse", cascade="all, delete-orphan")


class StorageArea(Base):
    __tablename__ = "storage_areas"

    id = Column(Integer, primary_key=True, index=True)
    warehouse_id = Column(Integer, ForeignKey("warehouses.id"), nullable=False)
    name = Column(String(100), nullable=False)
    area_type = Column(String(50), nullable=False)
    capacity_kg = Column(Float, nullable=False)
    min_temp = Column(Float, default=10.0)
    max_temp = Column(Float, default=25.0)
    created_at = Column(DateTime, default=datetime.utcnow)

    warehouse = relationship("Warehouse", back_populates="storage_areas")
    batches = relationship("GrainBatch", back_populates="storage_area", cascade="all, delete-orphan")
    temperature_records = relationship("TemperatureRecord", back_populates="storage_area", cascade="all, delete-orphan")
    pest_inspections = relationship("PestInspection", back_populates="storage_area", cascade="all, delete-orphan")


class GrainBatch(Base):
    __tablename__ = "grain_batches"

    id = Column(Integer, primary_key=True, index=True)
    storage_area_id = Column(Integer, ForeignKey("storage_areas.id"), nullable=False)
    grain_variety = Column(String(50), nullable=False)
    quantity_kg = Column(Float, nullable=False)
    remaining_kg = Column(Float, nullable=False)
    source = Column(String(200), nullable=False)
    inbound_date = Column(DateTime, default=datetime.utcnow)
    expiry_date = Column(DateTime, nullable=False)
    is_expired = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)

    storage_area = relationship("StorageArea", back_populates="batches")
    inbound_records = relationship("InboundRecord", back_populates="batch", cascade="all, delete-orphan")
    outbound_records = relationship("OutboundRecord", back_populates="batch", cascade="all, delete-orphan")


class InboundRecord(Base):
    __tablename__ = "inbound_records"

    id = Column(Integer, primary_key=True, index=True)
    batch_id = Column(Integer, ForeignKey("grain_batches.id"), nullable=False)
    storage_area_id = Column(Integer, ForeignKey("storage_areas.id"), nullable=False)
    grain_variety = Column(String(50), nullable=False)
    quantity_kg = Column(Float, nullable=False)
    source = Column(String(200), nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)

    batch = relationship("GrainBatch", back_populates="inbound_records")


class OutboundRecord(Base):
    __tablename__ = "outbound_records"

    id = Column(Integer, primary_key=True, index=True)
    batch_id = Column(Integer, ForeignKey("grain_batches.id"), nullable=False)
    storage_area_id = Column(Integer, ForeignKey("storage_areas.id"), nullable=False)
    grain_variety = Column(String(50), nullable=False)
    quantity_kg = Column(Float, nullable=False)
    destination = Column(String(200), nullable=False)
    purpose = Column(String(200), nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)

    batch = relationship("GrainBatch", back_populates="outbound_records")


class TemperatureRecord(Base):
    __tablename__ = "temperature_records"

    id = Column(Integer, primary_key=True, index=True)
    storage_area_id = Column(Integer, ForeignKey("storage_areas.id"), nullable=False)
    temperature = Column(Float, nullable=False)
    is_alert = Column(Boolean, default=False)
    recorded_at = Column(DateTime, default=datetime.utcnow)
    minute_key = Column(String(20), nullable=False)

    storage_area = relationship("StorageArea", back_populates="temperature_records")


class PestInspection(Base):
    __tablename__ = "pest_inspections"

    id = Column(Integer, primary_key=True, index=True)
    storage_area_id = Column(Integer, ForeignKey("storage_areas.id"), nullable=False)
    grain_variety = Column(String(50), nullable=False)
    pest_count_per_kg = Column(Float, nullable=False)
    inspection_result = Column(Text, nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)

    storage_area = relationship("StorageArea", back_populates="pest_inspections")


class Todo(Base):
    __tablename__ = "todos"

    id = Column(Integer, primary_key=True, index=True)
    title = Column(String(200), nullable=False)
    description = Column(Text)
    is_completed = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)
