from datetime import datetime, date
from typing import Optional, List
from enum import Enum

from sqlalchemy import (
    Column, Integer, String, Float, Text, Date, DateTime,
    ForeignKey, Enum as SQLEnum
)
from sqlalchemy.orm import relationship, DeclarativeBase


class Base(DeclarativeBase):
    pass


class Severity(str, Enum):
    LIGHT = "light"
    MODERATE = "moderate"
    SEVERE = "severe"


class FireRiskType(str, Enum):
    DRY_VEGETATION = "dry_vegetation"
    SMOKE = "smoke"
    OPEN_FIRE = "open_fire"
    OTHER = "other"


class PestType(str, Enum):
    INSECT = "insect"
    DISEASE = "disease"
    ANIMAL = "animal"
    OTHER = "other"


class ProcessingStatus(str, Enum):
    PENDING = "pending"
    PROCESSING = "processing"
    COMPLETED = "completed"


class TodoStatus(str, Enum):
    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    DONE = "done"
    OVERDUE = "overdue"


class UserRole(str, Enum):
    ADMIN = "admin"
    RANGER = "ranger"
    AREA_MANAGER = "area_manager"


class User(Base):
    __tablename__ = "users"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), nullable=False)
    phone = Column(String(20), unique=True, nullable=False)
    role = Column(SQLEnum(UserRole), nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)

    assigned_areas = relationship("Area", back_populates="ranger", foreign_keys="Area.ranger_id")
    managed_areas = relationship("Area", back_populates="manager", foreign_keys="Area.manager_id")
    tasks = relationship("PatrolTask", back_populates="ranger")
    reports = relationship("PatrolReport", back_populates="ranger")
    todos = relationship("Todo", back_populates="assignee")


class Area(Base):
    __tablename__ = "areas"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), nullable=False)
    area_km2 = Column(Float, nullable=False)
    main_tree_species = Column(String(200), nullable=False)
    ranger_id = Column(Integer, ForeignKey("users.id"))
    manager_id = Column(Integer, ForeignKey("users.id"))
    created_at = Column(DateTime, default=datetime.utcnow)

    ranger = relationship("User", back_populates="assigned_areas", foreign_keys=[ranger_id])
    manager = relationship("User", back_populates="managed_areas", foreign_keys=[manager_id])
    tasks = relationship("PatrolTask", back_populates="area")
    reports = relationship("PatrolReport", back_populates="area")


class PatrolTask(Base):
    __tablename__ = "patrol_tasks"

    id = Column(Integer, primary_key=True, index=True)
    area_id = Column(Integer, ForeignKey("areas.id"), nullable=False)
    ranger_id = Column(Integer, ForeignKey("users.id"), nullable=False)
    patrol_date = Column(Date, nullable=False)
    route_waypoints = Column(Text, nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)

    area = relationship("Area", back_populates="tasks")
    ranger = relationship("User", back_populates="tasks")
    report = relationship("PatrolReport", back_populates="task", uselist=False)


class PatrolReport(Base):
    __tablename__ = "patrol_reports"

    id = Column(Integer, primary_key=True, index=True)
    task_id = Column(Integer, ForeignKey("patrol_tasks.id"), nullable=False, unique=True)
    area_id = Column(Integer, ForeignKey("areas.id"), nullable=False)
    ranger_id = Column(Integer, ForeignKey("users.id"), nullable=False)
    actual_route = Column(Text, nullable=False)
    is_qualified = Column(Integer, default=1)
    submitted_at = Column(DateTime, default=datetime.utcnow)

    task = relationship("PatrolTask", back_populates="report")
    area = relationship("Area", back_populates="reports")
    ranger = relationship("User", back_populates="reports")
    anomalies = relationship("Anomaly", back_populates="report")


class Anomaly(Base):
    __tablename__ = "anomalies"

    id = Column(Integer, primary_key=True, index=True)
    report_id = Column(Integer, ForeignKey("patrol_reports.id"), nullable=False)
    anomaly_type = Column(String(50), nullable=False)
    location_lat = Column(Float, nullable=False)
    location_lng = Column(Float, nullable=False)
    description = Column(Text)
    is_duplicate = Column(Integer, default=0)
    duplicate_of_id = Column(Integer, ForeignKey("anomalies.id"), nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)

    report = relationship("PatrolReport", back_populates="anomalies")
    pest_data = relationship("PestAnomaly", back_populates="anomaly", uselist=False)
    fire_risk_data = relationship("FireRiskAnomaly", back_populates="anomaly", uselist=False)
    todo = relationship("Todo", back_populates="anomaly", uselist=False)


class PestAnomaly(Base):
    __tablename__ = "pest_anomalies"

    id = Column(Integer, primary_key=True, index=True)
    anomaly_id = Column(Integer, ForeignKey("anomalies.id"), nullable=False, unique=True)
    pest_type = Column(SQLEnum(PestType), nullable=False)
    severity = Column(SQLEnum(Severity), nullable=False)
    trees_affected = Column(Integer, default=1)
    notes = Column(Text)

    anomaly = relationship("Anomaly", back_populates="pest_data")


class FireRiskAnomaly(Base):
    __tablename__ = "fire_risk_anomalies"

    id = Column(Integer, primary_key=True, index=True)
    anomaly_id = Column(Integer, ForeignKey("anomalies.id"), nullable=False, unique=True)
    risk_type = Column(SQLEnum(FireRiskType), nullable=False)
    status = Column(SQLEnum(ProcessingStatus), default=ProcessingStatus.PENDING)
    notes = Column(Text)

    anomaly = relationship("Anomaly", back_populates="fire_risk_data")


class Todo(Base):
    __tablename__ = "todos"

    id = Column(Integer, primary_key=True, index=True)
    anomaly_id = Column(Integer, ForeignKey("anomalies.id"), nullable=False, unique=True)
    assignee_id = Column(Integer, ForeignKey("users.id"), nullable=False)
    status = Column(SQLEnum(TodoStatus), default=TodoStatus.PENDING)
    priority = Column(Integer, default=1)
    due_date = Column(Date, nullable=False)
    notes = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)
    completed_at = Column(DateTime, nullable=True)

    anomaly = relationship("Anomaly", back_populates="todo")
    assignee = relationship("User", back_populates="todos")
