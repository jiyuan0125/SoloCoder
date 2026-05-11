from datetime import datetime
from sqlalchemy import Column, Integer, String, Float, DateTime, ForeignKey, Date, Time, Boolean
from sqlalchemy.orm import relationship

from .database import Base


class Pond(Base):
    __tablename__ = "ponds"

    id = Column(Integer, primary_key=True, index=True)
    code = Column(String(50), unique=True, index=True, nullable=False)
    area = Column(Float, nullable=False)
    species = Column(String(100), nullable=False)
    initial_stock = Column(Integer, nullable=False)
    current_stock = Column(Integer, nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)

    thresholds = relationship(
        "WaterThreshold", back_populates="pond", cascade="all, delete-orphan"
    )
    water_records = relationship(
        "WaterRecord", back_populates="pond", cascade="all, delete-orphan"
    )
    feeding_plans = relationship(
        "FeedingPlan", back_populates="pond", cascade="all, delete-orphan"
    )
    feeding_records = relationship(
        "FeedingRecord", back_populates="pond", cascade="all, delete-orphan"
    )
    harvest_records = relationship(
        "HarvestRecord", back_populates="pond", cascade="all, delete-orphan"
    )


class WaterThreshold(Base):
    __tablename__ = "water_thresholds"

    id = Column(Integer, primary_key=True, index=True)
    pond_id = Column(Integer, ForeignKey("ponds.id"), nullable=False)
    indicator = Column(String(50), nullable=False)
    min_value = Column(Float, nullable=True)
    max_value = Column(Float, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)

    pond = relationship("Pond", back_populates="thresholds")


class WaterRecord(Base):
    __tablename__ = "water_records"

    id = Column(Integer, primary_key=True, index=True)
    pond_id = Column(Integer, ForeignKey("ponds.id"), nullable=False)
    temperature = Column(Float, nullable=False)
    dissolved_oxygen = Column(Float, nullable=False)
    ph = Column(Float, nullable=False)
    ammonia_nitrogen = Column(Float, nullable=False)
    is_abnormal = Column(Boolean, default=False)
    abnormal_indicators = Column(String(500), nullable=True)
    recorded_at = Column(DateTime, default=datetime.utcnow)

    pond = relationship("Pond", back_populates="water_records")


class FeedingPlan(Base):
    __tablename__ = "feeding_plans"

    id = Column(Integer, primary_key=True, index=True)
    pond_id = Column(Integer, ForeignKey("ponds.id"), nullable=False)
    plan_date = Column(Date, nullable=False)
    plan_time = Column(Time, nullable=False)
    feed_amount = Column(Float, nullable=False)
    is_executed = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)

    pond = relationship("Pond", back_populates="feeding_plans")


class FeedingRecord(Base):
    __tablename__ = "feeding_records"

    id = Column(Integer, primary_key=True, index=True)
    pond_id = Column(Integer, ForeignKey("ponds.id"), nullable=False)
    plan_id = Column(Integer, ForeignKey("feeding_plans.id"), nullable=True)
    feed_amount = Column(Float, nullable=False)
    executed_at = Column(DateTime, default=datetime.utcnow)

    pond = relationship("Pond", back_populates="feeding_records")


class HarvestRecord(Base):
    __tablename__ = "harvest_records"

    id = Column(Integer, primary_key=True, index=True)
    pond_id = Column(Integer, ForeignKey("ponds.id"), nullable=False)
    species = Column(String(100), nullable=False)
    quantity = Column(Integer, nullable=False)
    weight = Column(Float, nullable=False)
    harvested_at = Column(DateTime, default=datetime.utcnow)

    pond = relationship("Pond", back_populates="harvest_records")
