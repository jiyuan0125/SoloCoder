from pydantic import BaseModel, Field
from typing import Dict, Optional
from datetime import datetime


class LimiterConfig(BaseModel):
    global_qps: int = Field(default=10000, gt=0)
    tenant_qps: Dict[str, int] = Field(default_factory=dict)
    api_qps: Dict[str, int] = Field(default_factory=dict)
    default_tenant_qps: int = Field(default=100, gt=0)
    default_api_qps: int = Field(default=500, gt=0)
    max_borrow_per_tenant: int = Field(default=200, gt=0)


class TenantQuotaInfo(BaseModel):
    tenant_id: str
    base_quota: int
    current_usage: int
    borrowed: int
    remaining: int
    available_borrow: int


class GlobalQuotaInfo(BaseModel):
    total_quota: int
    current_usage: int
    remaining: int
    shared_pool: int


class APIQuotaInfo(BaseModel):
    api_path: str
    quota: int
    current_usage: int
    remaining: int


class StatusResponse(BaseModel):
    global_status: GlobalQuotaInfo
    tenants: Dict[str, TenantQuotaInfo]
    apis: Dict[str, APIQuotaInfo]


class TenantStatsResponse(BaseModel):
    tenant_id: str
    base_quota: int
    current_usage: int
    borrowed: int
    remaining: int
    available_borrow: int
    shared_pool_atm: int
    max_borrow_limit: int


class ConfigUpdate(BaseModel):
    global_qps: Optional[int] = Field(default=None, gt=0)
    tenant_qps: Optional[Dict[str, int]] = None
    api_qps: Optional[Dict[str, int]] = None
    default_tenant_qps: Optional[int] = Field(default=None, gt=0)
    default_api_qps: Optional[int] = Field(default=None, gt=0)
    max_borrow_per_tenant: Optional[int] = Field(default=None, gt=0)


class RateLimitExceeded(BaseModel):
    level: str
    message: str
    remaining_quota: int
    retry_after: float
