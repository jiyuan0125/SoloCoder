from datetime import datetime, date
from sqlalchemy import (
    Column, Integer, String, Date, DateTime, ForeignKey, Boolean, Text, Enum
)
from sqlalchemy.orm import relationship

from .database import Base


class Vessel(Base):
    __tablename__ = "vessels"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), nullable=False)
    imo_number = Column(String(20), unique=True, index=True)
    flag = Column(String(50))
    gross_tonnage = Column(Integer)
    built_year = Column(Integer)
    vessel_type = Column(String(50))
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    inspections = relationship("Inspection", back_populates="vessel")
    certificates = relationship("Certificate", back_populates="vessel")
    dockings = relationship("Docking", back_populates="vessel")
    todo_items = relationship("TodoItem", back_populates="vessel")
    operation_records = relationship("OperationRecord", back_populates="vessel")


class Inspection(Base):
    __tablename__ = "inspections"

    id = Column(Integer, primary_key=True, index=True)
    vessel_id = Column(Integer, ForeignKey("vessels.id"), nullable=False)
    inspection_type = Column(String(20), nullable=False)
    inspection_date = Column(Date, nullable=False)
    inspector = Column(String(100))
    result = Column(String(50))
    remarks = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    vessel = relationship("Vessel", back_populates="inspections")
    certificate = relationship("Certificate", back_populates="inspection", uselist=False)


class Certificate(Base):
    __tablename__ = "certificates"

    id = Column(Integer, primary_key=True, index=True)
    vessel_id = Column(Integer, ForeignKey("vessels.id"), nullable=False)
    inspection_id = Column(Integer, ForeignKey("inspections.id"))
    certificate_type = Column(String(100), nullable=False)
    certificate_number = Column(String(50))
    issue_date = Column(Date, nullable=False)
    expiry_date = Column(Date, nullable=False)
    issued_by = Column(String(100))
    status = Column(String(20), default="valid")
    remarks = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    vessel = relationship("Vessel", back_populates="certificates")
    inspection = relationship("Inspection", back_populates="certificate")


class Dock(Base):
    __tablename__ = "docks"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), nullable=False)
    location = Column(String(100))
    capacity = Column(String(50))
    created_at = Column(DateTime, default=datetime.utcnow)

    dockings = relationship("Docking", back_populates="dock")
    maintenances = relationship("DockMaintenance", back_populates="dock")


class DockMaintenance(Base):
    __tablename__ = "dock_maintenances"

    id = Column(Integer, primary_key=True, index=True)
    dock_id = Column(Integer, ForeignKey("docks.id"), nullable=False)
    start_date = Column(Date, nullable=False)
    end_date = Column(Date, nullable=False)
    description = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)

    dock = relationship("Dock", back_populates="maintenances")


class Docking(Base):
    __tablename__ = "dockings"

    id = Column(Integer, primary_key=True, index=True)
    vessel_id = Column(Integer, ForeignKey("vessels.id"), nullable=False)
    dock_id = Column(Integer, ForeignKey("docks.id"), nullable=False)
    start_date = Column(Date, nullable=False)
    end_date = Column(Date, nullable=False)
    purpose = Column(String(200))
    status = Column(String(20), default="scheduled")
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    vessel = relationship("Vessel", back_populates="dockings")
    dock = relationship("Dock", back_populates="dockings")


class TodoItem(Base):
    __tablename__ = "todo_items"

    id = Column(Integer, primary_key=True, index=True)
    vessel_id = Column(Integer, ForeignKey("vessels.id"), nullable=False)
    todo_type = Column(String(50), nullable=False)
    title = Column(String(200), nullable=False)
    inspection_type = Column(String(20))
    due_date = Column(Date, nullable=False)
    arrange_deadline = Column(Date)
    status = Column(String(20), default="pending")
    reminder_level = Column(Integer)
    related_certificate_id = Column(Integer, ForeignKey("certificates.id"))
    remarks = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    vessel = relationship("Vessel", back_populates="todo_items")


class Reminder(Base):
    __tablename__ = "reminders"

    id = Column(Integer, primary_key=True, index=True)
    vessel_id = Column(Integer, ForeignKey("vessels.id"), nullable=False)
    certificate_id = Column(Integer, ForeignKey("certificates.id"), nullable=False)
    reminder_type = Column(String(50), nullable=False)
    days_before_expiry = Column(Integer, nullable=False)
    message = Column(Text)
    sent_date = Column(Date)
    is_read = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)


class OperationRecord(Base):
    __tablename__ = "operation_records"

    id = Column(Integer, primary_key=True, index=True)
    vessel_id = Column(Integer, ForeignKey("vessels.id"), nullable=False)
    start_date = Column(Date, nullable=False)
    end_date = Column(Date)
    route = Column(String(200))
    status = Column(String(20), default="active")
    is_illegal = Column(Boolean, default=False)
    remarks = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    vessel = relationship("Vessel", back_populates="operation_records")
