from sqlalchemy import Column, Integer, String, Float, Date, DateTime, ForeignKey, Text, Boolean
from sqlalchemy.orm import relationship
from datetime import datetime

from server.database import Base


class Canal(Base):
    __tablename__ = "canals"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), nullable=False, index=True)
    code = Column(String(50), unique=True, index=True, nullable=False)
    level = Column(Integer, nullable=False)
    parent_id = Column(Integer, ForeignKey("canals.id"), nullable=True)
    design_flow = Column(Float, nullable=False)
    max_flow = Column(Float, nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    parent = relationship("Canal", remote_side=[id], backref="children")
    gates = relationship("Gate", back_populates="canal")
    water_plans = relationship("WaterPlan", back_populates="canal")
    water_usages = relationship("WaterUsage", back_populates="canal")
    dispatch_schemes = relationship("DispatchScheme", back_populates="canal")


class Gate(Base):
    __tablename__ = "gates"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), nullable=False)
    code = Column(String(50), unique=True, index=True, nullable=False)
    canal_id = Column(Integer, ForeignKey("canals.id"), nullable=False)
    max_flow = Column(Float, nullable=False)
    current_status = Column(String(20), default="closed")
    current_open_rate = Column(Float, default=0.0)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    canal = relationship("Canal", back_populates="gates")


class WaterPlan(Base):
    __tablename__ = "water_plans"

    id = Column(Integer, primary_key=True, index=True)
    canal_id = Column(Integer, ForeignKey("canals.id"), nullable=False)
    year = Column(Integer, nullable=False, index=True)
    initial_annual_quota = Column(Float, nullable=False)
    current_annual_quota = Column(Float, nullable=False)
    annual_used = Column(Float, default=0.0)
    status = Column(String(20), default="active")
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    canal = relationship("Canal", back_populates="water_plans")
    quarterly_quotas = relationship("QuarterlyQuota", back_populates="water_plan", cascade="all, delete-orphan")


class QuarterlyQuota(Base):
    __tablename__ = "quarterly_quotas"

    id = Column(Integer, primary_key=True, index=True)
    water_plan_id = Column(Integer, ForeignKey("water_plans.id"), nullable=False)
    quarter = Column(Integer, nullable=False)
    initial_quota = Column(Float, nullable=False)
    current_quota = Column(Float, nullable=False)
    used_amount = Column(Float, default=0.0)
    is_critical = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    water_plan = relationship("WaterPlan", back_populates="quarterly_quotas")


class WaterUsage(Base):
    __tablename__ = "water_usages"

    id = Column(Integer, primary_key=True, index=True)
    canal_id = Column(Integer, ForeignKey("canals.id"), nullable=False)
    year = Column(Integer, nullable=False, index=True)
    month = Column(Integer, nullable=False, index=True)
    usage_amount = Column(Float, default=0.0)
    quota_amount = Column(Float, default=0.0)
    is_over_quota = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    canal = relationship("Canal", back_populates="water_usages")


class DispatchScheme(Base):
    __tablename__ = "dispatch_schemes"

    id = Column(Integer, primary_key=True, index=True)
    date = Column(Date, nullable=False, index=True)
    canal_id = Column(Integer, ForeignKey("canals.id"), nullable=False)
    gate_id = Column(Integer, ForeignKey("gates.id"), nullable=True)
    target_flow = Column(Float, default=0.0)
    open_rate = Column(Float, default=0.0)
    reason = Column(Text)
    is_ecological = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)

    canal = relationship("Canal", back_populates="dispatch_schemes")


class WarningRecord(Base):
    __tablename__ = "warning_records"

    id = Column(Integer, primary_key=True, index=True)
    canal_id = Column(Integer, ForeignKey("canals.id"), nullable=False)
    year = Column(Integer, nullable=False)
    month = Column(Integer, nullable=False)
    type = Column(String(50), nullable=False)
    message = Column(Text, nullable=False)
    level = Column(String(20), default="warning")
    is_resolved = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)


class WaterUseCoefficient(Base):
    __tablename__ = "water_use_coefficients"

    id = Column(Integer, primary_key=True, index=True)
    canal_id = Column(Integer, ForeignKey("canals.id"), nullable=False)
    year = Column(Integer, nullable=False)
    quarter = Column(Integer, nullable=False)
    total_inflow = Column(Float, default=0.0)
    total_outflow = Column(Float, default=0.0)
    coefficient = Column(Float, default=0.0)
    created_at = Column(DateTime, default=datetime.utcnow)
