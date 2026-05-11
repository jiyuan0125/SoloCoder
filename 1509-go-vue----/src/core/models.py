from datetime import datetime
from enum import Enum
from typing import Optional, List
from pydantic import BaseModel, Field, validator


class SupplyChainStage(str, Enum):
    RAW_MATERIAL = "raw_material"
    PRODUCTION = "production"
    PROCESSING = "processing"
    PACKAGING = "packaging"
    WAREHOUSE = "warehouse"
    DISTRIBUTION = "distribution"
    RETAIL = "retail"


class InspectionStatus(str, Enum):
    PENDING = "pending"
    PASSED = "passed"
    FAILED = "failed"


class RecallStatus(str, Enum):
    ACTIVE = "active"
    COMPLETED = "completed"
    CANCELLED = "cancelled"


class TodoStatus(str, Enum):
    PENDING = "pending"
    CONFIRMED = "confirmed"
    OVERDUE = "overdue"


class SupplyChainRecord(BaseModel):
    id: Optional[str] = None
    batch_number: str
    stage: SupplyChainStage
    operation_time: datetime
    operator: str
    location: Optional[str] = None
    remarks: Optional[str] = None

    @validator('operation_time')
    def operation_time_cannot_be_future(cls, v: datetime) -> datetime:
        if v > datetime.now():
            raise ValueError("操作时间不能是未来时间")
        return v


class InspectionRecord(BaseModel):
    id: Optional[str] = None
    batch_number: str
    inspection_time: datetime
    inspector: str
    status: InspectionStatus
    items: List[str] = []
    report: Optional[str] = None
    remarks: Optional[str] = None

    @validator('inspection_time')
    def inspection_time_cannot_be_future(cls, v: datetime) -> datetime:
        if v > datetime.now():
            raise ValueError("检测时间不能是未来时间")
        return v


class Recall(BaseModel):
    id: Optional[str] = None
    batch_number: str
    reason: str
    created_at: datetime = Field(default_factory=datetime.now)
    status: RecallStatus = RecallStatus.ACTIVE
    tracked_stages: List[SupplyChainStage] = []
    can_track: bool = True
    completed_at: Optional[datetime] = None
    completed_by: Optional[str] = None


class TodoItem(BaseModel):
    id: Optional[str] = None
    recall_id: str
    batch_number: str
    stage: SupplyChainStage
    operator: str
    created_at: datetime = Field(default_factory=datetime.now)
    status: TodoStatus = TodoStatus.PENDING
    confirmed_at: Optional[datetime] = None
    confirmed_by: Optional[str] = None
    remarks: Optional[str] = None
