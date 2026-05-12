from sqlalchemy import create_engine, Column, Integer, String, Float, DateTime, Boolean, ForeignKey
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker, relationship
from datetime import datetime

SQLALCHEMY_DATABASE_URL = "sqlite:///./railway_station.db"

engine = create_engine(
    SQLALCHEMY_DATABASE_URL, connect_args={"check_same_thread": False}
)
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

Base = declarative_base()

class Station(Base):
    __tablename__ = "stations"
    
    id = Column(Integer, primary_key=True, index=True)
    code = Column(String(20), unique=True, index=True, nullable=False)
    name = Column(String(100), nullable=False)
    max_capacity = Column(Integer, default=5000)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    
    zones = relationship("Zone", back_populates="station", cascade="all, delete-orphan")
    trains = relationship("Train", back_populates="station", cascade="all, delete-orphan")

class Zone(Base):
    __tablename__ = "zones"
    
    id = Column(Integer, primary_key=True, index=True)
    station_id = Column(Integer, ForeignKey("stations.id"), nullable=False)
    name = Column(String(50), nullable=False)
    zone_type = Column(String(30), nullable=False)
    max_capacity = Column(Integer, default=1000)
    current_count = Column(Integer, default=0)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    
    station = relationship("Station", back_populates="zones")
    gates = relationship("SecurityGate", back_populates="zone", cascade="all, delete-orphan")
    passenger_counts = relationship("PassengerCount", back_populates="zone", cascade="all, delete-orphan")

class SecurityGate(Base):
    __tablename__ = "security_gates"
    
    id = Column(Integer, primary_key=True, index=True)
    zone_id = Column(Integer, ForeignKey("zones.id"), nullable=False)
    gate_number = Column(String(20), nullable=False)
    is_open = Column(Boolean, default=False)
    is_faulty = Column(Boolean, default=False)
    queue_length = Column(Integer, default=0)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    
    zone = relationship("Zone", back_populates="gates")

class Train(Base):
    __tablename__ = "trains"
    
    id = Column(Integer, primary_key=True, index=True)
    station_id = Column(Integer, ForeignKey("stations.id"), nullable=False)
    train_number = Column(String(20), unique=True, index=True, nullable=False)
    platform = Column(String(20), nullable=False)
    departure_time = Column(DateTime, nullable=False)
    checkin_start_time = Column(DateTime, nullable=False)
    checkin_end_time = Column(DateTime, nullable=False)
    is_checkin_active = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    
    station = relationship("Station", back_populates="trains")

class PassengerCount(Base):
    __tablename__ = "passenger_counts"
    
    id = Column(Integer, primary_key=True, index=True)
    zone_id = Column(Integer, ForeignKey("zones.id"), nullable=False)
    count = Column(Integer, default=0)
    timestamp = Column(DateTime, default=datetime.utcnow, index=True)
    count_type = Column(String(20), default="realtime")
    
    zone = relationship("Zone", back_populates="passenger_counts")

class Alert(Base):
    __tablename__ = "alerts"
    
    id = Column(Integer, primary_key=True, index=True)
    station_id = Column(Integer, ForeignKey("stations.id"), nullable=False)
    zone_id = Column(Integer, ForeignKey("zones.id"), nullable=True)
    alert_type = Column(String(50), nullable=False)
    severity = Column(String(20), nullable=False)
    message = Column(String(255), nullable=False)
    is_active = Column(Boolean, default=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    resolved_at = Column(DateTime, nullable=True)

class HolidayForecast(Base):
    __tablename__ = "holiday_forecasts"
    
    id = Column(Integer, primary_key=True, index=True)
    station_id = Column(Integer, ForeignKey("stations.id"), nullable=False)
    forecast_date = Column(DateTime, nullable=False)
    predicted_flow = Column(Float, default=0.0)
    normal_flow = Column(Float, default=0.0)
    is_peak_alert = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)

def init_db():
    Base.metadata.create_all(bind=engine)

def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()
