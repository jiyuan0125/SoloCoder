from pydantic import BaseModel, Field
from typing import List, Optional

from core.models import DispositionType, QCResult


class RawMaterialItemIn(BaseModel):
    name: str
    quantity: float
    is_key: bool


class RecipeIn(BaseModel):
    name: str
    category: str
    raw_materials: List[RawMaterialItemIn]


class RawMaterialItemOut(BaseModel):
    name: str
    quantity: float
    is_key: bool


class RecipeOut(BaseModel):
    id: str
    name: str
    category: str
    raw_materials: List[RawMaterialItemOut]
    created_at: str


class RawMaterialIn(BaseModel):
    name: str
    current_stock: float = Field(ge=0)
    safety_stock: float = Field(ge=0)


class RawMaterialOut(BaseModel):
    id: str
    name: str
    current_stock: float
    safety_stock: float
    is_below_safety: bool
    created_at: str
    updated_at: str


class ProductionBatchIn(BaseModel):
    recipe_id: str
    plan_quantity: int = Field(gt=0)
    actual_quantity: int = Field(ge=0)


class QualityCheckIn(BaseModel):
    inspector: str
    result: QCResult
    measured_value: Optional[float] = None
    notes: str = ""


class QualityCheckOut(BaseModel):
    id: str
    batch_id: str
    inspector: str
    result: str
    measured_value: Optional[float]
    notes: str
    created_at: str


class TodoOut(BaseModel):
    id: str
    batch_id: str
    status: str
    disposition: Optional[str]
    resolved_by: Optional[str]
    resolved_at: Optional[str]
    created_at: str


class ProductionBatchOut(BaseModel):
    id: str
    recipe_id: str
    plan_quantity: int
    actual_quantity: int
    has_passed_qc: bool
    has_pending_todo: bool
    quality_checks: List[QualityCheckOut]
    todos: List[TodoOut]
    created_at: str


class ResolveTodoIn(BaseModel):
    disposition: DispositionType
    resolved_by: str
