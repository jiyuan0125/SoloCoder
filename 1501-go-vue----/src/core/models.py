from enum import Enum
from datetime import datetime
from typing import Optional
from pydantic import BaseModel, Field, field_validator


class GreenhouseType(str, Enum):
    GLASS = "glass"
    PLASTIC = "plastic"


class PlanStatus(str, Enum):
    PENDING = "pending"
    EXECUTED = "executed"


class Greenhouse(BaseModel):
    id: str
    code: str
    area: float = Field(gt=0)
    crop: str
    type: GreenhouseType


class GreenhouseCreate(BaseModel):
    code: str
    area: float = Field(gt=0)
    crop: str
    type: GreenhouseType


class EnvironmentData(BaseModel):
    greenhouse_id: str
    timestamp: datetime
    temperature: float
    humidity: float
    soil_moisture: float
    light: float


class EnvironmentDataCreate(BaseModel):
    temperature: float
    humidity: float
    soil_moisture: float
    light: float


class IrrigationPlan(BaseModel):
    id: str
    greenhouse_id: str
    execution_time: datetime
    water_amount: float = Field(gt=0)
    status: PlanStatus
    created_at: datetime


class IrrigationPlanCreate(BaseModel):
    execution_time: datetime
    water_amount: float = Field(gt=0)


class FertilizerPlan(BaseModel):
    id: str
    greenhouse_id: str
    execution_time: datetime
    fertilizer_amount: float = Field(gt=0)
    status: PlanStatus
    created_at: datetime


class FertilizerPlanCreate(BaseModel):
    execution_time: datetime
    fertilizer_amount: float = Field(gt=0)

    @field_validator("fertilizer_amount")
    @classmethod
    def check_fertilizer_amount(cls, v: float) -> float:
        if v <= 0:
            raise ValueError("施肥用量不能为零")
        return v


class AggregatedData(BaseModel):
    greenhouse_id: str
    metric: str
    start_time: datetime
    end_time: datetime
    max_value: float
    min_value: float
    avg_value: float
