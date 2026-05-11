from datetime import date, datetime, time
from enum import Enum
from typing import List, Optional

from pydantic import BaseModel, Field, validator


class EquipmentStatus(str, Enum):
    RUNNING = "running"
    STOPPED = "stopped"
    MAINTENANCE = "maintenance"


class EquipmentType(str, Enum):
    CRUSHER = "crusher"
    FLOTATION = "flotation"


class Equipment(BaseModel):
    id: str
    name: str
    type: EquipmentType
    status: EquipmentStatus = EquipmentStatus.STOPPED
    total_running_hours: float = 0.0
    last_running_time: Optional[datetime] = None
    created_at: datetime = Field(default_factory=datetime.now)
    updated_at: datetime = Field(default_factory=datetime.now)


class EquipmentCreate(BaseModel):
    id: str
    name: str
    type: EquipmentType


class EquipmentUpdateStatus(BaseModel):
    status: EquipmentStatus
    timestamp: datetime = Field(default_factory=datetime.now)


class CrushingRecord(BaseModel):
    id: str
    crusher_id: str
    record_date: date
    feed_grade_size: float
    product_grade_size: float
    throughput: float
    created_at: datetime = Field(default_factory=datetime.now)

    @validator("product_grade_size")
    def product_size_less_than_feed(cls, v, values, **kwargs):
        if "feed_grade_size" in values and v >= values["feed_grade_size"]:
            raise ValueError(
                "破碎后粒度不能大于破碎前粒度"
            )
        return v

    @validator("throughput")
    def throughput_positive(cls, v):
        if v <= 0:
            raise ValueError("处理量必须大于0")
        return v


class CrushingRecordCreate(BaseModel):
    crusher_id: str
    record_date: date
    feed_grade_size: float
    product_grade_size: float
    throughput: float


class FlotationRecord(BaseModel):
    id: str
    date: date
    feed_grade: float
    concentrate_grade: float
    tailings_grade: float
    recovery: float
    created_at: datetime = Field(default_factory=datetime.now)

    @validator("concentrate_grade")
    def concentrate_grade_higher_than_feed(cls, v, values, **kwargs):
        if "feed_grade" in values and v <= values["feed_grade"]:
            raise ValueError(
                "精矿品位不能低于原矿品位"
            )
        return v

    @validator("recovery")
    def recovery_within_range(cls, v):
        if v < 0 or v > 100:
            raise ValueError(
                "回收率必须在0-100之间"
            )
        return v


class FlotationRecordCreate(BaseModel):
    date: date
    feed_grade: float
    concentrate_grade: float
    tailings_grade: float
    recovery: float


class GradeCalculationInput(BaseModel):
    feed_grade: float
    concentrate_grade: float
    tailings_grade: float
    feed_throughput: Optional[float] = None


class GradeCalculationResult(BaseModel):
    theoretical_recovery: float
    theoretical_concentrate_yield: Optional[float] = None


class DailyMetrics(BaseModel):
    date: date
    total_throughput: float
    average_recovery: float
    equipment_utilization: float


class EquipmentUtilization(BaseModel):
    equipment_id: str
    equipment_name: str
    utilization_rate: float
    running_hours: float
