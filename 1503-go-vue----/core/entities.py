from datetime import datetime, date
from sqlalchemy import Column, Integer, String, Float, Date, DateTime, ForeignKey, Boolean, Enum as SAEnum
from sqlalchemy.orm import relationship
from .database import Base
from .models import Role, AnomalyType, Severity, FireRiskType, FireRiskStatus, TodoStatus, ReportStatus


class UserEntity(Base):
    __tablename__ = "users"
    
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String, nullable=False)
    role = Column(SAEnum(Role), nullable=False)
    phone = Column(String, nullable=True)
    
    areas_as_ranger = relationship("ForestAreaEntity", back_populates="ranger")
    assigned_tasks = relationship("PatrolTaskEntity", back_populates="ranger")
    assigned_todos = relationship("TodoEntity", back_populates="assigned_user")


class ForestAreaEntity(Base):
    __tablename__ = "forest_areas"
    
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String, nullable=False, unique=True)
    area_km2 = Column(Float, nullable=False)
    main_tree_species = Column(String, nullable=False)
    ranger_id = Column(Integer, ForeignKey("users.id"), nullable=True)
    
    ranger = relationship("UserEntity", back_populates="areas_as_ranger")
    tasks = relationship("PatrolTaskEntity", back_populates="area")


class PatrolTaskEntity(Base):
    __tablename__ = "patrol_tasks"
    
    id = Column(Integer, primary_key=True, index=True)
    area_id = Column(Integer, ForeignKey("forest_areas.id"), nullable=False)
    ranger_id = Column(Integer, ForeignKey("users.id"), nullable=False)
    patrol_date = Column(Date, nullable=False)
    route_keypoints = Column(String, nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    
    area = relationship("ForestAreaEntity", back_populates="tasks")
    ranger = relationship("UserEntity", back_populates="assigned_tasks")
    report = relationship("PatrolReportEntity", back_populates="task", uselist=False)


class PatrolReportEntity(Base):
    __tablename__ = "patrol_reports"
    
    id = Column(Integer, primary_key=True, index=True)
    task_id = Column(Integer, ForeignKey("patrol_tasks.id"), nullable=False, unique=True)
    submitted_at = Column(DateTime, default=datetime.utcnow)
    actual_route = Column(String, nullable=False)
    status = Column(SAEnum(ReportStatus), default=ReportStatus.PENDING)
    
    task = relationship("PatrolTaskEntity", back_populates="report")
    anomalies = relationship("AnomalyEntity", back_populates="report")


class AnomalyEntity(Base):
    __tablename__ = "anomalies"
    
    id = Column(Integer, primary_key=True, index=True)
    report_id = Column(Integer, ForeignKey("patrol_reports.id"), nullable=False)
    anomaly_type = Column(SAEnum(AnomalyType), nullable=False)
    location = Column(String, nullable=False)
    is_duplicate = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    
    pest = relationship("PestAnomalyEntity", back_populates="anomaly", uselist=False)
    fire_risk = relationship("FireRiskAnomalyEntity", back_populates="anomaly", uselist=False)
    report = relationship("PatrolReportEntity", back_populates="anomalies")
    todo = relationship("TodoEntity", back_populates="anomaly", uselist=False)


class PestAnomalyEntity(Base):
    __tablename__ = "pest_anomalies"
    
    id = Column(Integer, primary_key=True, index=True)
    anomaly_id = Column(Integer, ForeignKey("anomalies.id"), nullable=False, unique=True)
    type = Column(String, nullable=False)
    severity = Column(SAEnum(Severity), nullable=False)
    trees_affected = Column(Integer, nullable=False, default=0)
    location = Column(String, nullable=False)
    
    anomaly = relationship("AnomalyEntity", back_populates="pest")


class FireRiskAnomalyEntity(Base):
    __tablename__ = "fire_risk_anomalies"
    
    id = Column(Integer, primary_key=True, index=True)
    anomaly_id = Column(Integer, ForeignKey("anomalies.id"), nullable=False, unique=True)
    type = Column(SAEnum(FireRiskType), nullable=False)
    status = Column(SAEnum(FireRiskStatus), default=FireRiskStatus.PENDING)
    location = Column(String, nullable=False)
    
    anomaly = relationship("AnomalyEntity", back_populates="fire_risk")


class TodoEntity(Base):
    __tablename__ = "todos"
    
    id = Column(Integer, primary_key=True, index=True)
    anomaly_id = Column(Integer, ForeignKey("anomalies.id"), nullable=False, unique=True)
    anomaly_type = Column(SAEnum(AnomalyType), nullable=False)
    area_id = Column(Integer, ForeignKey("forest_areas.id"), nullable=False)
    assigned_user_id = Column(Integer, ForeignKey("users.id"), nullable=False)
    status = Column(SAEnum(TodoStatus), default=TodoStatus.PENDING)
    created_at = Column(DateTime, default=datetime.utcnow)
    resolved_at = Column(DateTime, nullable=True)
    
    anomaly = relationship("AnomalyEntity", back_populates="todo")
    assigned_user = relationship("UserEntity", back_populates="assigned_todos")
