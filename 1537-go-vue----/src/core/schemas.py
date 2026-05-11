from decimal import Decimal
from datetime import datetime, date
from typing import Optional, List
from pydantic import BaseModel, Field, validator

from .models import (
    DataSourceType, DataStatus, TransactionType, ReportStatus
)

class IndustryBase(BaseModel):
    name: str = Field(..., max_length=100)
    code: str = Field(..., max_length=20)
    default_oxidation_rate: Decimal = Field(default=Decimal("1.0000"))
    min_oxidation_rate: Decimal = Field(default=Decimal("0.9000"))
    description: Optional[str] = None

class IndustryCreate(IndustryBase):
    pass

class IndustryUpdate(BaseModel):
    name: Optional[str] = Field(None, max_length=100)
    default_oxidation_rate: Optional[Decimal] = None
    min_oxidation_rate: Optional[Decimal] = None
    description: Optional[str] = None

class IndustryResponse(IndustryBase):
    id: int
    created_at: Optional[datetime] = None
    updated_at: Optional[datetime] = None
    
    class Config:
        from_attributes = True

class CompanyBase(BaseModel):
    name: str = Field(..., max_length=200)
    registration_no: str = Field(..., max_length=50)
    industry_id: int
    annual_output: Decimal = Field(..., gt=0)
    address: Optional[str] = Field(None, max_length=500)
    contact_person: Optional[str] = Field(None, max_length=50)
    contact_phone: Optional[str] = Field(None, max_length=20)
    initial_quota: Decimal = Field(default=Decimal("0.0000"))

class CompanyCreate(CompanyBase):
    pass

class CompanyUpdate(BaseModel):
    name: Optional[str] = Field(None, max_length=200)
    industry_id: Optional[int] = None
    annual_output: Optional[Decimal] = Field(None, gt=0)
    address: Optional[str] = Field(None, max_length=500)
    contact_person: Optional[str] = Field(None, max_length=50)
    contact_phone: Optional[str] = Field(None, max_length=20)
    initial_quota: Optional[Decimal] = None

class CompanyResponse(CompanyBase):
    id: int
    industry: Optional[IndustryResponse] = None
    created_at: datetime
    updated_at: datetime
    
    class Config:
        from_attributes = True

class EmissionSourceBase(BaseModel):
    name: str = Field(..., max_length=200)
    code: str = Field(..., max_length=50)
    emission_type: str = Field(..., max_length=100)
    emission_factor: Decimal = Field(..., gt=0)
    oxidation_rate: Decimal
    unit: str = Field(..., max_length=20)
    data_source: DataSourceType = DataSourceType.MANUAL
    description: Optional[str] = None

class EmissionSourceCreate(EmissionSourceBase):
    company_id: int

class EmissionSourceUpdate(BaseModel):
    name: Optional[str] = Field(None, max_length=200)
    emission_type: Optional[str] = Field(None, max_length=100)
    emission_factor: Optional[Decimal] = Field(None, gt=0)
    oxidation_rate: Optional[Decimal] = None
    unit: Optional[str] = Field(None, max_length=20)
    data_source: Optional[DataSourceType] = None
    description: Optional[str] = None

class EmissionSourceResponse(EmissionSourceBase):
    id: int
    company_id: int
    created_at: datetime
    updated_at: datetime
    
    class Config:
        from_attributes = True

class EmissionDataBase(BaseModel):
    activity_data: Decimal = Field(..., ge=0)
    data_source: DataSourceType

class EmissionDataCreate(EmissionDataBase):
    source_id: int
    record_date: date
    record_hour: int = Field(..., ge=0, le=23)

class EmissionDataUpdate(BaseModel):
    activity_data: Optional[Decimal] = Field(None, ge=0)
    status: Optional[DataStatus] = None

class EmissionDataResponse(EmissionDataBase):
    id: int
    source_id: int
    record_date: date
    record_hour: int
    emission_amount: Decimal
    status: DataStatus
    is_device_fault: bool
    device_fault_reason: Optional[str] = None
    created_at: datetime
    updated_at: datetime
    
    class Config:
        from_attributes = True

class QuotaTransactionBase(BaseModel):
    transaction_type: TransactionType
    quota_amount: Decimal = Field(..., gt=0)
    price_per_unit: Decimal = Field(..., gt=0)
    counterparty: Optional[str] = Field(None, max_length=200)
    transaction_date: date
    remarks: Optional[str] = None

class QuotaTransactionCreate(QuotaTransactionBase):
    company_id: int

class QuotaTransactionResponse(QuotaTransactionBase):
    id: int
    company_id: int
    total_amount: Decimal
    created_at: datetime
    
    class Config:
        from_attributes = True

class ReportBase(BaseModel):
    year: int
    month: int
    remarks: Optional[str] = None

class ReportCreate(ReportBase):
    company_id: int

class ReportResponse(ReportBase):
    id: int
    company_id: int
    total_emission: Decimal
    status: ReportStatus
    generated_at: datetime
    confirmed_at: Optional[datetime] = None
    generated_by: Optional[str] = None
    
    class Config:
        from_attributes = True

class MonthlySummary(BaseModel):
    company_id: int
    company_name: str
    year: int
    month: int
    total_emission: Decimal
    source_breakdown: List[dict]
    data_quality_summary: dict

class IntensityRanking(BaseModel):
    company_id: int
    company_name: str
    industry_name: str
    annual_output: Decimal
    total_emission: Decimal
    emission_intensity: Decimal
    rank: int

class QuotaBalance(BaseModel):
    company_id: int
    company_name: str
    initial_quota: Decimal
    total_bought: Decimal
    total_sold: Decimal
    current_balance: Decimal

class DeviceFaultInfo(BaseModel):
    source_id: int
    source_name: str
    company_id: int
    company_name: str
    start_time: datetime
    end_time: datetime
    fault_reason: str
