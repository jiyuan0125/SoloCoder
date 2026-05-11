from typing import List, Optional
from datetime import datetime
from pydantic import BaseModel, Field, validator
from server.enums import (
    AgentServiceStage,
    StageStatus,
    BerthStatus,
    BerthApplicationStatus,
    SupplyType,
    SupplyStatus,
    WasteStatus,
    TodoStatus,
    SettlementStatus,
)


class BerthBase(BaseModel):
    berth_number: str
    capacity: int


class BerthCreate(BerthBase):
    pass


class Berth(BerthBase):
    id: int
    status: BerthStatus

    class Config:
        from_attributes = True


class StageHistoryBase(BaseModel):
    stage: AgentServiceStage
    status: StageStatus
    notes: Optional[str] = None


class StageHistory(StageHistoryBase):
    id: int
    agent_service_id: int
    completed_at: Optional[datetime] = None
    created_at: datetime

    class Config:
        from_attributes = True


class AgentServiceBase(BaseModel):
    ship_name: str
    imo_number: str
    captain_name: str
    arrival_time: datetime


class AgentServiceCreate(AgentServiceBase):
    pass


class AgentServiceUpdate(BaseModel):
    ship_name: Optional[str] = None
    imo_number: Optional[str] = None
    captain_name: Optional[str] = None
    arrival_time: Optional[datetime] = None


class AgentService(AgentServiceBase):
    id: int
    current_stage: AgentServiceStage
    departure_time: Optional[datetime] = None
    created_at: datetime
    updated_at: datetime
    stage_history: List[StageHistory] = []

    class Config:
        from_attributes = True


class BerthApplicationBase(BaseModel):
    requested_berthing_time: datetime
    expected_duration_hours: int = Field(..., ge=2, le=72)
    berth_preference: Optional[str] = None


class BerthApplicationCreate(BerthApplicationBase):
    pass


class BerthApplicationApprove(BaseModel):
    approved: bool
    notes: Optional[str] = None


class BerthApplication(BerthApplicationBase):
    id: int
    agent_service_id: int
    status: BerthApplicationStatus
    assigned_berth_id: Optional[int] = None
    assigned_berth: Optional[Berth] = None
    assigned_berth_time: Optional[datetime] = None
    is_waiting_timeout: bool
    approved_at: Optional[datetime] = None
    created_at: datetime

    class Config:
        from_attributes = True


class SupplyBase(BaseModel):
    supply_type: SupplyType
    quantity: float = Field(..., gt=0)
    unit_price: float = Field(..., gt=0)
    scheduled_time: datetime
    notes: Optional[str] = None


class SupplyCreate(SupplyBase):
    pass


class Supply(SupplyBase):
    id: int
    agent_service_id: int
    total_price: float
    status: SupplyStatus
    delivered_at: Optional[datetime] = None
    created_at: datetime

    class Config:
        from_attributes = True


class WasteRecoveryBase(BaseModel):
    total_weight_kg: float = Field(..., gt=0)
    unit_price_per_kg: float = Field(..., gt=0)
    scheduled_time: datetime
    notes: Optional[str] = None


class WasteRecoveryCreate(WasteRecoveryBase):
    pass


class WasteRecovery(WasteRecoveryBase):
    id: int
    agent_service_id: int
    discounted_weight_kg: Optional[float] = None
    discounted_unit_price: Optional[float] = None
    total_fee: float
    status: WasteStatus
    completed_at: Optional[datetime] = None
    created_at: datetime

    class Config:
        from_attributes = True


class TodoBase(BaseModel):
    title: str
    description: Optional[str] = None
    due_time: datetime


class TodoCreate(TodoBase):
    pass


class Todo(TodoBase):
    id: int
    agent_service_id: int
    status: TodoStatus
    completed_at: Optional[datetime] = None
    sequence: int
    created_at: datetime

    class Config:
        from_attributes = True


class FeeSettlementBase(BaseModel):
    agent_fee: float = 0
    other_fees: float = 0
    notes: Optional[str] = None


class FeeSettlementCreate(FeeSettlementBase):
    pass


class FeeSettlement(FeeSettlementBase):
    id: int
    agent_service_id: int
    supplies_fee: float
    waste_recovery_fee: float
    total_amount: float
    final_amount: float
    status: SettlementStatus
    settled_at: Optional[datetime] = None
    created_at: datetime

    class Config:
        from_attributes = True


class AgentServiceDetail(AgentService):
    berth_application: Optional[BerthApplication] = None
    supplies: List[Supply] = []
    waste_recovery: Optional[WasteRecovery] = None
    todos: List[Todo] = []
    fee_settlement: Optional[FeeSettlement] = None


class Message(BaseModel):
    message: str
