from datetime import datetime, date
from enum import Enum as PyEnum
from sqlalchemy import String, Integer, Float, Date, DateTime, Boolean, ForeignKey, Text, Enum
from sqlalchemy.orm import Mapped, mapped_column, relationship
from sqlalchemy.sql import func

from .database import Base


class ControlLevel(PyEnum):
    NATIONAL = "national"
    PROVINCIAL = "provincial"
    MUNICIPAL = "municipal"


class UserRole(PyEnum):
    ADMIN = "admin"
    USER = "user"


class EvaluationStatus(PyEnum):
    PENDING = "pending"
    PUBLISHED = "published"


class IndicatorType(PyEnum):
    PH = "ph"
    DO = "do"
    COD = "cod"
    NH3N = "nh3n"
    TP = "tp"
    TN = "tn"


class AuditAction(PyEnum):
    CREATE = "create"
    UPDATE = "update"
    DELETE = "delete"


class User(Base):
    __tablename__ = "users"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    username: Mapped[str] = mapped_column(String(50), unique=True, nullable=False, index=True)
    hashed_password: Mapped[str] = mapped_column(String(255), nullable=False)
    full_name: Mapped[str | None] = mapped_column(String(100))
    role: Mapped[UserRole] = mapped_column(Enum(UserRole), default=UserRole.USER)
    is_active: Mapped[bool] = mapped_column(Boolean, default=True)
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), server_default=func.now())

    audit_logs: Mapped[list["AuditLog"]] = relationship("AuditLog", back_populates="operator")


class Section(Base):
    __tablename__ = "sections"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    name: Mapped[str] = mapped_column(String(100), nullable=False, index=True)
    code: Mapped[str] = mapped_column(String(50), unique=True, nullable=False, index=True)
    river: Mapped[str | None] = mapped_column(String(100))
    control_level: Mapped[ControlLevel] = mapped_column(Enum(ControlLevel), nullable=False)
    description: Mapped[str | None] = mapped_column(Text)
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), server_default=func.now())

    current_standard: Mapped["StandardLimit | None"] = relationship(
        "StandardLimit",
        back_populates="section",
        uselist=False,
        primaryjoin="and_(Section.id==StandardLimit.section_id, StandardLimit.is_current==True)",
        overlaps="standards",
    )
    standards: Mapped[list["StandardLimit"]] = relationship(
        "StandardLimit",
        back_populates="section",
        order_by="StandardLimit.created_at",
        overlaps="current_standard",
    )
    monitoring_data: Mapped[list["MonitoringData"]] = relationship("MonitoringData", back_populates="section", order_by="MonitoringData.monitoring_date")
    evaluations: Mapped[list["MonthlyEvaluation"]] = relationship("MonthlyEvaluation", back_populates="section", order_by="MonthlyEvaluation.year, MonthlyEvaluation.month")
    tasks: Mapped[list["MonitoringTask"]] = relationship("MonitoringTask", back_populates="section", order_by="MonitoringTask.scheduled_date")


class StandardLimit(Base):
    __tablename__ = "standard_limits"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    section_id: Mapped[int] = mapped_column(Integer, ForeignKey("sections.id"), nullable=False, index=True)
    ph_min: Mapped[float | None] = mapped_column(Float)
    ph_max: Mapped[float | None] = mapped_column(Float)
    do_limit: Mapped[float | None] = mapped_column(Float)
    cod_limit: Mapped[float | None] = mapped_column(Float)
    nh3n_limit: Mapped[float | None] = mapped_column(Float)
    tp_limit: Mapped[float | None] = mapped_column(Float)
    tn_limit: Mapped[float | None] = mapped_column(Float)
    ph_class: Mapped[str | None] = mapped_column(String(20))
    do_class: Mapped[str | None] = mapped_column(String(20))
    cod_class: Mapped[str | None] = mapped_column(String(20))
    nh3n_class: Mapped[str | None] = mapped_column(String(20))
    tp_class: Mapped[str | None] = mapped_column(String(20))
    tn_class: Mapped[str | None] = mapped_column(String(20))
    overall_class: Mapped[str | None] = mapped_column(String(20))
    is_current: Mapped[bool] = mapped_column(Boolean, default=True)
    effective_date: Mapped[date] = mapped_column(Date, default=date.today)
    description: Mapped[str | None] = mapped_column(Text)
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), server_default=func.now())

    section: Mapped[Section] = relationship("Section", back_populates="standards")


