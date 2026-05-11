import enum
from sqlalchemy import Column, Integer, String, Float, DateTime, ForeignKey, Enum, Text, Boolean
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func
from .database import Base


class StationType(str, enum.Enum):
    SHORE = "shore"
    BUOY = "buoy"


class RedtideStatus(str, enum.Enum):
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

    id = Column(String, primary_key=True, index=True)
    name = Column(String, index=True)
    station_type = Column(Enum(StationType), nullable=False)
    location = Column(String)
    sea_area = Column(String, index=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())

    tide_measurements = relationship("TideMeasurement", back_populates="station")
    wave_measurements = relationship("WaveMeasurement", back_populates="station")
    water_temp_measurements = relationship("WaterTempMeasurement", back_populates="station")
    salinity_measurements = relationship("SalinityMeasurement", back_populates="station")
    do_measurements = relationship("DissolvedOxygenMeasurement", back_populates="station")
    chlorophyll_measurements = relationship("ChlorophyllMeasurement", back_populates="station")
    redtide_events = relationship("RedtideEvent", back_populates="station")


class TideMeasurement(Base):
    __tablename__ = "tide_measurements"

    id = Column(Integer, primary_key=True, index=True)
    station_id = Column(String, ForeignKey("stations.id"))
    measured_at = Column(DateTime(timezone=True), index=True)
    tide_level = Column(Float, nullable=False)

    station = relationship("Station", back_populates="tide_measurements")


class WaveMeasurement(Base):
    __tablename__ = "wave_measurements"

    id = Column(Integer, primary_key=True, index=True)
    station_id = Column(String, ForeignKey("stations.id"))
    measured_at = Column(DateTime(timezone=True), index=True)
    significant_wave_height = Column(Float, nullable=False)
    wave_period = Column(Float)
    wave_direction = Column(Float)

    station = relationship("Station", back_populates="wave_measurements")


class WaterTempMeasurement(Base):
    __tablename__ = "water_temp_measurements"

    id = Column(Integer, primary_key=True, index=True)
    station_id = Column(String, ForeignKey("stations.id"))
    measured_at = Column(DateTime(timezone=True), index=True)
    temperature = Column(Float, nullable=False)

    station = relationship("Station", back_populates="water_temp_measurements")


class SalinityMeasurement(Base):
    __tablename__ = "salinity_measurements"

    id = Column(Integer, primary_key=True, index=True)
    station_id = Column(String, ForeignKey("stations.id"))
    measured_at = Column(DateTime(timezone=True), index=True)
    salinity = Column(Float, nullable=False)

    station = relationship("Station", back_populates="salinity_measurements")


class DissolvedOxygenMeasurement(Base):
    __tablename__ = "do_measurements"

    id = Column(Integer, primary_key=True, index=True)
    station_id = Column(String, ForeignKey("stations.id"))
    measured_at = Column(DateTime(timezone=True), index=True)
    dissolved_oxygen = Column(Float, nullable=False)

    station = relationship("Station", back_populates="do_measurements")


class ChlorophyllMeasurement(Base):
    __tablename__ = "chlorophyll_measurements"

    id = Column(Integer, primary_key=True, index=True)
    station_id = Column(String, ForeignKey("stations.id"))
    measured_at = Column(DateTime(timezone=True), index=True)
    chlorophyll = Column(Float, nullable=False)

    station = relationship("Station", back_populates="chlorophyll_measurements")


class RedtideEvent(Base):
    __tablename__ = "redtide_events"

    id = Column(Integer, primary_key=True, index=True)
    station_id = Column(String, ForeignKey("stations.id"))
    status = Column(Enum(RedtideStatus), nullable=False, default=RedtideStatus.SUSPECTED)
    suspected_at = Column(DateTime(timezone=True), server_default=func.now())
    confirmed_at = Column(DateTime(timezone=True))
    published_at = Column(DateTime(timezone=True))
    resolved_at = Column(DateTime(timezone=True))
    severity_level = Column(String)
    description = Column(Text)

    station = relationship("Station", back_populates="redtide_events")
    notifications = relationship("Notification", back_populates="redtide_event")


class Farmer(Base):
    __tablename__ = "farmers"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String, nullable=False)
    phone = Column(String)
    email = Column(String)
    sea_area = Column(String, index=True)
    farm_name = Column(String)
    is_active = Column(Boolean, default=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())


class Notification(Base):
    __tablename__ = "notifications"

    id = Column(Integer, primary_key=True, index=True)
    redtide_event_id = Column(Integer, ForeignKey("redtide_events.id"))
    farmer_id = Column(Integer, ForeignKey("farmers.id"))
    message = Column(Text, nullable=False)
    sent_at = Column(DateTime(timezone=True), server_default=func.now())
    is_read = Column(Boolean, default=False)

    redtide_event = relationship("RedtideEvent", back_populates="notifications")
