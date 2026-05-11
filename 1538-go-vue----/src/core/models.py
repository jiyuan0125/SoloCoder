from datetime import datetime, date
from sqlalchemy import (
    Column, String, Integer, Float, DateTime, Date, Text, ForeignKey, Boolean
)
from sqlalchemy.orm import relationship

from src.core.database import Base
from src.core.enums import WasteType, WaybillStatus, UnitType


class Unit(Base):
    __tablename__ = "units"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(255), nullable=False, index=True)
    unit_type = Column(String(50), nullable=False)
    address = Column(String(500), nullable=True)
    contact_person = Column(String(100), nullable=True)
    contact_phone = Column(String(20), nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    waste_ledgers = relationship("WasteLedger", back_populates="producer")
    qualifications = relationship("Qualification", back_populates="disposer")
    outgoing_waybills = relationship(
        "Waybill",
        foreign_keys="Waybill.producer_id",
        back_populates="producer"
    )
    incoming_waybills = relationship(
        "Waybill",
        foreign_keys="Waybill.disposer_id",
        back_populates="disposer"
    )


class Qualification(Base):
    __tablename__ = "qualifications"

    id = Column(Integer, primary_key=True, index=True)
    disposer_id = Column(Integer, ForeignKey("units.id"), nullable=False, index=True)
    hw_codes = Column(String(1000), nullable=False)
    valid_from = Column(Date, nullable=False)
    valid_until = Column(Date, nullable=False)
    is_active = Column(Boolean, default=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    disposer = relationship("Unit", back_populates="qualifications")


class WasteLedger(Base):
    __tablename__ = "waste_ledgers"

    id = Column(Integer, primary_key=True, index=True)
    producer_id = Column(Integer, ForeignKey("units.id"), nullable=False, index=True)
    waste_type = Column(String(50), nullable=False)
    hw_code = Column(String(50), nullable=True, index=True)
    waste_name = Column(String(255), nullable=False)
    quantity = Column(Float, nullable=False)
    unit = Column(String(20), default="吨")
    production_date = Column(Date, nullable=False, index=True)
    remarks = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    producer = relationship("Unit", back_populates="waste_ledgers")


class Waybill(Base):
    __tablename__ = "waybills"

    id = Column(Integer, primary_key=True, index=True)
    waybill_code = Column(String(100), nullable=False, unique=True, index=True)
    producer_id = Column(Integer, ForeignKey("units.id"), nullable=False, index=True)
    disposer_id = Column(Integer, ForeignKey("units.id"), nullable=False, index=True)
    waste_type = Column(String(50), nullable=False)
    hw_code = Column(String(50), nullable=True)
    waste_name = Column(String(255), nullable=False)
    quantity = Column(Float, nullable=False)
    unit = Column(String(20), default="吨")
    status = Column(String(50), nullable=False)
    reject_reason = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    producer = relationship(
        "Unit",
        foreign_keys=[producer_id],
        back_populates="outgoing_waybills"
    )
    disposer = relationship(
        "Unit",
        foreign_keys=[disposer_id],
        back_populates="incoming_waybills"
    )


class MonthlyLedger(Base):
    __tablename__ = "monthly_ledgers"

    id = Column(Integer, primary_key=True, index=True)
    unit_id = Column(Integer, ForeignKey("units.id"), nullable=False, index=True)
    year = Column(Integer, nullable=False, index=True)
    month = Column(Integer, nullable=False, index=True)
    opening_balance = Column(Float, nullable=False, default=0.0)
    production = Column(Float, nullable=False, default=0.0)
    transfer_out = Column(Float, nullable=False, default=0.0)
    closing_balance = Column(Float, nullable=False, default=0.0)
    calculated_closing = Column(Float, nullable=False, default=0.0)
    difference = Column(Float, nullable=False, default=0.0)
    is_approved = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)


class Alert(Base):
    __tablename__ = "alerts"

    id = Column(Integer, primary_key=True, index=True)
    unit_id = Column(Integer, ForeignKey("units.id"), nullable=False, index=True)
    alert_type = Column(String(100), nullable=False)
    message = Column(Text, nullable=False)
    is_resolved = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)
