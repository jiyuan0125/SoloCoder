from datetime import datetime
from decimal import Decimal
from typing import Optional, List
from pydantic import BaseModel, Field


class StationBase(BaseModel):
    name: str
    price_zone: int = Field(ge=1, le=8)


class StationCreate(StationBase):
    pass


class StationUpdate(BaseModel):
    name: Optional[str] = None
    price_zone: Optional[int] = Field(None, ge=1, le=8)


class Station(StationBase):
    id: int
    created_at: datetime
    updated_at: datetime
    
    class Config:
        from_attributes = True


class RateBase(BaseModel):
    from_zone: int = Field(ge=1, le=8)
    to_zone: int = Field(ge=1, le=8)
    rate_per_ton_fen: int = Field(ge=0)


class RateCreate(RateBase):
    pass


class RateUpdate(BaseModel):
    rate_per_ton_fen: Optional[int] = Field(None, ge=0)


class Rate(RateBase):
    id: int
    effective_from: datetime
    effective_to: Optional[datetime] = None
    created_at: datetime
    updated_at: datetime
    
    class Config:
        from_attributes = True


class WaybillCreate(BaseModel):
    from_station_name: str
    to_station_name: str
    weight_ton: Decimal = Field(gt=0)
    is_hazardous: bool = False
    customer_code: Optional[str] = None


class WaybillListQuery(BaseModel):
    start_date: Optional[datetime] = None
    end_date: Optional[datetime] = None
    status: Optional[str] = None
    customer_code: Optional[str] = None


class Waybill(BaseModel):
    id: int
    waybill_no: str
    from_station_name: str
    to_station_name: str
    weight_ton: Decimal
    is_hazardous: bool
    customer_code: Optional[str]
    
    basic_rate_fen: int
    basic_charge_fen: int
    hazardous_surcharge_fen: int
    discount_fen: int
    total_charge_fen: int
    
    status: str
    settled_at: Optional[datetime] = None
    
    created_at: datetime
    updated_at: datetime
    
    class Config:
        from_attributes = True


class SettlementResponse(BaseModel):
    waybill_id: int
    waybill_no: str
    old_total_charge_fen: int
    new_total_charge_fen: int
    adjustment_fen: int
    has_adjustment: bool


class MonthlyReport(BaseModel):
    year_month: str
    total_count: int
    total_weight_ton: Decimal
    total_amount_fen: int
    settled_amount_fen: int
    unsettled_amount_fen: int


class CustomerMonthlyStat(BaseModel):
    customer_code: str
    year_month: str
    total_count: int
    total_amount_fen: int
    settled_amount_fen: int
    
    class Config:
        from_attributes = True
