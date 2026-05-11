from datetime import datetime
from enum import Enum
from typing import Optional, List
from pydantic import BaseModel, Field, validator


class SupplyChainStage(str, Enum):
    RAW_MATERIAL = "raw_material"
    PROCESSING = "processing"
    PACKAGING = "packaging"
    DISTRIBUTION = "distribution"
    RETAIL = "retail"
    ON_SHELF = "on_shelf"


class InspectionStatus(str, Enum):
    PENDING = "pending"
    PASSED = "passed"
    FAILED = "failed"


class RecallStatus(str, Enum):
    ACTIVE = "active"
    COMPLETED = "completed"


class TodoStatus(str, Enum):
    PENDING = "pending"
    COMPLETED = "completed"
    OVERDUE = "overdue"


class SupplyChainCreate(BaseModel):
    batch_number: str
    stage: SupplyChainStage
    operation_time: datetime
    operator: str
    location: Optional[str] = None
    notes: Optional[str] = None


class SupplyChainRecord(SupplyChainCreate):
    id: str
    created_at: datetime


class InspectionCreate(BaseModel):
    batch_number: str
    inspector: str
    inspection_time: datetime
    status: InspectionStatus
    report: str
    item: Optional[str] = None


class InspectionRecord(InspectionCreate):
    id: str
    created_at: datetime


class RecallCreate(BaseModel):
    batch_number: str
    reason: str
    initiator: str


class RecallRecord(BaseModel):
    id: str
    batch_number: str
    reason: str
    initiator: str
    status: RecallStatus
    start_time: datetime
    end_time: Optional[datetime] = None
    can_track: bool
    affected_stages: List[str]
    created_at: datetime


class TodoUpdate(BaseModel):
    status: TodoStatus
    processed_by: Optional[str] = None
    notes: Optional[str] = None


class TodoItem(BaseModel):
    id: str
    recall_id: str
    batch_number: str
    stage: str
    handler: str
    status: TodoStatus
    created_at: datetime
    processed_at: Optional[datetime] = None
    processed_by: Optional[str] = None
    notes: Optional[str] = None


class TraceabilityExport(BaseModel):
    batch_number: str
    supply_chain_records: List[SupplyChainRecord]
    inspection_records: List[InspectionRecord]
    export_time: datetime
