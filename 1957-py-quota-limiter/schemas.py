from pydantic import BaseModel
from typing import Optional, List, Dict
from datetime import datetime


class QuotaSetRequest(BaseModel):
    user_id: str
    resource_type: str
    limit: int


class QuotaSetResponse(BaseModel):
    user_id: str
    resource_type: str
    limit: int
    status: str


class QuotaReserveRequest(BaseModel):
    user_id: str
    resource_type: str
    reserved_amount: int


class QuotaReserveResponse(BaseModel):
    user_id: str
    resource_type: str
    reserved_amount: int
    reserved_used: int
    status: str


class QuotaCheckResponse(BaseModel):
    user_id: str
    resource_type: str
    total_limit: int
    total_used: int
    remaining: int
    reserved_amount: int
    reserved_used: int
    can_consume: bool


class ResourceUsageItem(BaseModel):
    resource_type: str
    total_limit: int
    total_used: int
    remaining: int
    reserved_amount: int
    reserved_used: int


class UserUsageResponse(BaseModel):
    user_id: str
    usage: List[ResourceUsageItem]


class StatsResponse(BaseModel):
    total_created: int
    total_released: int
    total_rejected: int
    start_time: datetime
    end_time: datetime


class TopUserItem(BaseModel):
    user_id: str
    usage: int


class ResourceStatsItem(BaseModel):
    resource_type: str
    total_created: int
    total_released: int
    total_rejected: int
    top_users: List[TopUserItem]


class ResourceStatsResponse(BaseModel):
    resources: List[ResourceStatsItem]
    start_time: datetime
    end_time: datetime


class ConsumeRequest(BaseModel):
    user_id: str
    resource_type: str
    amount: int = 1


class ConsumeResponse(BaseModel):
    user_id: str
    resource_type: str
    amount: int
    success: bool
    remaining: int
    message: Optional[str] = None


class ReleaseRequest(BaseModel):
    user_id: str
    resource_type: str
    amount: int = 1


class ReleaseResponse(BaseModel):
    user_id: str
    resource_type: str
    amount: int
    success: bool
    released_from_reserved: int
    released_from_shared: int
