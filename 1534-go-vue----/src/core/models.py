from datetime import datetime, date
from typing import Optional, List
from pydantic import BaseModel, Field
from sqlalchemy import Column, Integer, String, DateTime, Float, ForeignKey, Date, Text, Boolean
from sqlalchemy.orm import relationship, declarative_base
from sqlalchemy.sql import func

Base = declarative_base()


class FacilityType(str):
    pass


class EnterpriseORM(Base):
    __tablename__ = "enterprises"
    id = Column(Integer, primary_key=True, autoincrement=True)
    name = Column(String(200), nullable=False, index=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    facilities = relationship("FacilityORM", back_populates="enterprise", cascade="all, delete-orphan")


class FacilityORM(Base):
    __tablename__ = "facilities"
    id = Column(Integer, primary_key=True, autoincrement=True)
    enterprise_id = Column(Integer, ForeignKey("enterprises.id"), nullable=False, index=True)
    name = Column(String(200), nullable=False)
    facility_type = Column(String(100), nullable=False)
    installed_at = Column(Date, nullable=False)
    is_running = Column(Boolean, default=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    enterprise = relationship("EnterpriseORM", back_populates="facilities")
    maintenances = relationship("MaintenanceORM", back_populates="facility", cascade="all, delete-orphan")
    emissions = relationship("EmissionORM", back_populates="facility", cascade="all, delete-orphan")
    todos = relationship("TodoORM", back_populates="facility", cascade="all, delete-orphan")


class MaintenanceORM(Base):
    __tablename__ = "maintenances"
    id = Column(Integer, primary_key=True, autoincrement=True)
    facility_id = Column(Integer, ForeignKey("facilities.id"), nullable=False, index=True)
    maintenance_date = Column(Date, nullable=False, index=True)
    description = Column(Text, nullable=True)
    total_cost = Column(Float, default=0.0)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    facility = relationship("FacilityORM", back_populates="maintenances")
    parts = relationship("MaintenancePartORM", back_populates="maintenance", cascade="all, delete-orphan")


class MaintenancePartORM(Base):
    __tablename__ = "maintenance_parts"
    id = Column(Integer, primary_key=True, autoincrement=True)
    maintenance_id = Column(Integer, ForeignKey("maintenances.id"), nullable=False, index=True)
    part_name = Column(String(200), nullable=False)
    quantity = Column(Integer, default=1)
    unit_cost = Column(Float, default=0.0)
    maintenance = relationship("MaintenanceORM", back_populates="parts")


class EmissionORM(Base):
    __tablename__ = "emissions"
    id = Column(Integer, primary_key=True, autoincrement=True)
    facility_id = Column(Integer, ForeignKey("facilities.id"), nullable=False, index=True)
    recorded_at = Column(DateTime(timezone=True), nullable=False, index=True)
    pollutant = Column(String(100), nullable=False)
    value = Column(Float, nullable=False)
    limit_value = Column(Float, nullable=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    facility = relationship("FacilityORM", back_populates="emissions")


class TodoORM(Base):
    __tablename__ = "todos"
    id = Column(Integer, primary_key=True, autoincrement=True)
    facility_id = Column(Integer, ForeignKey("facilities.id"), nullable=False, index=True)
    todo_type = Column(String(50), nullable=False)
    title = Column(String(200), nullable=False)
    description = Column(Text, nullable=True)
    due_date = Column(Date, nullable=False, index=True)
    status = Column(String(50), default="pending")
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    facility = relationship("FacilityORM", back_populates="todos")


class AgingAlertORM(Base):
    __tablename__ = "aging_alerts"
    id = Column(Integer, primary_key=True, autoincrement=True)
    facility_id = Column(Integer, ForeignKey("facilities.id"), nullable=False, index=True)
    part_name = Column(String(200), nullable=False)
    alert_date = Column(Date, nullable=False, index=True)
    resolved = Column(Boolean, default=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now())


class EnterpriseCreate(BaseModel):
    name: str = Field(..., min_length=1, max_length=200)


class EnterpriseResponse(BaseModel):
    id: int
    name: str
    created_at: datetime

    class Config:
        from_attributes = True


class FacilityCreate(BaseModel):
    enterprise_id: int
    name: str = Field(..., min_length=1, max_length=200)
    facility_type: str
    installed_at: date
    is_running: bool = True


class FacilityResponse(BaseModel):
    id: int
    enterprise_id: int
    name: str
    facility_type: str
    installed_at: date
    is_running: bool
    created_at: datetime

    class Config:
        from_attributes = True


class MaintenancePartCreate(BaseModel):
    part_name: str
    quantity: int = 1
    unit_cost: float = 0.0


class MaintenanceCreate(BaseModel):
    facility_id: int
    maintenance_date: date
    description: Optional[str] = None
    parts: List[MaintenancePartCreate] = []


class MaintenancePartResponse(BaseModel):
    id: int
    part_name: str
    quantity: int
    unit_cost: float

    class Config:
        from_attributes = True


class MaintenanceResponse(BaseModel):
    id: int
    facility_id: int
    maintenance_date: date
    description: Optional[str]
    total_cost: float
    parts: List[MaintenancePartResponse]
    created_at: datetime

    class Config:
        from_attributes = True


class EmissionCreate(BaseModel):
    facility_id: int
    recorded_at: datetime
    pollutant: str
    value: float
    limit_value: float


class EmissionResponse(BaseModel):
    id: int
    facility_id: int
    recorded_at: datetime
    pollutant: str
    value: float
    limit_value: float
    is_compliant: bool

    class Config:
        from_attributes = True


class TodoResponse(BaseModel):
    id: int
    facility_id: int
    todo_type: str
    title: str
    description: Optional[str]
    due_date: date
    status: str
    created_at: datetime

    class Config:
        from_attributes = True


class AgingAlertResponse(BaseModel):
    id: int
    facility_id: int
    part_name: str
    alert_date: date
    resolved: bool

    class Config:
        from_attributes = True


class MonthlyStatResponse(BaseModel):
    year: int
    month: int
    facility_id: int
    facility_name: str
    operation_rate: float
    compliance_rate: float
    maintenance_completion_rate: float
    is_focus: bool
