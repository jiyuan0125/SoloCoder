from datetime import date, datetime
from sqlalchemy import (
    Column,
    Integer,
    String,
    Float,
    Date,
    DateTime,
    ForeignKey,
    Text,
    Boolean,
    Enum,
)
from sqlalchemy.orm import relationship
from server.database import Base
import enum


class ProjectStatus(str, enum.Enum):
    INITIATED = "initiated"
    IMPLEMENTING = "implementing"
    ACCEPTING = "accepting"
    RECTIFYING = "rectifying"
    COMPLETED = "completed"
    REJECTED = "rejected"


class MeasureType(str, enum.Enum):
    ENGINEERING = "engineering"
    PLANT = "plant"
    FARMING = "farming"


class TodoType(str, enum.Enum):
    LAG = "lag"
    OVERRUN = "overrun"
    RECTIFY = "rectify"


class AcceptanceResult(str, enum.Enum):
    PASS = "pass"
    RECTIFY = "rectify"
    REJECT = "reject"


class Project(Base):
    __tablename__ = "projects"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(255), nullable=False)
    code = Column(String(100), unique=True, index=True)
    description = Column(Text)
    location = Column(String(500))
    status = Column(Enum(ProjectStatus), default=ProjectStatus.INITIATED)

    start_date = Column(Date, nullable=False)
    end_date = Column(Date, nullable=False)

    total_budget = Column(Float, default=0.0)
    initial_total_budget = Column(Float, default=0.0)

    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    measures = relationship("Measure", back_populates="project", cascade="all, delete-orphan")
    monitoring = relationship("MonitoringRecord", back_populates="project", cascade="all, delete-orphan")
    acceptances = relationship("Acceptance", back_populates="project", cascade="all, delete-orphan")
    budgets = relationship("BudgetItem", back_populates="project", cascade="all, delete-orphan")
    expenditures = relationship("Expenditure", back_populates="project", cascade="all, delete-orphan")
    todos = relationship("Todo", back_populates="project", cascade="all, delete-orphan")


class Measure(Base):
    __tablename__ = "measures"

    id = Column(Integer, primary_key=True, index=True)
    project_id = Column(Integer, ForeignKey("projects.id"), nullable=False)
    measure_type = Column(Enum(MeasureType), nullable=False)
    name = Column(String(255), nullable=False)
    description = Column(Text)

    planned_quantity = Column(Float, default=0.0)
    unit = Column(String(50))

    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    project = relationship("Project", back_populates="measures")
    progress_records = relationship("MeasureProgress", back_populates="measure", cascade="all, delete-orphan")


class MeasureProgress(Base):
    __tablename__ = "measure_progress"

    id = Column(Integer, primary_key=True, index=True)
    measure_id = Column(Integer, ForeignKey("measures.id"), nullable=False)
    year = Column(Integer, nullable=False)
    month = Column(Integer, nullable=False)

    completed_quantity = Column(Float, default=0.0)
    progress_percent = Column(Float, default=0.0)

    recorded_at = Column(Date, default=date.today)
    notes = Column(Text)

    measure = relationship("Measure", back_populates="progress_records")


class MonitoringRecord(Base):
    __tablename__ = "monitoring"

    id = Column(Integer, primary_key=True, index=True)
    project_id = Column(Integer, ForeignKey("projects.id"), nullable=False)

    year = Column(Integer, nullable=False)
    month = Column(Integer, nullable=False)
    quarter = Column(Integer, nullable=False)

    erosion_modulus = Column(Float, default=0.0)
    vegetation_coverage = Column(Float, default=0.0)

    notes = Column(Text)
    recorded_at = Column(Date, default=date.today)

    project = relationship("Project", back_populates="monitoring")


class BudgetItem(Base):
    __tablename__ = "budget_items"

    id = Column(Integer, primary_key=True, index=True)
    project_id = Column(Integer, ForeignKey("projects.id"), nullable=False)
    measure_type = Column(Enum(MeasureType), nullable=False)

    budget_amount = Column(Float, default=0.0)
    original_budget = Column(Float, default=0.0)
    budget_ratio = Column(Float, default=0.0)

    is_completed = Column(Boolean, default=False)
    acceptance_id = Column(Integer, ForeignKey("acceptances.id"), nullable=True)

    project = relationship("Project", back_populates="budgets")
    expenditures = relationship("Expenditure", back_populates="budget_item", cascade="all, delete-orphan")


class Expenditure(Base):
    __tablename__ = "expenditures"

    id = Column(Integer, primary_key=True, index=True)
    project_id = Column(Integer, ForeignKey("projects.id"), nullable=False)
    budget_item_id = Column(Integer, ForeignKey("budget_items.id"), nullable=False)

    amount = Column(Float, default=0.0)
    description = Column(String(500))
    expenditure_date = Column(Date, default=date.today)

    created_at = Column(DateTime, default=datetime.utcnow)

    project = relationship("Project", back_populates="expenditures")
    budget_item = relationship("BudgetItem", back_populates="expenditures")


class Acceptance(Base):
    __tablename__ = "acceptances"

    id = Column(Integer, primary_key=True, index=True)
    project_id = Column(Integer, ForeignKey("projects.id"), nullable=False)

    engineering_score = Column(Float, default=0.0)
    plant_score = Column(Float, default=0.0)
    farming_score = Column(Float, default=0.0)
    temporary_score = Column(Float, default=0.0)

    weighted_score = Column(Float, default=0.0)
    result = Column(Enum(AcceptanceResult), nullable=True)

    reinspection_count = Column(Integer, default=0)
    is_final = Column(Boolean, default=False)

    notes = Column(Text)
    acceptance_date = Column(Date, default=date.today)
    created_at = Column(DateTime, default=datetime.utcnow)

    project = relationship("Project", back_populates="acceptances")


class Todo(Base):
    __tablename__ = "todos"

    id = Column(Integer, primary_key=True, index=True)
    project_id = Column(Integer, ForeignKey("projects.id"), nullable=False)

    todo_type = Column(Enum(TodoType), nullable=False)
    title = Column(String(255), nullable=False)
    description = Column(Text)

    is_resolved = Column(Boolean, default=False)
    resolved_at = Column(DateTime, nullable=True)

    related_measure_type = Column(Enum(MeasureType), nullable=True)
    lag_amount = Column(Float, nullable=True)
    overrun_amount = Column(Float, nullable=True)

    created_at = Column(DateTime, default=datetime.utcnow)

    project = relationship("Project", back_populates="todos")
