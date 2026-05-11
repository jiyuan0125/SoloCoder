from datetime import datetime, date
from enum import Enum
from typing import Optional, Dict
from pydantic import BaseModel, Field
from uuid import uuid4, UUID


class AreaType(str, Enum):
    GRAIN = "grain"
    OIL = "oil"
    MIXED = "mixed"


class GrainVariety(str, Enum):
    WHEAT = "wheat"
    RICE = "rice"
    CORN = "corn"
    SOYBEAN = "soybean"
    RAPESEED_OIL = "rapeseed_oil"
    SOYBEAN_OIL = "soybean_oil"
    OTHER = "other"


VARIETY_SHELF_LIFE_DAYS: Dict[GrainVariety, int] = {
    GrainVariety.WHEAT: 365 * 2,
    GrainVariety.RICE: 365 * 1,
    GrainVariety.CORN: 365 * 1,
    GrainVariety.SOYBEAN: 365 * 1,
    GrainVariety.RAPESEED_OIL: 365 * 1,
    GrainVariety.SOYBEAN_OIL: 365 * 1,
    GrainVariety.OTHER: 365,
}


def get_shelf_life_days(variety: GrainVariety) -> int:
    return VARIETY_SHELF_LIFE_DAYS.get(variety, 365)


class Storehouse(BaseModel):
    id: UUID = Field(default_factory=uuid4)
    name: str
    location: str
    created_at: datetime = Field(default_factory=datetime.utcnow)


class WarehouseArea(BaseModel):
    id: UUID = Field(default_factory=uuid4)
    storehouse_id: UUID
    name: str
    area_type: AreaType
    capacity_kg: float
    current_used_kg: float = 0.0
    min_temp: float = 10.0
    max_temp: float = 25.0
    created_at: datetime = Field(default_factory=datetime.utcnow)


class StockInRecord(BaseModel):
    id: UUID = Field(default_factory=uuid4)
    area_id: UUID
    variety: GrainVariety
    quantity_kg: float
    remaining_kg: float
    source: str
    batch_no: str
    in_date: date
    expiry_date: date
    created_at: datetime = Field(default_factory=datetime.utcnow)


class StockOutRecord(BaseModel):
    id: UUID = Field(default_factory=uuid4)
    stock_in_id: UUID
    area_id: UUID
    variety: GrainVariety
    quantity_kg: float
    destination: str
    purpose: str
    out_date: date = Field(default_factory=date.today)
    created_at: datetime = Field(default_factory=datetime.utcnow)


class TemperatureRecord(BaseModel):
    id: UUID = Field(default_factory=uuid4)
    area_id: UUID
    temperature: float
    record_time: datetime
    is_alarm: bool = False
    created_at: datetime = Field(default_factory=datetime.utcnow)


class PestInspectionRecord(BaseModel):
    id: UUID = Field(default_factory=uuid4)
    area_id: UUID
    variety: Optional[GrainVariety] = None
    pest_count_per_kg: float
    inspection_date: date = Field(default_factory=date.today)
    notes: Optional[str] = None
    created_at: datetime = Field(default_factory=datetime.utcnow)


class TodoItem(BaseModel):
    id: UUID = Field(default_factory=uuid4)
    area_id: UUID
    title: str
    description: str
    source_type: str
    source_id: Optional[UUID] = None
    is_completed: bool = False
    created_at: datetime = Field(default_factory=datetime.utcnow)
