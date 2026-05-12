from datetime import datetime
from enum import Enum
from sqlalchemy import Column, Integer, String, Float, DateTime, ForeignKey, Boolean, Text
from sqlalchemy.orm import relationship
from database import Base


class InspectionType(str, Enum):
    DAILY = "日常"
    COMPREHENSIVE = "综合"
    SPECIAL = "专项"


class DefectLevel(str, Enum):
    LEVEL_1 = "一级"
    LEVEL_2 = "二级"
    LEVEL_3 = "三级"


class TaskStatus(str, Enum):
    PENDING = "待处理"
    IN_PROGRESS = "处理中"
    COMPLETED = "已完成"
    CANCELLED = "已取消"


class TeamStatus(str, Enum):
    IDLE = "空闲"
    BUSY = "工作中"


class LineSection(Base):
    __tablename__ = "line_sections"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), unique=True, nullable=False)
    start_km = Column(Float, nullable=False)
    end_km = Column(Float, nullable=False)
    description = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    inspections = relationship("Inspection", back_populates="section")
    rails = relationship("Rail", back_populates="section")


class MaintenanceTeam(Base):
    __tablename__ = "maintenance_teams"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(50), unique=True, nullable=False)
    leader = Column(String(50), nullable=False)
    phone = Column(String(20))
    status = Column(String(20), default=TeamStatus.IDLE.value)
    current_task_id = Column(Integer, ForeignKey("tasks.id"), nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    current_task = relationship("Task", foreign_keys=[current_task_id], post_update=True)
    tasks = relationship("Task", back_populates="team", foreign_keys="Task.team_id")
    inspections = relationship("Inspection", back_populates="team")


class Inspection(Base):
    __tablename__ = "inspections"

    id = Column(Integer, primary_key=True, index=True)
    section_id = Column(Integer, ForeignKey("line_sections.id"), nullable=False)
    team_id = Column(Integer, ForeignKey("maintenance_teams.id"), nullable=False)
    inspection_type = Column(String(20), nullable=False)
    inspection_date = Column(DateTime, default=datetime.utcnow)
    inspector = Column(String(50))
    notes = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)

    section = relationship("LineSection", back_populates="inspections")
    team = relationship("MaintenanceTeam", back_populates="inspections")
    defects = relationship("Defect", back_populates="inspection")


class Defect(Base):
    __tablename__ = "defects"

    id = Column(Integer, primary_key=True, index=True)
    inspection_id = Column(Integer, ForeignKey("inspections.id"), nullable=False)
    section_id = Column(Integer, ForeignKey("line_sections.id"), nullable=False)
    defect_type = Column(String(100), nullable=False)
    location_km = Column(Float, nullable=False)
    level = Column(String(20), nullable=False)
    description = Column(Text)
    is_resolved = Column(Boolean, default=False)
    resolved_at = Column(DateTime, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    inspection = relationship("Inspection", back_populates="defects")
    tasks = relationship("Task", back_populates="defect")


class Task(Base):
    __tablename__ = "tasks"

    id = Column(Integer, primary_key=True, index=True)
    defect_id = Column(Integer, ForeignKey("defects.id"), nullable=False)
    team_id = Column(Integer, ForeignKey("maintenance_teams.id"), nullable=True)
    task_type = Column(String(50), nullable=False)
    status = Column(String(20), default=TaskStatus.PENDING.value)
    priority = Column(Integer, default=0)
    due_date = Column(DateTime, nullable=False)
    started_at = Column(DateTime, nullable=True)
    completed_at = Column(DateTime, nullable=True)
    notes = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    defect = relationship("Defect", back_populates="tasks")
    team = relationship("MaintenanceTeam", back_populates="tasks", foreign_keys=[team_id])


class Rail(Base):
    __tablename__ = "rails"

    id = Column(Integer, primary_key=True, index=True)
    section_id = Column(Integer, ForeignKey("line_sections.id"), nullable=False)
    rail_number = Column(String(50), nullable=False)
    start_km = Column(Float, nullable=False)
    end_km = Column(Float, nullable=False)
    max_total_weight = Column(Float, nullable=False, comment="最大通过总重（百万吨）")
    current_total_weight = Column(Float, default=0, comment="当前通过总重（百万吨）")
    wear_mm = Column(Float, default=0, comment="磨耗量（毫米）")
    installation_date = Column(DateTime, default=datetime.utcnow)
    is_replaced = Column(Boolean, default=False)
    replaced_at = Column(DateTime, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    section = relationship("LineSection", back_populates="rails")
    maintenance_records = relationship("RailMaintenance", back_populates="rail")
    replacement_records = relationship("RailReplacement", back_populates="rail")


class RailMaintenance(Base):
    __tablename__ = "rail_maintenance"

    id = Column(Integer, primary_key=True, index=True)
    rail_id = Column(Integer, ForeignKey("rails.id"), nullable=False)
    maintenance_type = Column(String(50), nullable=False)
    maintenance_date = Column(DateTime, default=datetime.utcnow)
    before_wear_mm = Column(Float)
    after_wear_mm = Column(Float)
    notes = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)

    rail = relationship("Rail", back_populates="maintenance_records")


class RailReplacement(Base):
    __tablename__ = "rail_replacements"

    id = Column(Integer, primary_key=True, index=True)
    rail_id = Column(Integer, ForeignKey("rails.id"), nullable=False)
    old_rail_number = Column(String(50), nullable=False)
    new_rail_number = Column(String(50), nullable=False)
    replacement_reason = Column(String(100), nullable=False)
    replacement_date = Column(DateTime, default=datetime.utcnow)
    notes = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)

    rail = relationship("Rail", back_populates="replacement_records")


class Warning(Base):
    __tablename__ = "warnings"

    id = Column(Integer, primary_key=True, index=True)
    warning_type = Column(String(50), nullable=False)
    target_type = Column(String(50), nullable=False)
    target_id = Column(Integer, nullable=False)
    message = Column(Text, nullable=False)
    is_acknowledged = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)