class MonitoringTask(Base):
    __tablename__ = "monitoring_tasks"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    section_id: Mapped[int] = mapped_column(Integer, ForeignKey("sections.id"), nullable=False, index=True)
    scheduled_date: Mapped[date] = mapped_column(Date, nullable=False)
    is_flood_season: Mapped[bool] = mapped_column(Boolean, default=False)
    status: Mapped[str] = mapped_column(String(20), default="pending")
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), server_default=func.now())

    section: Mapped[Section] = relationship("Section", back_populates="tasks")
    monitoring_data: Mapped[list["MonitoringData"]] = relationship("MonitoringData", back_populates="task")


class MonitoringData(Base):
    __tablename__ = "monitoring_data"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    section_id: Mapped[int] = mapped_column(Integer, ForeignKey("sections.id"), nullable=False, index=True)
    task_id: Mapped[int | None] = mapped_column(Integer, ForeignKey("monitoring_tasks.id"))
    monitoring_date: Mapped[date] = mapped_column(Date, nullable=False, index=True)
    ph_value: Mapped[float | None] = mapped_column(Float)
    do_value: Mapped[float | None] = mapped_column(Float)
    cod_value: Mapped[float | None] = mapped_column(Float)
    nh3n_value: Mapped[float | None] = mapped_column(Float)
    tp_value: Mapped[float | None] = mapped_column(Float)
    tn_value: Mapped[float | None] = mapped_column(Float)
    sampler_name: Mapped[str | None] = mapped_column(String(100))
    sampler_id: Mapped[str | None] = mapped_column(String(50))
    remarks: Mapped[str | None] = mapped_column(Text)
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), server_default=func.now())

    section: Mapped[Section] = relationship("Section", back_populates="monitoring_data")
    task: Mapped[MonitoringTask | None] = relationship("MonitoringTask", back_populates="monitoring_data")


class MonthlyEvaluation(Base):
    __tablename__ = "monthly_evaluations"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    section_id: Mapped[int] = mapped_column(Integer, ForeignKey("sections.id"), nullable=False, index=True)
    year: Mapped[int] = mapped_column(Integer, nullable=False)
    month: Mapped[int] = mapped_column(Integer, nullable=False)
    total_count: Mapped[int] = mapped_column(Integer, default=0)
    pass_count: Mapped[int] = mapped_column(Integer, default=0)
    pass_rate: Mapped[float | None] = mapped_column(Float)
    is_meet_standard: Mapped[bool | None] = mapped_column(Boolean)
    ph_pass_count: Mapped[int] = mapped_column(Integer, default=0)
    do_pass_count: Mapped[int] = mapped_column(Integer, default=0)
    cod_pass_count: Mapped[int] = mapped_column(Integer, default=0)
    nh3n_pass_count: Mapped[int] = mapped_column(Integer, default=0)
    tp_pass_count: Mapped[int] = mapped_column(Integer, default=0)
    tn_pass_count: Mapped[int] = mapped_column(Integer, default=0)
    worst_indicator: Mapped[str | None] = mapped_column(String(20))
    worst_value: Mapped[float | None] = mapped_column(Float)
    status: Mapped[EvaluationStatus] = mapped_column(Enum(EvaluationStatus), default=EvaluationStatus.PENDING)
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), server_default=func.now())
    updated_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now())

    section: Mapped[Section] = relationship("Section", back_populates="evaluations")


class AuditLog(Base):
    __tablename__ = "audit_logs"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    user_id: Mapped[int] = mapped_column(Integer, ForeignKey("users.id"), nullable=False, index=True)
    action: Mapped[AuditAction] = mapped_column(Enum(AuditAction), nullable=False)
    target_type: Mapped[str] = mapped_column(String(50), nullable=False)
    target_id: Mapped[int | None] = mapped_column(Integer)
    details: Mapped[str] = mapped_column(Text)
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), server_default=func.now())

    operator: Mapped[User] = relationship("User", back_populates="audit_logs")
