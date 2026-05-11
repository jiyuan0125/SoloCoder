from datetime import date, datetime
from sqlalchemy import (
    Column,
    Integer,
    String,
    Date,
    Float,
    DateTime,
    ForeignKey,
    Boolean,
    Text,
)
from sqlalchemy.orm import relationship
from server.database import Base


class Crew(Base):
    __tablename__ = "crews"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), nullable=False)
    identity_number = Column(String(50), unique=True, index=True, nullable=False)
    phone = Column(String(20))
    email = Column(String(100))
    position = Column(String(50))
    is_intern = Column(Boolean, default=False)
    daily_rate = Column(Float, nullable=False)
    status = Column(String(20), default="待派")
    
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    certificates = relationship("Certificate", back_populates="crew", cascade="all, delete-orphan")
    trainings = relationship("TrainingRecord", back_populates="crew", cascade="all, delete-orphan")
    assignments = relationship("Assignment", back_populates="crew", cascade="all, delete-orphan")
    attendances = relationship("Attendance", back_populates="crew", cascade="all, delete-orphan")
    salaries = relationship("Salary", back_populates="crew", cascade="all, delete-orphan")


class Certificate(Base):
    __tablename__ = "certificates"

    id = Column(Integer, primary_key=True, index=True)
    crew_id = Column(Integer, ForeignKey("crews.id"), nullable=False)
    certificate_type = Column(String(100), nullable=False)
    certificate_number = Column(String(100))
    issue_date = Column(Date, nullable=False)
    expiry_date = Column(Date, nullable=False)
    validity_days = Column(Integer)
    status = Column(String(20), default="有效")
    
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    crew = relationship("Crew", back_populates="certificates")


class TrainingRecord(Base):
    __tablename__ = "training_records"

    id = Column(Integer, primary_key=True, index=True)
    crew_id = Column(Integer, ForeignKey("crews.id"), nullable=False)
    training_type = Column(String(100), nullable=False)
    training_name = Column(String(200), nullable=False)
    training_date = Column(Date, nullable=False)
    next_renewal_date = Column(Date)
    certificate_id = Column(Integer, ForeignKey("certificates.id"))
    
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    crew = relationship("Crew", back_populates="trainings")
    certificate = relationship("Certificate")


class Ship(Base):
    __tablename__ = "ships"

    id = Column(Integer, primary_key=True, index=True)
    ship_name = Column(String(100), nullable=False, unique=True)
    imo_number = Column(String(50), unique=True)
    vessel_type = Column(String(50))
    total_crew_quota = Column(Integer, nullable=False)
    current_crew_count = Column(Integer, default=0)
    status = Column(String(20), default="在航")
    
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    assignments = relationship("Assignment", back_populates="ship", cascade="all, delete-orphan")


class Assignment(Base):
    __tablename__ = "assignments"

    id = Column(Integer, primary_key=True, index=True)
    crew_id = Column(Integer, ForeignKey("crews.id"), nullable=False)
    ship_id = Column(Integer, ForeignKey("ships.id"), nullable=False)
    position = Column(String(50), nullable=False)
    is_watch_keeper = Column(Boolean, default=False)
    embark_date = Column(Date, nullable=False)
    disembark_date = Column(Date)
    status = Column(String(20), default="在船")
    
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    crew = relationship("Crew", back_populates="assignments")
    ship = relationship("Ship", back_populates="assignments")


class Attendance(Base):
    __tablename__ = "attendances"

    id = Column(Integer, primary_key=True, index=True)
    crew_id = Column(Integer, ForeignKey("crews.id"), nullable=False)
    assignment_id = Column(Integer, ForeignKey("assignments.id"))
    ship_id = Column(Integer, ForeignKey("ships.id"))
    attendance_date = Column(Date, nullable=False)
    status = Column(String(20), nullable=False)
    is_half_day = Column(Boolean, default=False)
    remarks = Column(Text)
    
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    crew = relationship("Crew", back_populates="attendances")
    assignment = relationship("Assignment")
    ship = relationship("Ship")


class Salary(Base):
    __tablename__ = "salaries"

    id = Column(Integer, primary_key=True, index=True)
    crew_id = Column(Integer, ForeignKey("crews.id"), nullable=False)
    assignment_id = Column(Integer, ForeignKey("assignments.id"))
    ship_id = Column(Integer, ForeignKey("ships.id"))
    salary_month = Column(String(7), nullable=False)
    base_amount = Column(Float, nullable=False, default=0.0)
    normal_days = Column(Float, default=0.0)
    overtime_days = Column(Float, default=0.0)
    sick_leave_days = Column(Float, default=0.0)
    personal_leave_days = Column(Float, default=0.0)
    normal_amount = Column(Float, default=0.0)
    overtime_amount = Column(Float, default=0.0)
    sick_leave_amount = Column(Float, default=0.0)
    total_amount = Column(Float, nullable=False, default=0.0)
    status = Column(String(20), default="待结算")
    settlement_date = Column(Date)
    remarks = Column(Text)
    
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    crew = relationship("Crew", back_populates="salaries")
    assignment = relationship("Assignment")
    ship = relationship("Ship")
