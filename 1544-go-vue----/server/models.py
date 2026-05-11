from datetime import datetime
from enum import Enum as PyEnum
from sqlalchemy import Column, Integer, String, Float, DateTime, ForeignKey, Boolean, Enum, Index
from sqlalchemy.orm import relationship, DeclarativeBase


class Base(DeclarativeBase):
    pass


class StationType(PyEnum):
    WATER_LEVEL = "water_level"
    RAINFALL = "rainfall"
    FLOW = "flow"


class DataQualityStatus(PyEnum):
    VALID = "valid"
    SUSPICIOUS = "suspicious"
    INVALID = "invalid"


class AggregationGranularity(PyEnum):
    HOURLY = "hourly"
    DAILY = "daily"


class AggregationMissingStatus(PyEnum):
    COMPLETE = "complete"
    MISSING = "missing"


class Station(Base):
    __tablename__ = "stations"

    id = Column(Integer, primary_key=True, autoincrement=True)
    name = Column(String(100), nullable=False)
    station_type = Column(Enum(StationType), nullable=False)
    device_model = Column(String(100), nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    calibrations = relationship("Calibration", back_populates="station", cascade="all, delete-orphan")
    raw_data = relationship("RawData", back_populates="station", cascade="all, delete-orphan")
    aggregated_data = relationship("AggregatedData", back_populates="station", cascade="all, delete-orphan")
    quality_records = relationship("QualityRecord", back_populates="station", cascade="all, delete-orphan")


class Calibration(Base):
    __tablename__ = "calibrations"

    id = Column(Integer, primary_key=True, autoincrement=True)
    station_id = Column(Integer, ForeignKey("stations.id"), nullable=False)
    last_calibration_date = Column(DateTime, nullable=False)
    next_calibration_date = Column(DateTime, nullable=False)
    is_done = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)

    station = relationship("Station", back_populates="calibrations")


class RawData(Base):
    __tablename__ = "raw_data"

    id = Column(Integer, primary_key=True, autoincrement=True)
    station_id = Column(Integer, ForeignKey("stations.id"), nullable=False)
    data_type = Column(Enum(StationType), nullable=False)
    timestamp = Column(DateTime, nullable=False)
    value = Column(Float, nullable=False)
    quality_status = Column(Enum(DataQualityStatus), default=DataQualityStatus.VALID, nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    station = relationship("Station", back_populates="raw_data")
    quality_records = relationship("QualityRecord", back_populates="raw_data", cascade="all, delete-orphan")

    __table_args__ = (
        Index('ix_raw_data_station_timestamp', 'station_id', 'timestamp'),
    )


class QualityRecord(Base):
    __tablename__ = "quality_records"

    id = Column(Integer, primary_key=True, autoincrement=True)
    station_id = Column(Integer, ForeignKey("stations.id"), nullable=False)
    raw_data_id = Column(Integer, ForeignKey("raw_data.id"), nullable=True)
    issue_type = Column(String(100), nullable=False)
    description = Column(String(255), nullable=False)
    detected_at = Column(DateTime, default=datetime.utcnow)
    reviewed = Column(Boolean, default=False)
    reviewer = Column(String(100), nullable=True)
    review_notes = Column(String(255), nullable=True)
    reviewed_at = Column(DateTime, nullable=True)

    station = relationship("Station", back_populates="quality_records")
    raw_data = relationship("RawData", back_populates="quality_records")


class AggregatedData(Base):
    __tablename__ = "aggregated_data"

    id = Column(Integer, primary_key=True, autoincrement=True)
    station_id = Column(Integer, ForeignKey("stations.id"), nullable=False)
    data_type = Column(Enum(StationType), nullable=False)
    granularity = Column(Enum(AggregationGranularity), nullable=False)
    period_start = Column(DateTime, nullable=False)
    period_end = Column(DateTime, nullable=False)
    average_value = Column(Float, nullable=True)
    max_value = Column(Float, nullable=True)
    min_value = Column(Float, nullable=True)
    missing_status = Column(Enum(AggregationMissingStatus), default=AggregationMissingStatus.COMPLETE, nullable=False)
    sample_count = Column(Integer, default=0, nullable=False)
    device_model = Column(String(100), nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    station = relationship("Station", back_populates="aggregated_data")

    __table_args__ = (
        Index('ix_aggregated_data_station_period', 'station_id', 'data_type', 'granularity', 'period_start'),
    )
