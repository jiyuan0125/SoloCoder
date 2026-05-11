from datetime import date, datetime, timedelta
from typing import List, Optional
from sqlalchemy import create_engine, Column, Integer, String, Date, DateTime, ForeignKey, Float, Text, Boolean
from sqlalchemy.orm import declarative_base, sessionmaker, relationship, Session

Base = declarative_base()

DATABASE_URL = "sqlite:///./vet_clinic.db"

engine = create_engine(DATABASE_URL, connect_args={"check_same_thread": False})
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)


def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()


class Farmer(Base):
    __tablename__ = "farmers"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), nullable=False)
    address = Column(String(255), nullable=False)
    livestock_type = Column(String(50), nullable=False)
    scale = Column(Integer, nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    last_visit_date = Column(Date, nullable=True)

    visits = relationship("VisitRecord", back_populates="farmer")


class Medicine(Base):
    __tablename__ = "medicines"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), nullable=False, unique=True)
    specification = Column(String(100), nullable=False)
    stock = Column(Float, nullable=False, default=0)
    unit = Column(String(20), nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)

    prescriptions = relationship("Prescription", back_populates="medicine")


class Route(Base):
    __tablename__ = "routes"

    id = Column(Integer, primary_key=True, index=True)
    route_date = Column(Date, nullable=False, unique=True)
    is_generated = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)

    items = relationship("RouteItem", back_populates="route", order_by="RouteItem.order_index")


class RouteItem(Base):
    __tablename__ = "route_items"

    id = Column(Integer, primary_key=True, index=True)
    route_id = Column(Integer, ForeignKey("routes.id"), nullable=False)
    farmer_id = Column(Integer, ForeignKey("farmers.id"), nullable=False)
    order_index = Column(Integer, nullable=False)
    is_completed = Column(Boolean, default=False)

    route = relationship("Route", back_populates="items")
    farmer = relationship("Farmer")


class VisitRecord(Base):
    __tablename__ = "visit_records"

    id = Column(Integer, primary_key=True, index=True)
    farmer_id = Column(Integer, ForeignKey("farmers.id"), nullable=False)
    visit_date = Column(Date, nullable=False)
    has_abnormality = Column(Boolean, nullable=False)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)

    farmer = relationship("Farmer", back_populates="visits")
    cases = relationship("Case", back_populates="visit", cascade="all, delete-orphan")


class Case(Base):
    __tablename__ = "cases"

    id = Column(Integer, primary_key=True, index=True)
    visit_id = Column(Integer, ForeignKey("visit_records.id"), nullable=False)
    animal_type = Column(String(50), nullable=False)
    symptoms = Column(Text, nullable=False)
    diagnosis = Column(String(255), nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)

    visit = relationship("VisitRecord", back_populates="cases")
    prescriptions = relationship("Prescription", back_populates="case", cascade="all, delete-orphan")


class Prescription(Base):
    __tablename__ = "prescriptions"

    id = Column(Integer, primary_key=True, index=True)
    case_id = Column(Integer, ForeignKey("cases.id"), nullable=False)
    medicine_id = Column(Integer, ForeignKey("medicines.id"), nullable=False)
    dosage_per_day = Column(Float, nullable=False)
    treatment_days = Column(Integer, nullable=False)
    total_dosage = Column(Float, nullable=False)
    notes = Column(Text, nullable=True)
    start_date = Column(Date, nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)

    case = relationship("Case", back_populates="prescriptions")
    medicine = relationship("Medicine", back_populates="prescriptions")
    todos = relationship("Todo", back_populates="prescription", cascade="all, delete-orphan")


class Todo(Base):
    __tablename__ = "todos"

    id = Column(Integer, primary_key=True, index=True)
    prescription_id = Column(Integer, ForeignKey("prescriptions.id"), nullable=False)
    todo_date = Column(Date, nullable=False)
    is_completed = Column(Boolean, default=False)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)

    prescription = relationship("Prescription", back_populates="todos")


def init_db():
    Base.metadata.create_all(bind=engine)
