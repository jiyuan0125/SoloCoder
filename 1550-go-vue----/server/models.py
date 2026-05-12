import enum
from datetime import datetime
from sqlalchemy import Column, Integer, String, Float, DateTime, ForeignKey, Boolean, Text, Enum
from sqlalchemy.orm import relationship

from .database import Base


class StationType(str, enum.Enum):
    COASTAL = "coastal"
    BUOY = "buoy"


class RedTideStatus(str, enum.Enum):
    NORMAL = "normal"
    SUSPECTED = "suspected"
    CONFIRMED = "confirmed"
    PUBLISHED = "published"
    MONITORING = "monitoring"
    RESOLVED = "resolved"


class WaveAlertLevel(str, enum.Enum):
    NONE = "none"
    BLUE = "blue"
    YELLOW = "yellow"
    ORANGE = "orange"
    RED = "red"


class Station(Base):
    __tablename__ = "stations"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), nullable=False)
    station_type = Column(Enum(StationType), nullable=False)
    latitude = Column(Float, nullable=True)
    longitude = Column(Float, nullable=True)
    description = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    tide_readings = relationship("TideReading", back_populates="station", cascade="all, delete-orphan")
    wave_readings = relationship("WaveReading", back_populates="station", cascade="all, delete-orphan")
    water_temp_readings = relationship("WaterTempReading", back_populates="station", cascade="all, delete-orphan")
    salinity_readings = relationship("SalinityReading", back_populates="station", cascade="all, delete-orphan")
    do_readings = relationship("DissolvedOxygenReading", back_populates="station", cascade="all, delete-orphan")
    chlorophyll_readings = relationship("ChlorophyllReading", back_populates="station", cascade="all, delete-orphan")
    red_tide_events = relationship("RedTideEvent", back_populates="station", cascade="all, delete-orphan")


class TideReading(Base):
    __tablename__ = "tide_readings"

    id = Column(Integer, primary_key=True, index=True)
    station_id = Column(Integer, ForeignKey("stations.id"), nullable=False)
    value = Column(Float, nullable=False)
    reading_time = Column(DateTime, nullable=False, index=True)
    created_at = Column(DateTime, default=datetime.utcnow)

    station = relationship("Station", back_populates="tide_readings")


class WaveReading(Base):
    __tablename__ = "wave_readings"

    id = Column(Integer, primary_key=True, index=True)
    station_id = Column(Integer, ForeignKey("stations.id"), nullable=False)
    significant_wave_height = Column(Float, nullable=False)
    reading_time = Column(DateTime, nullable=False, index=True)
    alert_level = Column(Enum(WaveAlertLevel), default=WaveAlertLevel.NONE)
    created_at = Column(DateTime, default=datetime.utcnow)

    station = relationship("Station", back_populates="wave_readings")


class WaterTempReading(Base):
    __tablename__ = "water_temp_readings"

    id = Column(Integer, primary_key=True, index=True)
    station_id = Column(Integer, ForeignKey("stations.id"), nullable=False)
    value = Column(Float, nullable=False)
    reading_time = Column(DateTime, nullable=False, index=True)
    created_at = Column(DateTime, default=datetime.utcnow)

    station = relationship("Station", back_populates="water_temp_readings")


class SalinityReading(Base):
    __tablename__ = "salinity_readings"

    id = Column(Integer, primary_key=True, index=True)
    station_id = Column(Integer, ForeignKey("stations.id"), nullable=False)
    value = Column(Float, nullable=False)
    reading_time = Column(DateTime, nullable=False, index=True)
    created_at = Column(DateTime, default=datetime.utcnow)

    station = relationship("Station", back_populates="salinity_readings")


class DissolvedOxygenReading(Base):
    __tablename__ = "do_readings"

    id = Column(Integer, primary_key=True, index=True)
    station_id = Column(Integer, ForeignKey("stations.id"), nullable=False)
    value = Column(Float, nullable=False)
    reading_time = Column(DateTime, nullable=False, index=True)
    created_at = Column(DateTime, default=datetime.utcnow)

    station = relationship("Station", back_populates="do_readings")


class ChlorophyllReading(Base):
    __tablename__ = "chlorophyll_readings"

    id = Column(Integer, primary_key=True, index=True)
    station_id = Column(Integer, ForeignKey("stations.id"), nullable=False)
    value = Column(Float, nullable=False)
    reading_time = Column(DateTime, nullable=False, index=True)
    is_anomaly = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)

    station = relationship("Station", back_populates="chlorophyll_readings")


class ChlorophyllBaseline(Base):
    __tablename__ = "chlorophyll_baselines"

    id = Column(Integer, primary_key=True, index=True)
    station_id = Column(Integer, ForeignKey("stations.id"), nullable=False, unique=True)
    mean = Column(Float, nullable=False)
    std_dev = Column(Float, nullable=False)
    threshold = Column(Float, nullable=False)
    sample_count = Column(Integer, default=0)
    last_updated = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)


class RedTideEvent(Base):
    __tablename__ = "red_tide_events"

    id = Column(Integer, primary_key=True, index=True)
    station_id = Column(Integer, ForeignKey("stations.id"), nullable=False)
    status = Column(Enum(RedTideStatus), default=RedTideStatus.SUSPECTED)
    suspected_at = Column(DateTime, default=datetime.utcnow)
    confirmed_at = Column(DateTime, nullable=True)
    published_at = Column(DateTime, nullable=True)
    resolved_at = Column(DateTime, nullable=True)
    confirmed_by = Column(String(100), nullable=True)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    station = relationship("Station", back_populates="red_tide_events")
    readings = relationship("RedTideReading", back_populates="event", cascade="all, delete-orphan")


class RedTideReading(Base):
    __tablename__ = "red_tide_readings"

    id = Column(Integer, primary_key=True, index=True)
    event_id = Column(Integer, ForeignKey("red_tide_events.id"), nullable=False)
    chlorophyll_value = Column(Float, nullable=False)
    do_value = Column(Float, nullable=True)
    reading_time = Column(DateTime, nullable=False, index=True)
    is_over_threshold = Column(Boolean, default=True)
    created_at = Column(DateTime, default=datetime.utcnow)

    event = relationship("RedTideEvent", back_populates="readings")


class Farmer(Base):
    __tablename__ = "farmers"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), nullable=False)
    contact = Column(String(200), nullable=True)
    email = Column(String(200), nullable=True)
    phone = Column(String(50), nullable=True)
    station_ids = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)


class Notification(Base):
    __tablename__ = "notifications"

    id = Column(Integer, primary_key=True, index=True)
    farmer_id = Column(Integer, ForeignKey("farmers.id"), nullable=True)
    station_id = Column(Integer, ForeignKey("stations.id"), nullable=True)
    event_id = Column(Integer, ForeignKey("red_tide_events.id"), nullable=True)
    message = Column(Text, nullable=False)
    notification_type = Column(String(50), nullable=False)
    sent_at = Column(DateTime, default=datetime.utcnow)
    is_read = Column(Boolean, default=False)
