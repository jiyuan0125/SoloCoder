from datetime import date, datetime, timedelta
from enum import Enum
from typing import Optional, List
from sqlalchemy import (
    Column, Integer, String, Date, DateTime, Text, Boolean,
    ForeignKey, Enum as SQLEnum, Float
)
from sqlalchemy.orm import relationship, declarative_base, Session
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker

Base = declarative_base()

class CageStatus(str, Enum):
    FREE = "空闲"
    OCCUPIED = "占用"
    CLEANING = "清洁中"
    MAINTENANCE = "维修中"

class BoardingStatus(str, Enum):
    ACTIVE = "入住中"
    OVERDUE = "逾期"
    COMPLETED = "已退房"

class PetType(str, Enum):
    DOG = "狗"
    CAT = "猫"
    BIRD = "鸟"
    OTHER = "其他"

class TodoStatus(str, Enum):
    PENDING = "待处理"
    IN_PROGRESS = "处理中"
    DONE = "已完成"

class Zone(Base):
    __tablename__ = "zones"
    
    id = Column(Integer, primary_key=True, autoincrement=True)
    name = Column(String(100), nullable=False, unique=True)
    description = Column(String(500), nullable=True)
    created_at = Column(DateTime, default=datetime.now)
    
    cages = relationship("Cage", back_populates="zone", cascade="all, delete-orphan")
    
    def to_dict(self):
        return {
            "id": self.id,
            "name": self.name,
            "description": self.description,
            "created_at": self.created_at.isoformat() if self.created_at else None
        }

class Cage(Base):
    __tablename__ = "cages"
    
    id = Column(Integer, primary_key=True, autoincrement=True)
    zone_id = Column(Integer, ForeignKey("zones.id"), nullable=False)
    code = Column(String(50), nullable=False, unique=True)
    status = Column(SQLEnum(CageStatus), nullable=False, default=CageStatus.FREE)
    daily_rate = Column(Float, nullable=False, default=50.0)
    description = Column(String(500), nullable=True)
    created_at = Column(DateTime, default=datetime.now)
    
    zone = relationship("Zone", back_populates="cages")
    boarding_records = relationship("BoardingRecord", back_populates="cage")
    
    def to_dict(self):
        return {
            "id": self.id,
            "zone_id": self.zone_id,
            "zone_name": self.zone.name if self.zone else None,
            "code": self.code,
            "status": self.status.value,
            "daily_rate": self.daily_rate,
            "description": self.description,
            "created_at": self.created_at.isoformat() if self.created_at else None
        }

class Pet(Base):
    __tablename__ = "pets"
    
    id = Column(Integer, primary_key=True, autoincrement=True)
    name = Column(String(100), nullable=False)
    type = Column(SQLEnum(PetType), nullable=False)
    breed = Column(String(100), nullable=True)
    age = Column(Integer, nullable=True)
    owner_name = Column(String(100), nullable=False)
    owner_phone = Column(String(50), nullable=False)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.now)
    
    boarding_records = relationship("BoardingRecord", back_populates="pet")
    
    def to_dict(self):
        return {
            "id": self.id,
            "name": self.name,
            "type": self.type.value,
            "breed": self.breed,
            "age": self.age,
            "owner_name": self.owner_name,
            "owner_phone": self.owner_phone,
            "notes": self.notes,
            "created_at": self.created_at.isoformat() if self.created_at else None
        }

