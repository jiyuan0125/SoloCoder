from datetime import datetime
from enum import Enum
from typing import List, Optional

from pydantic import BaseModel, Field


class InspectionResult(str, Enum):
    PASS = "合格"
    FAIL = "不合格"


class TodoStatus(str, Enum):
    PENDING = "待处理"
    REWORK = "返工"
    DOWNGRADE = "降级"
    SCRAP = "报废"


class MaterialInventory(BaseModel):
    id: int
    name: str
    current_stock: float
    safety_stock: float


class RecipeIngredient(BaseModel):
    material_name: str
    quantity: float
    is_key_material: bool


class RecipeBase(BaseModel):
    name: str
    product_category: str
    ingredients: List[RecipeIngredient]


class RecipeCreate(RecipeBase):
    pass


class Recipe(RecipeBase):
    id: int
    created_at: datetime


class BatchBase(BaseModel):
    recipe_id: int
    planned_quantity: float
    actual_quantity: Optional[float] = None
    production_date: Optional[datetime] = None


class BatchCreate(BatchBase):
    pass


class Batch(BatchBase):
    id: int
    created_at: datetime
    has_passed_inspection: bool = False


class QualityInspectionBase(BaseModel):
    batch_id: int
    result: InspectionResult
    inspection_value: Optional[float] = None
    inspector: Optional[str] = None
    notes: Optional[str] = None


class QualityInspectionCreate(QualityInspectionBase):
    pass


class QualityInspection(QualityInspectionBase):
    id: int
    created_at: datetime


class TodoBase(BaseModel):
    batch_id: int
    status: TodoStatus = TodoStatus.PENDING
    assigned_to: Optional[str] = None
    notes: Optional[str] = None


class TodoCreate(TodoBase):
    pass


class TodoUpdate(BaseModel):
    status: TodoStatus
    notes: Optional[str] = None


class Todo(TodoBase):
    id: int
    created_at: datetime
    updated_at: datetime


class BatchSummary(BaseModel):
    batch_id: int
    recipe_name: str
    product_category: str
    planned_quantity: float
    actual_quantity: Optional[float]
    production_date: Optional[datetime]
    inspection_result: Optional[str]
    todo_status: Optional[str]


class ValidationError(Exception):
    pass
