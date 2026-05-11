from datetime import datetime
from enum import Enum
from typing import List, Optional
from pydantic import BaseModel, Field, field_validator
from decimal import Decimal


class TodoStatus(str, Enum):
    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    RESOLVED = "resolved"
    OVERDUE = "overdue"


class TodoResolution(str, Enum):
    REWORK = "rework"
    DOWNGRADE = "downgrade"
    SCRAP = "scrap"


class InspectionStatus(str, Enum):
    PASS = "pass"
    FAIL = "fail"


class BatchStatus(str, Enum):
    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"


class RecipeStatus(str, Enum):
    ACTIVE = "active"
    INACTIVE = "inactive"


class RawMaterialCreate(BaseModel):
    name: str
    percentage: Decimal = Field(ge=0, le=100)


class RawMaterial(RawMaterialCreate):
    id: int
    recipe_id: int

    class Config:
        from_attributes = True


class RecipeCreate(BaseModel):
    name: str
    description: Optional[str] = None
    raw_materials: List[RawMaterialCreate]

    @field_validator("raw_materials")
    @classmethod
    def validate_raw_materials(cls, v: List[RawMaterialCreate]) -> List[RawMaterialCreate]:
        if len(v) < 2:
            raise ValueError("配方原材料清单至少需要两条")
        total = sum(m.percentage for m in v)
        if total != Decimal("100"):
            raise ValueError(f"原材料占比总和必须为100%，当前为{total}%")
        return v


class RecipeUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None


class Recipe(BaseModel):
    id: int
    name: str
    description: Optional[str] = None
    status: RecipeStatus
    created_at: datetime
    raw_materials: List[RawMaterial] = []

    class Config:
        from_attributes = True


class ProductionBatchCreate(BaseModel):
    recipe_id: int
    planned_quantity: Decimal = Field(gt=0)
    actual_quantity: Optional[Decimal] = Field(default=None, ge=0)

    @field_validator("actual_quantity")
    @classmethod
    def validate_actual_quantity(cls, v: Optional[Decimal], info) -> Optional[Decimal]:
        if v is None:
            return v
        planned = info.data.get("planned_quantity")
        if planned is not None:
            max_allowed = planned * Decimal("1.05")
            if v > max_allowed:
                raise ValueError(f"实际产量不能超过计划的105%（最大{max_allowed}）")
        return v


class ProductionBatchUpdate(BaseModel):
    actual_quantity: Optional[Decimal] = Field(default=None, ge=0)
    status: Optional[BatchStatus] = None


class ProductionBatch(BaseModel):
    id: int
    recipe_id: int
    planned_quantity: Decimal
    actual_quantity: Optional[Decimal] = None
    status: BatchStatus
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class QualityInspectionCreate(BaseModel):
    batch_id: int
    status: InspectionStatus
    notes: Optional[str] = None


class QualityInspection(BaseModel):
    id: int
    batch_id: int
    status: InspectionStatus
    notes: Optional[str] = None
    created_at: datetime

    class Config:
        from_attributes = True


class TodoItemUpdate(BaseModel):
    status: TodoStatus
    resolution: Optional[TodoResolution] = None
    notes: Optional[str] = None


class TodoItem(BaseModel):
    id: int
    batch_id: int
    status: TodoStatus
    resolution: Optional[TodoResolution] = None
    notes: Optional[str] = None
    created_at: datetime
    resolved_at: Optional[datetime] = None
    is_overdue: bool = False

    class Config:
        from_attributes = True
