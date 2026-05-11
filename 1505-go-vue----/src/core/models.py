from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from typing import List, Optional
from uuid import uuid4


def _generate_id() -> str:
    return str(uuid4())


class QCResult(str, Enum):
    PASS = "pass"
    FAIL = "fail"


class DispositionType(str, Enum):
    REWORK = "rework"
    DOWNGRADE = "downgrade"
    SCRAP = "scrap"


class TodoStatus(str, Enum):
    PENDING = "pending"
    RESOLVED = "resolved"


@dataclass
class RawMaterialItem:
    name: str
    quantity: float
    is_key: bool


@dataclass
class RawMaterial:
    id: str = field(default_factory=_generate_id)
    name: str = ""
    current_stock: float = 0.0
    safety_stock: float = 0.0
    created_at: datetime = field(default_factory=datetime.utcnow)
    updated_at: datetime = field(default_factory=datetime.utcnow)

    @property
    def is_below_safety(self) -> bool:
        return self.current_stock < self.safety_stock


@dataclass
class Recipe:
    id: str = field(default_factory=_generate_id)
    name: str = ""
    category: str = ""
    raw_materials: List[RawMaterialItem] = field(default_factory=list)
    created_at: datetime = field(default_factory=datetime.utcnow)


@dataclass
class QualityCheck:
    id: str = field(default_factory=_generate_id)
    batch_id: str = ""
    inspector: str = ""
    result: QCResult = QCResult.PASS
    measured_value: Optional[float] = None
    notes: str = ""
    created_at: datetime = field(default_factory=datetime.utcnow)


@dataclass
class Todo:
    id: str = field(default_factory=_generate_id)
    batch_id: str = ""
    status: TodoStatus = TodoStatus.PENDING
    disposition: Optional[DispositionType] = None
    resolved_by: Optional[str] = None
    resolved_at: Optional[datetime] = None
    created_at: datetime = field(default_factory=datetime.utcnow)


@dataclass
class ProductionBatch:
    id: str = field(default_factory=_generate_id)
    recipe_id: str = ""
    plan_quantity: int = 0
    actual_quantity: int = 0
    quality_checks: List[QualityCheck] = field(default_factory=list)
    todos: List[Todo] = field(default_factory=list)
    created_at: datetime = field(default_factory=datetime.utcnow)

    @property
    def has_passed_qc(self) -> bool:
        return any(qc.result == QCResult.PASS for qc in self.quality_checks)

    @property
    def has_pending_todo(self) -> bool:
        return any(todo.status == TodoStatus.PENDING for todo in self.todos)
