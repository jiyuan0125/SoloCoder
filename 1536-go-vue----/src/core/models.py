from datetime import datetime
from enum import Enum as PyEnum
from sqlalchemy import (
    Column,
    Integer,
    String,
    Text,
    DateTime,
    Float,
    Boolean,
    ForeignKey,
    Enum,
)
from sqlalchemy.orm import relationship, declarative_base

Base = declarative_base()


class StageType(PyEnum):
    PLANNING = "planning"
    PRE_AUDIT = "pre_audit"
    AUDIT = "audit"
    SOLUTION_GENERATION = "solution_generation"
    IMPLEMENTATION = "implementation"
    ACCEPTANCE = "acceptance"


class SolutionType(PyEnum):
    NO_LOW_COST = "no_low_cost"
    MEDIUM_HIGH_COST = "medium_high_cost"


class SolutionStatus(PyEnum):
    DRAFT = "draft"
    PENDING_FILTER = "pending_filter"
    PASSED_FILTER = "passed_filter"
    REJECTED = "rejected"
    IMPLEMENTING = "implementing"
    COMPLETED = "completed"
    NOT_COMPLIANT = "not_compliant"


class AuditStatus(PyEnum):
    IN_PROGRESS = "in_progress"
    OVERDUE = "overdue"
    PASSED = "passed"
    FAILED = "failed"


class Audit(Base):
    __tablename__ = "audits"

    id = Column(Integer, primary_key=True, autoincrement=True)
    company_name = Column(String(255), nullable=False)
    company_id = Column(String(100), nullable=True)
    start_date = Column(DateTime, nullable=False, default=datetime.utcnow)
    acceptance_date = Column(DateTime, nullable=True)
    status = Column(Enum(AuditStatus), default=AuditStatus.IN_PROGRESS, nullable=False)
    acceptance_score = Column(Float, nullable=True)
    is_overdue = Column(Boolean, default=False, nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow, nullable=False)

    stages = relationship("Stage", back_populates="audit", cascade="all, delete-orphan")
    solutions = relationship("Solution", back_populates="audit", cascade="all, delete-orphan")


class Stage(Base):
    __tablename__ = "stages"

    id = Column(Integer, primary_key=True, autoincrement=True)
    audit_id = Column(Integer, ForeignKey("audits.id"), nullable=False)
    stage_type = Column(Enum(StageType), nullable=False)
    start_date = Column(DateTime, default=datetime.utcnow, nullable=False)
    end_date = Column(DateTime, nullable=True)
    is_completed = Column(Boolean, default=False, nullable=False)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow, nullable=False)

    audit = relationship("Audit", back_populates="stages")


class Solution(Base):
    __tablename__ = "solutions"

    id = Column(Integer, primary_key=True, autoincrement=True)
    audit_id = Column(Integer, ForeignKey("audits.id"), nullable=False)
    name = Column(String(255), nullable=False)
    description = Column(Text, nullable=True)
    solution_type = Column(Enum(SolutionType), nullable=False)
    status = Column(Enum(SolutionStatus), default=SolutionStatus.DRAFT, nullable=False)
    expected_energy_saving = Column(Float, nullable=True)
    expected_investment = Column(Float, nullable=True)
    actual_energy_saving = Column(Float, nullable=True)
    actual_investment = Column(Float, nullable=True)
    implementation_date = Column(DateTime, nullable=True)
    is_effective = Column(Boolean, nullable=True)
    filter_notes = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow, nullable=False)

    audit = relationship("Audit", back_populates="solutions")
    dimensions = relationship("Dimension", back_populates="solution", cascade="all, delete-orphan")


class Dimension(Base):
    __tablename__ = "dimensions"

    id = Column(Integer, primary_key=True, autoincrement=True)
    solution_id = Column(Integer, ForeignKey("solutions.id"), nullable=False)
    dimension_name = Column(String(100), nullable=False)
    score = Column(Integer, nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)

    solution = relationship("Solution", back_populates="dimensions")


class AuditLog(Base):
    __tablename__ = "audit_logs"

    id = Column(Integer, primary_key=True, autoincrement=True)
    audit_id = Column(Integer, ForeignKey("audits.id"), nullable=True)
    user_id = Column(String(100), nullable=True)
    user_name = Column(String(100), nullable=True)
    action = Column(String(255), nullable=False)
    table_name = Column(String(100), nullable=True)
    record_id = Column(Integer, nullable=True)
    old_value = Column(Text, nullable=True)
    new_value = Column(Text, nullable=True)
    timestamp = Column(DateTime, default=datetime.utcnow, nullable=False)
