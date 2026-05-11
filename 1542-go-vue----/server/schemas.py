from pydantic import BaseModel, Field
from datetime import datetime, date
from typing import Optional, List


class CanalBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=100)
    code: str = Field(..., min_length=1, max_length=50)
    level: int = Field(..., ge=1, le=4)
    parent_id: Optional[int] = None
    design_flow: float = Field(..., gt=0)
    max_flow: float = Field(..., gt=0)


class CanalCreate(CanalBase):
    pass


class CanalUpdate(BaseModel):
    name: Optional[str] = None
    code: Optional[str] = None
    level: Optional[int] = None
    parent_id: Optional[int] = None
    design_flow: Optional[float] = None
    max_flow: Optional[float] = None


class CanalResponse(CanalBase):
    id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class GateBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=100)
    code: str = Field(..., min_length=1, max_length=50)
    canal_id: int
    max_flow: float = Field(..., gt=0)


class GateCreate(GateBase):
    pass


class GateUpdate(BaseModel):
    name: Optional[str] = None
    code: Optional[str] = None
    canal_id: Optional[int] = None
    max_flow: Optional[float] = None
    current_status: Optional[str] = None
    current_open_rate: Optional[float] = None


class GateResponse(GateBase):
    id: int
    current_status: str
    current_open_rate: float
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class QuarterlyQuotaBase(BaseModel):
    quarter: int = Field(..., ge=1, le=4)
    quota: float = Field(..., ge=0)


class QuarterlyQuotaResponse(BaseModel):
    id: int
    quarter: int
    initial_quota: float
    current_quota: float
    used_amount: float
    remaining_amount: float
    is_critical: bool
    updated_at: datetime

    class Config:
        from_attributes = True


class WaterPlanBase(BaseModel):
    canal_id: int
    year: int = Field(..., ge=2000, le=2100)
    initial_annual_quota: float = Field(..., gt=0)
    quarterly_quotas: List[QuarterlyQuotaBase]


class WaterPlanCreate(WaterPlanBase):
    pass


class WaterPlanResponse(BaseModel):
    id: int
    canal_id: int
    year: int
    initial_annual_quota: float
    current_annual_quota: float
    annual_used: float
    status: str
    quarterly_quotas: List[QuarterlyQuotaResponse]
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class QuotaAdjustment(BaseModel):
    quarter: int = Field(..., ge=1, le=4)
    new_quota: float = Field(..., ge=0)
    reason: Optional[str] = None


class WaterUsageBase(BaseModel):
    canal_id: int
    year: int
    month: int = Field(..., ge=1, le=12)
    usage_amount: float = Field(..., ge=0)


class WaterUsageCreate(WaterUsageBase):
    pass


class WaterUsageResponse(WaterUsageBase):
    id: int
    quota_amount: float
    is_over_quota: bool
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class DispatchSchemeResponse(BaseModel):
    id: int
    date: date
    canal_id: int
    gate_id: Optional[int]
    target_flow: float
    open_rate: float
    reason: Optional[str]
    is_ecological: bool
    created_at: datetime

    class Config:
        from_attributes = True


class WarningRecordResponse(BaseModel):
    id: int
    canal_id: int
    year: int
    month: int
    type: str
    message: str
    level: str
    is_resolved: bool
    created_at: datetime

    class Config:
        from_attributes = True


class WaterUseCoefficientResponse(BaseModel):
    id: int
    canal_id: int
    year: int
    quarter: int
    total_inflow: float
    total_outflow: float
    coefficient: float
    created_at: datetime

    class Config:
        from_attributes = True


class DailyDispatchSummary(BaseModel):
    date: date
    is_irrigation_season: bool
    total_schemes: int
    ecological_flow: float


class MonthlyStatistics(BaseModel):
    canal_id: int
    canal_name: str
    year: int
    month: int
    total_usage: float
    quota: float
    usage_rate: float
    is_warning: bool
    is_over_quota: bool


class QuarterlyCoefficientReport(BaseModel):
    canal_id: int
    canal_name: str
    year: int
    quarter: int
    total_inflow: float
    total_outflow: float
    coefficient: float
