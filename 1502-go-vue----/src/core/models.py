from datetime import datetime, date, time
from typing import Optional, Dict, Any
from pydantic import BaseModel, Field, validator


class Pond(BaseModel):
    id: Optional[int] = None
    code: str
    area: float
    species: str
    stock_quantity: int
    created_at: datetime = Field(default_factory=datetime.now)
    updated_at: datetime = Field(default_factory=datetime.now)


class WaterQualityThreshold(BaseModel):
    id: Optional[int] = None
    pond_id: int
    parameter: str
    min_value: Optional[float] = None
    max_value: Optional[float] = None


class WaterQualityRecord(BaseModel):
    id: Optional[int] = None
    pond_id: int
    water_temp: float
    dissolved_oxygen: float
    ph: float
    ammonia: float
    recorded_at: datetime = Field(default_factory=datetime.now)
    anomalies: Dict[str, bool] = Field(default_factory=dict)


class FeedingPlan(BaseModel):
    id: Optional[int] = None
    pond_id: int
    feed_date: date
    feed_time: time
    amount: float
    created_at: datetime = Field(default_factory=datetime.now)

    @validator('amount')
    def check_amount(cls, v):
        if v <= 0:
            raise ValueError('投喂量必须大于0')
        return v


class FeedingRecord(BaseModel):
    id: Optional[int] = None
    pond_id: int
    plan_id: Optional[int] = None
    feed_date: date
    feed_time: time
    amount: float
    created_at: datetime = Field(default_factory=datetime.now)


class HarvestRecord(BaseModel):
    id: Optional[int] = None
    pond_id: int
    species: str
    quantity: int
    weight: float
    harvested_at: datetime = Field(default_factory=datetime.now)


class WaterQualityAggregation(BaseModel):
    parameter: str
    min_value: float
    max_value: float
    avg_value: float