class BoardingRecord(Base):
    __tablename__ = "boarding_records"
    
    id = Column(Integer, primary_key=True, autoincrement=True)
    pet_id = Column(Integer, ForeignKey("pets.id"), nullable=False)
    cage_id = Column(Integer, ForeignKey("cages.id"), nullable=False)
    start_date = Column(Date, nullable=False)
    expected_end_date = Column(Date, nullable=False)
    actual_end_date = Column(Date, nullable=True)
    status = Column(SQLEnum(BoardingStatus), nullable=False, default=BoardingStatus.ACTIVE)
    service_fee = Column(Float, nullable=False, default=0.0)
    total_amount = Column(Float, nullable=True)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.now)
    
    pet = relationship("Pet", back_populates="boarding_records")
    cage = relationship("Cage", back_populates="boarding_records")
    feeding_logs = relationship("FeedingLog", back_populates="boarding", cascade="all, delete-orphan")
    health_data = relationship("HealthData", back_populates="boarding", cascade="all, delete-orphan")
    
    def to_dict(self):
        return {
            "id": self.id,
            "pet_id": self.pet_id,
            "pet_name": self.pet.name if self.pet else None,
            "cage_id": self.cage_id,
            "cage_code": self.cage.code if self.cage else None,
            "start_date": self.start_date.isoformat(),
            "expected_end_date": self.expected_end_date.isoformat(),
            "actual_end_date": self.actual_end_date.isoformat() if self.actual_end_date else None,
            "status": self.status.value,
            "service_fee": self.service_fee,
            "total_amount": self.total_amount,
            "notes": self.notes,
            "created_at": self.created_at.isoformat() if self.created_at else None
        }

class FeedingLog(Base):
    __tablename__ = "feeding_logs"
    
    id = Column(Integer, primary_key=True, autoincrement=True)
    boarding_id = Column(Integer, ForeignKey("boarding_records.id"), nullable=False)
    log_date = Column(Date, nullable=False, default=date.today)
    feeding_time = Column(String(50), nullable=True)
    food_type = Column(String(200), nullable=True)
    mental_state = Column(String(200), nullable=True)
    defecation = Column(String(200), nullable=True)
    abnormal_symptoms = Column(Text, nullable=True)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.now)
    
    boarding = relationship("BoardingRecord", back_populates="feeding_logs")
    
    def to_dict(self):
        return {
            "id": self.id,
            "boarding_id": self.boarding_id,
            "log_date": self.log_date.isoformat(),
            "feeding_time": self.feeding_time,
            "food_type": self.food_type,
            "mental_state": self.mental_state,
            "defecation": self.defecation,
            "abnormal_symptoms": self.abnormal_symptoms,
            "notes": self.notes,
            "created_at": self.created_at.isoformat() if self.created_at else None
        }

class HealthData(Base):
    __tablename__ = "health_data"
    
    id = Column(Integer, primary_key=True, autoincrement=True)
    boarding_id = Column(Integer, ForeignKey("boarding_records.id"), nullable=False)
    record_date = Column(Date, nullable=False, default=date.today)
    temperature = Column(Float, nullable=True)
    weight = Column(Float, nullable=True)
    heart_rate = Column(Integer, nullable=True)
    respiratory_rate = Column(Integer, nullable=True)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.now)
    
    boarding = relationship("BoardingRecord", back_populates="health_data")
    
    def to_dict(self):
        return {
            "id": self.id,
            "boarding_id": self.boarding_id,
            "record_date": self.record_date.isoformat(),
            "temperature": self.temperature,
            "weight": self.weight,
            "heart_rate": self.heart_rate,
            "respiratory_rate": self.respiratory_rate,
            "notes": self.notes,
            "created_at": self.created_at.isoformat() if self.created_at else None
        }

class Todo(Base):
    __tablename__ = "todos"
    
    id = Column(Integer, primary_key=True, autoincrement=True)
    boarding_id = Column(Integer, ForeignKey("boarding_records.id"), nullable=True)
    title = Column(String(200), nullable=False)
    description = Column(Text, nullable=True)
    status = Column(SQLEnum(TodoStatus), nullable=False, default=TodoStatus.PENDING)
    is_veterinary = Column(Boolean, nullable=False, default=False)
    due_date = Column(Date, nullable=True)
    created_at = Column(DateTime, default=datetime.now)
    
    def to_dict(self):
        return {
            "id": self.id,
            "boarding_id": self.boarding_id,
            "title": self.title,
            "description": self.description,
            "status": self.status.value,
            "is_veterinary": self.is_veterinary,
            "due_date": self.due_date.isoformat() if self.due_date else None,
            "created_at": self.created_at.isoformat() if self.created_at else None
        }

DATABASE_URL = "sqlite:///./pet_boarding.db"
engine = create_engine(DATABASE_URL, connect_args={"check_same_thread": False})
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

def init_db():
    Base.metadata.create_all(bind=engine)

def get_db() -> Session:
    db = SessionLocal()
    try:
        return db
    finally:
        db.close()
