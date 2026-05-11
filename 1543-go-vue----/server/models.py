from sqlalchemy import Column, Integer, String, Float, DateTime, Text, ForeignKey, Boolean, JSON
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func
from server.database import Base
from enum import Enum as PyEnum


class StationType(PyEnum):
    RAIN = "rain"
    WATER = "water"


class AlertLevel(PyEnum):
    NORMAL = "normal"
    RAIN_WARNING = "rain_warning"
    HEAVY_RAIN_WARNING = "heavy_rain_warning"
    FLOOD_WARNING = "flood_warning"


class DispatchStatus(PyEnum):
    PENDING = "pending"
    APPROVED = "approved"
    REJECTED = "rejected"
    EXECUTED = "executed"


class FloodPhaseType(PyEnum):
    WARNING = "warning"
    RISING = "rising"
    PEAK = "peak"
    RECEDING = "receding"
    ENDED = "ended"


class Watershed(Base):
    __tablename__ = "watersheds"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), unique=True, index=True)
    description = Column(Text, nullable=True)
    area = Column(Float, nullable=True, comment="流域面积，平方公里")
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now())

    reservoirs = relationship("Reservoir", back_populates="watershed")
    stations = relationship("MonitoringStation", back_populates="watershed")


class Reservoir(Base):
    __tablename__ = "reservoirs"

    id = Column(Integer, primary_key=True, index=True)
    watershed_id = Column(Integer, ForeignKey("watersheds.id"), index=True)
    name = Column(String(100), index=True)
    code = Column(String(50), unique=True, index=True)
    capacity = Column(Float, comment="总库容，立方米")
    flood_limit_level = Column(Float, comment="汛限水位，米")
    warning_level = Column(Float, comment="警戒水位，米")
    capacity_curve = Column(JSON, comment="库容曲线：水位-库容对应关系，如{\"30.0\": 1000000, \"31.0\": 2000000}")
    downstream_safe_discharge = Column(Float, comment="下游河道安全泄量，立方米/秒")
    current_level = Column(Float, default=0.0, comment="当前水位，米")
    current_discharge = Column(Float, default=0.0, comment="当前泄洪流量，立方米/秒")
    is_discharging = Column(Boolean, default=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now())

    watershed = relationship("Watershed", back_populates="reservoirs")
    operations = relationship("ReservoirOperation", back_populates="reservoir")


class MonitoringStation(Base):
    __tablename__ = "monitoring_stations"

    id = Column(Integer, primary_key=True, index=True)
    watershed_id = Column(Integer, ForeignKey("watersheds.id"), index=True)
    name = Column(String(100), index=True)
    code = Column(String(50), unique=True, index=True)
    station_type = Column(String(20), comment="rain或water")
    latitude = Column(Float, nullable=True)
    longitude = Column(Float, nullable=True)
    warning_threshold = Column(Float, nullable=True, comment="雨量站：6小时累计降雨量阈值(mm)，水位站：警戒水位(m)")
    is_active = Column(Boolean, default=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now())

    watershed = relationship("Watershed", back_populates="stations")
    data_records = relationship("HydrologicalData", back_populates="station")


class HydrologicalData(Base):
    __tablename__ = "hydrological_data"

    id = Column(Integer, primary_key=True, index=True)
    station_id = Column(Integer, ForeignKey("monitoring_stations.id"), index=True)
    value = Column(Float, comment="雨量(mm)或水位(m)")
    rainfall_6h = Column(Float, nullable=True, comment="6小时累计降雨量")
    recorded_at = Column(DateTime(timezone=True), index=True, server_default=func.now())
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    flood_event_id = Column(Integer, ForeignKey("flood_events.id"), nullable=True, index=True)

    station = relationship("MonitoringStation", back_populates="data_records")
    flood_event = relationship("FloodEvent", back_populates="hydrological_data")


