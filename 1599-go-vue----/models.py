from sqlalchemy import Column, Integer, String, DateTime, Boolean, Text, ForeignKey, Enum
from sqlalchemy.orm import relationship
from datetime import datetime
from database import Base
import enum


class PlantStatus(str, enum.Enum):
    NORMAL = "正常"
    DORMANT = "休眠"
    INTRODUCTION_OBSERVATION = "引种观察"
    DEAD = "已死亡"


class ProtectionLevel(str, enum.Enum):
    NONE = "无"
    LEVEL_1 = "一级"
    LEVEL_2 = "二级"


class Plant(Base):
    __tablename__ = "plants"

    id = Column(Integer, primary_key=True, index=True)
    scientific_name = Column(String(255), nullable=False, index=True)
    family = Column(String(100), nullable=False, index=True)
    genus = Column(String(100), nullable=False, index=True)
    origin = Column(String(255), nullable=False)
    status = Column(Enum(PlantStatus), default=PlantStatus.NORMAL)
    protection_level = Column(Enum(ProtectionLevel), default=ProtectionLevel.NONE)
    location = Column(String(255), nullable=True)
    description = Column(Text, nullable=True)
    is_published = Column(Boolean, default=False)
    introduction_date = Column(DateTime, nullable=True)
    observation_end_date = Column(DateTime, nullable=True)
    previous_status = Column(Enum(PlantStatus), nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    observation_records = relationship("ObservationRecord", back_populates="plant", cascade="all, delete-orphan")
    exhibition_participations = relationship("ExhibitionParticipant", back_populates="plant")


class ObservationRecord(Base):
    __tablename__ = "observation_records"

    id = Column(Integer, primary_key=True, index=True)
    plant_id = Column(Integer, ForeignKey("plants.id"), nullable=False)
    record_date = Column(DateTime, default=datetime.utcnow)
    growth_condition = Column(String(500), nullable=False)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)

    plant = relationship("Plant", back_populates="observation_records")


class Exhibition(Base):
    __tablename__ = "exhibitions"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(255), nullable=False)
    description = Column(Text, nullable=True)
    start_date = Column(DateTime, nullable=False)
    end_date = Column(DateTime, nullable=True)
    is_active = Column(Boolean, default=True)
    created_at = Column(DateTime, default=datetime.utcnow)

    participants = relationship("ExhibitionParticipant", back_populates="exhibition", cascade="all, delete-orphan")


class ExhibitionParticipant(Base):
    __tablename__ = "exhibition_participants"

    id = Column(Integer, primary_key=True, index=True)
    exhibition_id = Column(Integer, ForeignKey("exhibitions.id"), nullable=False)
    plant_id = Column(Integer, ForeignKey("plants.id"), nullable=False)
    original_status = Column(Enum(PlantStatus), nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)

    exhibition = relationship("Exhibition", back_populates="participants")
    plant = relationship("Plant", back_populates="exhibition_participations")
