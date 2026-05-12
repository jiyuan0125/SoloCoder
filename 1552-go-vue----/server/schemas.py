from datetime import datetime
from typing import Optional, List
from pydantic import BaseModel, Field, validator
from .models import (
    AgentServiceStatus,
    BerthingRequestStatus,
    MaterialType,
    DeliveryStatus,
    WasteCollectionStatus,
    TodoStatus
)


class BerthBase(BaseModel):
    name: str
    location: Optional[str] = None
    max_length: Optional[float] = None
    max_draft: Optional[float] = None
    is_available: bool = True
    description: Optional[str] = None


class BerthCreate(BerthBase):
    pass


class BerthUpdate(BaseModel):
    name: Optional[str] = None
    location: Optional[str] = None
    max_length: Optional[float] = None
    max_draft: Optional[float] = None
    is_available: Optional[bool] = None
    description: Optional[str] = None


class Berth(BerthBase):
    id: int

    class Config:
        from_attributes = True


class AgentServiceBase(BaseModel):
    ship_name: str
    imo_number: Optional[str] = None
    port_of_call: Optional[str] = None
    arrival_time: datetime
    estimated_departure_time: Optional[datetime] = None
    agency_fee: float = 0
    notes: Optional[str] = None


class AgentServiceCreate(AgentServiceBase):
    pass


class AgentServiceUpdate(BaseModel):
    ship_name: Optional[str] = None
    imo_number: Optional[str] = None
    port_of_call: Optional[str] = None
    arrival_time: Optional[datetime] = None
    estimated_departure_time: Optional[datetime] = None
    agency_fee: Optional[float] = None
    notes: Optional[str] = None


class AgentService(AgentServiceBase):
    id: int
    status: AgentServiceStatus
    total_fee: float = 0
    material_fee: float = 0
    waste_fee: float = 0
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class AgentServiceDetail(AgentService):
    berthing_request: Optional["BerthingRequest"] = None
    material_deliveries: List["MaterialDelivery"] = []
    waste_collections: List["WasteCollection"] = []
    todos: List["Todo"] = []
    settlements: List["Settlement"] = []


class BerthingRequestBase(BaseModel):
    agent_service_id: int
    requested_berthing_time: datetime
    estimated_duration_hours: float
    berth_preference: Optional[str] = None

    @validator("estimated_duration_hours")
    def check_duration(cls, v):
        if not (2 <= v <= 72):
            raise ValueError("预计靠泊时长必须在2到72小时之间")
        return v

    @validator("requested_berthing_time")
    def check_future_time(cls, v):
        if v <= datetime.utcnow():
            raise ValueError("申请靠泊时间必须是未来时间")
        return v


class BerthingRequestCreate(BerthingRequestBase):
    pass


class BerthingRequestApprove(BaseModel):
    approved_by: str
    auto_assign_berth: bool = True


class BerthingRequest(BerthingRequestBase):
    id: int
    status: BerthingRequestStatus
    assigned_berth_id: Optional[int] = None
    actual_berthing_time: Optional[datetime] = None
    approved_by: Optional[str] = None
    approval_time: Optional[datetime] = None
    is_timeout: bool = False
    created_at: datetime
    updated_at: datetime
    assigned_berth: Optional[Berth] = None

    class Config:
        from_attributes = True


class MaterialDeliveryBase(BaseModel):
    agent_service_id: int
    material_type: MaterialType
    quantity_tons: float
    notes: Optional[str] = None


class MaterialDeliveryCreate(MaterialDeliveryBase):
    pass


class MaterialDelivery(MaterialDeliveryBase):
    id: int
    unit_price: float
    total_price: float
    delivery_time: Optional[datetime] = None
    status: DeliveryStatus
    created_at: datetime

    class Config:
        from_attributes = True


class WasteCollectionBase(BaseModel):
    agent_service_id: int
    waste_type: Optional[str] = None
    weight_kg: float
    notes: Optional[str] = None


class WasteCollectionCreate(WasteCollectionBase):
    pass


class WasteCollection(WasteCollectionBase):
    id: int
    unit_price: float
    total_price: float
    collection_time: Optional[datetime] = None
    status: WasteCollectionStatus
    created_at: datetime

    class Config:
        from_attributes = True


class TodoBase(BaseModel):
    agent_service_id: int
    title: str
    description: Optional[str] = None
    due_time: datetime


class TodoCreate(TodoBase):
    pass


class TodoUpdate(BaseModel):
    title: Optional[str] = None
    description: Optional[str] = None
    due_time: Optional[datetime] = None


class Todo(TodoBase):
    id: int
    status: TodoStatus
    completed_time: Optional[datetime] = None
    created_at: datetime

    class Config:
        from_attributes = True


class SettlementBase(BaseModel):
    agent_service_id: int
    notes: Optional[str] = None


class SettlementCreate(SettlementBase):
    pass


class Settlement(SettlementBase):
    id: int
    agency_fee: float = 0
    material_fee: float = 0
    waste_fee: float = 0
    total_amount: float = 0
    settlement_time: Optional[datetime] = None
    created_at: datetime

    class Config:
        from_attributes = True


class StatusTransition(BaseModel):
    target_status: AgentServiceStatus


AgentServiceDetail.model_rebuild()
BerthingRequest.model_rebuild()
MaterialDelivery.model_rebuild()
WasteCollection.model_rebuild()
Todo.model_rebuild()
Settlement.model_rebuild()
