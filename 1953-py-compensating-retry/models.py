from enum import Enum
from dataclasses import dataclass, field
from typing import Dict, Any, List, Optional
from datetime import datetime
import uuid


class TransactionStatus(str, Enum):
    PENDING = "pending"
    EXECUTING = "executing"
    COMPLETED = "completed"
    COMPENSATING = "compensating"
    COMPENSATED = "compensated"
    COMPENSATION_FAILED = "compensation_failed"


class StepStatus(str, Enum):
    PENDING = "pending"
    SUCCESS = "success"
    FAILED = "failed"
    COMPENSATED = "compensated"
    COMPENSATION_FAILED = "compensation_failed"


@dataclass
class Step:
    id: str
    operation_type: str
    params: Dict[str, Any]
    status: StepStatus = StepStatus.PENDING
    result: Optional[Dict[str, Any]] = None
    error_message: Optional[str] = None
    compensation_status: Optional[str] = None
    compensation_error: Optional[str] = None
    created_at: datetime = field(default_factory=datetime.utcnow)
    executed_at: Optional[datetime] = None
    compensated_at: Optional[datetime] = None

    def to_dict(self) -> Dict[str, Any]:
        return {
            "id": self.id,
            "operation_type": self.operation_type,
            "params": self.params,
            "status": self.status.value,
            "result": self.result,
            "error_message": self.error_message,
            "compensation_status": self.compensation_status,
            "compensation_error": self.compensation_error,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "executed_at": self.executed_at.isoformat() if self.executed_at else None,
            "compensated_at": self.compensated_at.isoformat() if self.compensated_at else None,
        }


@dataclass
class Transaction:
    id: str
    status: TransactionStatus = TransactionStatus.PENDING
    steps: List[Step] = field(default_factory=list)
    error_message: Optional[str] = None
    created_at: datetime = field(default_factory=datetime.utcnow)
    updated_at: datetime = field(default_factory=datetime.utcnow)

    def to_dict(self) -> Dict[str, Any]:
        return {
            "id": self.id,
            "status": self.status.value,
            "steps": [step.to_dict() for step in self.steps],
            "error_message": self.error_message,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None,
        }

    @classmethod
    def create(cls, steps_data: List[Dict[str, Any]]) -> "Transaction":
        steps = []
        for step_data in steps_data:
            step = Step(
                id=str(uuid.uuid4()),
                operation_type=step_data["operation_type"],
                params=step_data.get("params", {})
            )
            steps.append(step)
        return cls(
            id=str(uuid.uuid4()),
            steps=steps
        )
