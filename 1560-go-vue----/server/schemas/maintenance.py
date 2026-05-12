from datetime import datetime
from typing import List, Optional

from pydantic import BaseModel


class AlarmBase(BaseModel):
    lighthouse_id: int
    alarm_type: str
    severity: str
    message: str
    status: Optional[str] = "open"


class AlarmCreate(AlarmBase):
    pass


class AlarmResponse(AlarmBase):
    id: int
    created_at: datetime
    resolved_at: Optional[datetime] = None
    resolved_by: Optional[str] = None

    class Config:
        from_attributes = True


class WorkOrderBase(BaseModel):
    lighthouse_id: int
    order_type: str
    priority: str
    title: str
    description: Optional[str] = None


class WorkOrderCreate(WorkOrderBase):
    pass


class WorkOrderAssign(BaseModel):
    assigned_to: str


class WorkOrderExecute(BaseModel):
    executed_by: str
    execution_notes: Optional[str] = None


class WorkOrderAccept(BaseModel):
    accepted_by: str
    acceptance_result: str


class WorkOrderResponse(WorkOrderBase):
    id: int
    status: str
    assigned_to: Optional[str] = None
    assigned_at: Optional[datetime] = None
    executed_at: Optional[datetime] = None
    executed_by: Optional[str] = None
    execution_notes: Optional[str] = None
    accepted_at: Optional[datetime] = None
    accepted_by: Optional[str] = None
    acceptance_result: Optional[str] = None
    created_at: datetime

    class Config:
        from_attributes = True


class MaintenancePlanBase(BaseModel):
    lighthouse_id: int
    plan_type: str
    frequency_days: int
    title: str
    description: Optional[str] = None
    is_active: Optional[int] = 1


class MaintenancePlanCreate(MaintenancePlanBase):
    pass


class MaintenancePlanUpdate(BaseModel):
    plan_type: Optional[str] = None
    frequency_days: Optional[int] = None
    title: Optional[str] = None
    description: Optional[str] = None
    is_active: Optional[int] = None


class MaintenancePlanResponse(MaintenancePlanBase):
    id: int
    last_execution_date: Optional[datetime] = None
    next_execution_date: Optional[datetime] = None
    created_at: datetime

    class Config:
        from_attributes = True
