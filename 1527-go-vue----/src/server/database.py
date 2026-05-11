import os
from sqlalchemy import (
    create_engine,
    Column,
    Integer,
    String,
    Float,
    DateTime,
    Boolean,
    Enum as SAEnum,
    ForeignKey,
    Index,
)
from sqlalchemy.orm import declarative_base, sessionmaker, relationship
from datetime import datetime

from core.models import ShiftType, AlarmLevel, MetricType


DB_PATH = os.environ.get("MONITOR_DB_PATH", "./monitor.db")
DATABASE_URL = f"sqlite:///{DB_PATH}"

engine = create_engine(
    DATABASE_URL,
    connect_args={"check_same_thread": False},
    future=True,
)

SessionLocal = sessionmaker(
    autocommit=False,
    autoflush=False,
    bind=engine,
    future=True,
)

Base = declarative_base()


class MineDB(Base):
    __tablename__ = "mines"
    
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(255), nullable=False, unique=True)
    description = Column(String(500), nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    
    zones = relationship("MonitorZoneDB", back_populates="mine")


class MonitorZoneDB(Base):
    __tablename__ = "monitor_zones"
    
    id = Column(Integer, primary_key=True, index=True)
    mine_id = Column(Integer, ForeignKey("mines.id"), nullable=False, index=True)
    name = Column(String(255), nullable=False)
    description = Column(String(500), nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    
    mine = relationship("MineDB", back_populates="zones")
    thresholds = relationship("ThresholdConfigDB", back_populates="zone")
    readings = relationship("SensorReadingDB", back_populates="zone")
    alarms = relationship("AlarmEventDB", back_populates="zone")
    shift_stats = relationship("ShiftStatsDB", back_populates="zone")
    daily_reports = relationship("DailyReportDB", back_populates="zone")


class ThresholdConfigDB(Base):
    __tablename__ = "threshold_configs"
    
    id = Column(Integer, primary_key=True, index=True)
    zone_id = Column(Integer, ForeignKey("monitor_zones.id"), nullable=False, index=True)
    metric_type = Column(SAEnum(MetricType), nullable=False, index=True)
    level1_threshold = Column(Float, nullable=False)
    level2_threshold = Column(Float, nullable=False)
    level3_threshold = Column(Float, nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    
    zone = relationship("MonitorZoneDB", back_populates="thresholds")
    
    __table_args__ = (
        Index("uq_zone_metric", "zone_id", "metric_type", unique=True),
    )


class SensorReadingDB(Base):
    __tablename__ = "sensor_readings"
    
    id = Column(Integer, primary_key=True, index=True)
    zone_id = Column(Integer, ForeignKey("monitor_zones.id"), nullable=False, index=True)
    timestamp = Column(DateTime, nullable=False, index=True)
    methane = Column(Float, nullable=False)
    co = Column(Float, nullable=False)
    wind_speed = Column(Float, nullable=False)
    temperature = Column(Float, nullable=False)
    dust = Column(Float, nullable=False)
    shift_type = Column(SAEnum(ShiftType), nullable=True, index=True)
    
    zone = relationship("MonitorZoneDB", back_populates="readings")
    
    __table_args__ = (
        Index("idx_zone_time", "zone_id", "timestamp"),
    )


class AlarmEventDB(Base):
    __tablename__ = "alarm_events"
    
    id = Column(Integer, primary_key=True, index=True)
    zone_id = Column(Integer, ForeignKey("monitor_zones.id"), nullable=False, index=True)
    metric_type = Column(SAEnum(MetricType), nullable=False, index=True)
    alarm_level = Column(SAEnum(AlarmLevel), nullable=False)
    value = Column(Float, nullable=False)
    threshold = Column(Float, nullable=False)
    start_time = Column(DateTime, nullable=False, index=True)
    end_time = Column(DateTime, nullable=True)
    acknowledged = Column(Boolean, default=False)
    acknowledged_at = Column(DateTime, nullable=True)
    acknowledged_by = Column(String(255), nullable=True)
    
    zone = relationship("MonitorZoneDB", back_populates="alarms")


class ShiftStatsDB(Base):
    __tablename__ = "shift_stats"
    
    id = Column(Integer, primary_key=True, index=True)
    zone_id = Column(Integer, ForeignKey("monitor_zones.id"), nullable=False, index=True)
    shift_date = Column(String(10), nullable=False, index=True)
    shift_type = Column(SAEnum(ShiftType), nullable=False, index=True)
    methane_avg = Column(Float, nullable=True)
    methane_max = Column(Float, nullable=True)
    methane_over_count = Column(Integer, default=0)
    co_avg = Column(Float, nullable=True)
    co_max = Column(Float, nullable=True)
    co_over_count = Column(Integer, default=0)
    wind_speed_avg = Column(Float, nullable=True)
    wind_speed_max = Column(Float, nullable=True)
    wind_speed_over_count = Column(Integer, default=0)
    temperature_avg = Column(Float, nullable=True)
    temperature_max = Column(Float, nullable=True)
    temperature_over_count = Column(Integer, default=0)
    dust_avg = Column(Float, nullable=True)
    dust_max = Column(Float, nullable=True)
    dust_over_count = Column(Integer, default=0)
    reading_count = Column(Integer, default=0)
    created_at = Column(DateTime, default=datetime.utcnow)
    
    zone = relationship("MonitorZoneDB", back_populates="shift_stats")
    
    __table_args__ = (
        Index("uq_zone_shift", "zone_id", "shift_date", "shift_type", unique=True),
    )


class DailyReportDB(Base):
    __tablename__ = "daily_reports"
    
    id = Column(Integer, primary_key=True, index=True)
    zone_id = Column(Integer, ForeignKey("monitor_zones.id"), nullable=False, index=True)
    report_date = Column(String(10), nullable=False, index=True)
    methane_avg = Column(Float, nullable=True)
    methane_max = Column(Float, nullable=True)
    methane_over_count = Column(Integer, default=0)
    co_avg = Column(Float, nullable=True)
    co_max = Column(Float, nullable=True)
    co_over_count = Column(Integer, default=0)
    wind_speed_avg = Column(Float, nullable=True)
    wind_speed_max = Column(Float, nullable=True)
    wind_speed_over_count = Column(Integer, default=0)
    temperature_avg = Column(Float, nullable=True)
    temperature_max = Column(Float, nullable=True)
    temperature_over_count = Column(Integer, default=0)
    dust_avg = Column(Float, nullable=True)
    dust_max = Column(Float, nullable=True)
    dust_over_count = Column(Integer, default=0)
    alarm_count = Column(Integer, default=0)
    reading_count = Column(Integer, default=0)
    created_at = Column(DateTime, default=datetime.utcnow)
    
    zone = relationship("MonitorZoneDB", back_populates="daily_reports")
    
    __table_args__ = (
        Index("uq_zone_report_date", "zone_id", "report_date", unique=True),
    )


def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()


def init_db():
    Base.metadata.create_all(bind=engine)
