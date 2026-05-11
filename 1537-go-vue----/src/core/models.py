from decimal import Decimal
from datetime import datetime, date
from enum import Enum as PyEnum
from sqlalchemy import (
    Column, Integer, String, DateTime, Date, Boolean, 
    Numeric, Text, ForeignKey, Enum, UniqueConstraint
)
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func

from .database import Base

class DataSourceType(PyEnum):
    MANUAL = "manual"
    ONLINE = "online"

class DataStatus(PyEnum):
    PENDING_REVIEW = "pending_review"
    APPROVED = "approved"
    REJECTED = "rejected"
    AUTO_FILLED = "auto_filled"

class TransactionType(PyEnum):
    BUY = "buy"
    SELL = "sell"

class ReportStatus(PyEnum):
    DRAFT = "draft"
    CONFIRMED = "confirmed"

class Industry(Base):
    __tablename__ = "industries"
    
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(100), unique=True, nullable=False, index=True)
    code = Column(String(20), unique=True, nullable=False, index=True)
    default_oxidation_rate = Column(Numeric(precision=10, scale=4), default=Decimal("1.0000"))
    min_oxidation_rate = Column(Numeric(precision=10, scale=4), default=Decimal("0.9000"))
    description = Column(Text, nullable=True)
    
    companies = relationship("Company", back_populates="industry")

class Company(Base):
    __tablename__ = "companies"
    
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(200), nullable=False, index=True)
    registration_no = Column(String(50), unique=True, nullable=False, index=True)
    industry_id = Column(Integer, ForeignKey("industries.id"), nullable=False)
    annual_output = Column(Numeric(precision=20, scale=4), nullable=False)
    address = Column(String(500), nullable=True)
    contact_person = Column(String(50), nullable=True)
    contact_phone = Column(String(20), nullable=True)
    initial_quota = Column(Numeric(precision=20, scale=4), default=Decimal("0.0000"))
    created_at = Column(DateTime, default=func.now(), nullable=False)
    updated_at = Column(DateTime, default=func.now(), onupdate=func.now(), nullable=False)
    
    industry = relationship("Industry", back_populates="companies")
    emission_sources = relationship("EmissionSource", back_populates="company", cascade="all, delete-orphan")
    quota_transactions = relationship("QuotaTransaction", back_populates="company", cascade="all, delete-orphan")
    reports = relationship("Report", back_populates="company", cascade="all, delete-orphan")

class EmissionSource(Base):
    __tablename__ = "emission_sources"
    
    id = Column(Integer, primary_key=True, index=True)
    company_id = Column(Integer, ForeignKey("companies.id"), nullable=False)
    name = Column(String(200), nullable=False, index=True)
    code = Column(String(50), nullable=False)
    emission_type = Column(String(100), nullable=False)
    emission_factor = Column(Numeric(precision=10, scale=4), nullable=False)
    oxidation_rate = Column(Numeric(precision=10, scale=4), nullable=False)
    unit = Column(String(20), nullable=False)
    data_source = Column(Enum(DataSourceType), default=DataSourceType.MANUAL, nullable=False)
    description = Column(Text, nullable=True)
    created_at = Column(DateTime, default=func.now(), nullable=False)
    updated_at = Column(DateTime, default=func.now(), onupdate=func.now(), nullable=False)
    
    company = relationship("Company", back_populates="emission_sources")
    emission_data = relationship("EmissionData", back_populates="source", cascade="all, delete-orphan")
    
    __table_args__ = (UniqueConstraint('company_id', 'code', name='uq_company_source_code'),)

class EmissionData(Base):
    __tablename__ = "emission_data"
    
    id = Column(Integer, primary_key=True, index=True)
    source_id = Column(Integer, ForeignKey("emission_sources.id"), nullable=False)
    record_date = Column(Date, nullable=False, index=True)
    record_hour = Column(Integer, nullable=False)
    activity_data = Column(Numeric(precision=20, scale=4), nullable=False)
    emission_amount = Column(Numeric(precision=20, scale=4), nullable=False)
    data_source = Column(Enum(DataSourceType), nullable=False)
    status = Column(Enum(DataStatus), default=DataStatus.PENDING_REVIEW, nullable=False)
    is_device_fault = Column(Boolean, default=False, nullable=False)
    device_fault_reason = Column(String(200), nullable=True)
    created_at = Column(DateTime, default=func.now(), nullable=False)
    updated_at = Column(DateTime, default=func.now(), onupdate=func.now(), nullable=False)
    
    source = relationship("EmissionSource", back_populates="emission_data")
    
    __table_args__ = (UniqueConstraint('source_id', 'record_date', 'record_hour', name='uq_source_hourly'),)

class QuotaTransaction(Base):
    __tablename__ = "quota_transactions"
    
    id = Column(Integer, primary_key=True, index=True)
    company_id = Column(Integer, ForeignKey("companies.id"), nullable=False)
    transaction_type = Column(Enum(TransactionType), nullable=False)
    quota_amount = Column(Numeric(precision=20, scale=4), nullable=False)
    price_per_unit = Column(Numeric(precision=20, scale=4), nullable=False)
    total_amount = Column(Numeric(precision=20, scale=4), nullable=False)
    counterparty = Column(String(200), nullable=True)
    transaction_date = Column(Date, nullable=False, index=True)
    remarks = Column(Text, nullable=True)
    created_at = Column(DateTime, default=func.now(), nullable=False)
    
    company = relationship("Company", back_populates="quota_transactions")

class Report(Base):
    __tablename__ = "reports"
    
    id = Column(Integer, primary_key=True, index=True)
    company_id = Column(Integer, ForeignKey("companies.id"), nullable=False)
    year = Column(Integer, nullable=False, index=True)
    month = Column(Integer, nullable=False, index=True)
    total_emission = Column(Numeric(precision=20, scale=4), default=Decimal("0.0000"))
    status = Column(Enum(ReportStatus), default=ReportStatus.DRAFT, nullable=False)
    generated_at = Column(DateTime, default=func.now(), nullable=False)
    confirmed_at = Column(DateTime, nullable=True)
    generated_by = Column(String(50), nullable=True)
    remarks = Column(Text, nullable=True)
    
    company = relationship("Company", back_populates="reports")
    
    __table_args__ = (UniqueConstraint('company_id', 'year', 'month', name='uq_company_monthly_report'),)
