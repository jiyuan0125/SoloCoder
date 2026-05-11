from datetime import datetime, date, time
from typing import List, Optional
from pydantic import BaseModel, Field, field_validator


class PondBase(BaseModel):
    code: str = Field(..., min_length=1, max_length=50)
    area: float = Field(..., gt=0)
    species: str = Field(..., min_length=1, max_length=100)
    initial_stock: int = Field(..., gt=0)


class PondCreate(PondBase):
    pass


class PondUpdate(BaseModel):
    area: Optional[float] = Field(None, gt=0)
    species: Optional[str] = Field(None, min_length=1, max_length=100)


class PondResponse(PondBase):
    id: int
    current_stock: int
    created_at: datetime

    class Config:
        from_attributes = True


class PondList(BaseModel):
    items: List[PondResponse]


class ThresholdBase(BaseModel):
    indicator: str
    min_value: Optional[float] = None
    max_value: Optional[float] = None


class ThresholdCreate(ThresholdBase):
    pass


class ThresholdResponse(ThresholdBase):
    id: int
    pond_id: int
    created_at: datetime

    class Config:
        from_attributes = True


class WaterRecordBase(BaseModel):
    temperature: float
    dissolved_oxygen: float
    ph: float
    ammonia_nitrogen: float


class WaterRecordCreate(WaterRecordBase):
    pass


class WaterRecordResponse(WaterRecordBase):
    id: int
    pond_id: int
    is_abnormal: bool
    abnormal_indicators: Optional[str]
    recorded_at: datetime

    class Config:
        from_attributes = True


class WaterStats(BaseModel):
    indicator: str
    count: int
    min_value: float
    max_value: float
    avg_value: float


class WaterStatsResponse(BaseModel):
    pond_id: int
    start_time: datetime
    end_time: datetime
    stats: List[WaterStats]


class FeedingPlanBase(BaseModel):
    plan_date: date
    plan_time: time
    feed_amount: float

    @field_validator("feed_amount")
    @classmethod
    def validate_feed_amount(cls, v: float) -> float:
        if v <= 0:
            raise ValueError("投喂量必须大于0")
        return v


class FeedingPlanCreate(FeedingPlanBase):
    pass


class FeedingPlanResponse(FeedingPlanBase):
    id: int
    pond_id: int
    is_executed: bool
    created_at: datetime

    class Config:
        from_attributes = True


class FeedingRecordResponse(BaseModel):
    id: int
    pond_id: int
    plan_id: Optional[int]
    feed_amount: float
    executed_at: datetime

    class Config:
        from_attributes = True


class HarvestBase(BaseModel):
    species: str = Field(..., min_length=1, max_length=100)
    quantity: int = Field(..., gt=0)
    weight: float = Field(..., gt=0)


class HarvestCreate(HarvestBase):
    pass


class HarvestResponse(HarvestBase):
    id: int
    pond_id: int
    harvested_at: datetime

    class Config:
        from_attributes = True


class ErrorResponse(BaseModel):
    detail: str
