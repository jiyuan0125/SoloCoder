from sqlalchemy import Column, DateTime, Float, ForeignKey, Integer, String
from sqlalchemy.orm import relationship

from server.core.database import Base

from .base import BaseModel


class LightRecord(Base, BaseModel):
    __tablename__ = "light_records"

    lighthouse_id = Column(Integer, ForeignKey("lighthouses.id"), nullable=False)
    timestamp = Column(DateTime, nullable=False)
    illuminance = Column(Float, nullable=False)
    is_on = Column(Integer, default=1)
    status = Column(String(20), default="normal")

    lighthouse = relationship("Lighthouse")


class EnergyRecord(Base, BaseModel):
    __tablename__ = "energy_records"

    lighthouse_id = Column(Integer, ForeignKey("lighthouses.id"), nullable=False)
    timestamp = Column(DateTime, nullable=False)
    solar_generation = Column(Float, nullable=False)
    battery_level = Column(Float, nullable=False)
    battery_percent = Column(Float, nullable=False)

    lighthouse = relationship("Lighthouse")


class Alarm(Base, BaseModel):
    __tablename__ = "alarms"

    lighthouse_id = Column(Integer, ForeignKey("lighthouses.id"), nullable=False)
    alarm_type = Column(String(50), nullable=False)
    severity = Column(String(20), nullable=False)
    message = Column(String(255), nullable=False)
    status = Column(String(20), default="open")
    resolved_at = Column(DateTime, nullable=True)
    resolved_by = Column(String(100), nullable=True)

    lighthouse = relationship("Lighthouse")


class WorkOrder(Base, BaseModel):
    __tablename__ = "work_orders"

    lighthouse_id = Column(Integer, ForeignKey("lighthouses.id"), nullable=False)
    order_type = Column(String(50), nullable=False)
    priority = Column(String(20), nullable=False)
    title = Column(String(200), nullable=False)
    description = Column(String(1000), nullable=True)
    status = Column(String(20), default="created")
    assigned_to = Column(String(100), nullable=True)
    assigned_at = Column(DateTime, nullable=True)
    executed_at = Column(DateTime, nullable=True)
    executed_by = Column(String(100), nullable=True)
    execution_notes = Column(String(1000), nullable=True)
    accepted_at = Column(DateTime, nullable=True)
    accepted_by = Column(String(100), nullable=True)
    acceptance_result = Column(String(20), nullable=True)

    lighthouse = relationship("Lighthouse")


class MaintenancePlan(Base, BaseModel):
    __tablename__ = "maintenance_plans"

    lighthouse_id = Column(Integer, ForeignKey("lighthouses.id"), nullable=False)
    plan_type = Column(String(50), nullable=False)
    frequency_days = Column(Integer, nullable=False)
    last_execution_date = Column(DateTime, nullable=True)
    next_execution_date = Column(DateTime, nullable=True)
    title = Column(String(200), nullable=False)
    description = Column(String(1000), nullable=True)
    is_active = Column(Integer, default=1)

    lighthouse = relationship("Lighthouse")
