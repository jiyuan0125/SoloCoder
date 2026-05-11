from datetime import date, datetime
from typing import Optional
from enum import Enum

from pydantic import BaseModel, Field, field_validator


class WineStatus(str, Enum):
    FERMENTING = "fermenting"
    AGING = "aging"
    AGING_READY = "aging_ready"
    BOTTLED = "bottled"


class PlotBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=100)
    grape_variety: str = Field(..., min_length=1, max_length=100)
    planting_year: int = Field(..., ge=1900, le=2100)


class PlotCreate(PlotBase):
    pass


class Plot(PlotBase):
    id: str
    created_at: datetime

    class Config:
        from_attributes = True


class HarvestBase(BaseModel):
    plot_id: str
    harvest_date: date
    quantity: float = Field(..., gt=0)
    brix: float = Field(..., gt=0)
    acidity: float = Field(..., gt=0)

    @field_validator('brix', 'acidity')
    @classmethod
    def check_positive(cls, v: float) -> float:
        if v <= 0:
            raise ValueError("Must be a positive number")
        return v


class HarvestCreate(HarvestBase):
    pass


class Harvest(HarvestBase):
    id: str
    created_at: datetime
    used_quantity: float = 0.0

    @property
    def available_quantity(self) -> float:
        return max(0.0, self.quantity - self.used_quantity)

    class Config:
        from_attributes = True


class BatchBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=100)
    description: Optional[str] = None
    harvest_ids: list[str]
    harvest_quantities: list[float]
    fermentation_start: date
    fermentation_end: Optional[date] = None

    @field_validator('harvest_quantities')
    @classmethod
    def check_quantities_positive(cls, v: list[float]) -> list[float]:
        for q in v:
            if q <= 0:
                raise ValueError("All harvest quantities must be positive")
        return v


class BatchCreate(BatchBase):
    pass


class Batch(BatchBase):
    id: str
    created_at: datetime
    status: WineStatus = WineStatus.FERMENTING

    class Config:
        from_attributes = True


class CellarBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=100)
    location: Optional[str] = None
    total_slots: int = Field(..., gt=0)


class CellarCreate(CellarBase):
    pass


class Cellar(CellarBase):
    id: str
    created_at: datetime

    class Config:
        from_attributes = True


class StorageBase(BaseModel):
    batch_id: str
    cellar_id: str
    shelf_number: int = Field(..., gt=0)
    position: str = Field(..., min_length=1, max_length=50)
    expected_aging_months: int = Field(..., gt=0)
    start_date: date
    expected_end_date: Optional[date] = None


class StorageCreate(StorageBase):
    pass


class Storage(StorageBase):
    id: str
    created_at: datetime
    is_active: bool = True

    class Config:
        from_attributes = True


class TastingBase(BaseModel):
    batch_id: str
    tasting_date: date
    notes: Optional[str] = None


class TastingCreate(TastingBase):
    decision: str = Field(..., pattern="^(continue_aging|bottle)$")
    additional_months: Optional[int] = Field(None, ge=1)

    @field_validator('additional_months')
    @classmethod
    def check_additional_months(cls, v: Optional[int], info) -> Optional[int]:
        if info.data.get('decision') == 'continue_aging' and (v is None or v <= 0):
            raise ValueError("additional_months must be positive when continuing aging")
        if info.data.get('decision') == 'bottle':
            return None
        return v


class Tasting(TastingBase):
    id: str
    created_at: datetime
    decision: str
    additional_months: Optional[int] = None

    class Config:
        from_attributes = True


class TodoItem(BaseModel):
    batch_id: str
    batch_name: str
    storage_id: str
    expected_end_date: date
    days_overdue: int