class Alert(Base):
    __tablename__ = "alerts"

    id = Column(Integer, primary_key=True, index=True)
    station_id = Column(Integer, ForeignKey("monitoring_stations.id"), nullable=True, index=True)
    alert_level = Column(String(30))
    alert_message = Column(Text)
    trigger_value = Column(Float)
    threshold = Column(Float, nullable=True)
    triggered_at = Column(DateTime(timezone=True), server_default=func.now())
    flood_event_id = Column(Integer, ForeignKey("flood_events.id"), nullable=True, index=True)
    resolved_at = Column(DateTime(timezone=True), nullable=True)
    is_active = Column(Boolean, default=True)

    flood_event = relationship("FloodEvent", back_populates="alerts")


class DispatchRecommendation(Base):
    __tablename__ = "dispatch_recommendations"

    id = Column(Integer, primary_key=True, index=True)
    flood_event_id = Column(Integer, ForeignKey("flood_events.id"), nullable=True, index=True)
    reservoir_id = Column(Integer, ForeignKey("reservoirs.id"), nullable=True, index=True)
    recommended_discharge = Column(Float, comment="建议泄洪流量，立方米/秒")
    rationale = Column(Text, comment="调度建议理由")
    current_capacity = Column(Float, nullable=True, comment="当前库容")
    downstream_safe = Column(Float, nullable=True, comment="下游安全泄量")
    status = Column(String(20), default="pending", comment="pending/approved/rejected/executed")
    created_by = Column(String(100), default="system")
    approved_by = Column(String(100), nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    approved_at = Column(DateTime(timezone=True), nullable=True)
    executed_at = Column(DateTime(timezone=True), nullable=True)

    flood_event = relationship("FloodEvent", back_populates="dispatches")


class ReservoirOperation(Base):
    __tablename__ = "reservoir_operations"

    id = Column(Integer, primary_key=True, index=True)
    reservoir_id = Column(Integer, ForeignKey("reservoirs.id"), index=True)
    flood_event_id = Column(Integer, ForeignKey("flood_events.id"), nullable=True, index=True)
    dispatch_id = Column(Integer, ForeignKey("dispatch_recommendations.id"), nullable=True)
    previous_discharge = Column(Float)
    new_discharge = Column(Float)
    previous_level = Column(Float)
    reason = Column(Text)
    operated_by = Column(String(100), default="system")
    operated_at = Column(DateTime(timezone=True), server_default=func.now())

    reservoir = relationship("Reservoir", back_populates="operations")
    flood_event = relationship("FloodEvent", back_populates="operations")


class FloodPhase(Base):
    __tablename__ = "flood_phases"

    id = Column(Integer, primary_key=True, index=True)
    flood_event_id = Column(Integer, ForeignKey("flood_events.id"), index=True)
    phase_type = Column(String(30), comment="warning/rising/peak/receding/ended")
    description = Column(Text, nullable=True)
    started_at = Column(DateTime(timezone=True), server_default=func.now())
    ended_at = Column(DateTime(timezone=True), nullable=True)

    flood_event = relationship("FloodEvent", back_populates="phases")


class FloodEvent(Base):
    __tablename__ = "flood_events"

    id = Column(Integer, primary_key=True, index=True)
    watershed_id = Column(Integer, ForeignKey("watersheds.id"), nullable=True, index=True)
    title = Column(String(200))
    description = Column(Text, nullable=True)
    trigger_alert_id = Column(Integer, ForeignKey("alerts.id"), nullable=True)
    started_at = Column(DateTime(timezone=True), server_default=func.now())
    ended_at = Column(DateTime(timezone=True), nullable=True)
    is_active = Column(Boolean, default=True)
    summary = Column(Text, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())

    alerts = relationship("Alert", back_populates="flood_event")
    dispatches = relationship("DispatchRecommendation", back_populates="flood_event")
    operations = relationship("ReservoirOperation", back_populates="flood_event")
    phases = relationship("FloodPhase", back_populates="flood_event", order_by="FloodPhase.started_at")
    hydrological_data = relationship("HydrologicalData", back_populates="flood_event")
