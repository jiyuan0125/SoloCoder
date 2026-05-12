from pydantic import BaseModel, validator
from datetime import datetime
from app.models import MaintenanceType, WorkOrderStatus, AlarmSeverity
from typing import Optional


class WorkOrderCreate(BaseModel):
    section: str
    maintenance_type: MaintenanceType
    plan_start_time: Optional[datetime] = None
    plan_end_time: Optional[datetime] = None
    person_in_charge: str
    description: Optional[str] = None

    @validator('plan_end_time', always=True)
    def check_time_range(cls, v, values):
        start = values.get('plan_start_time')
        if start and v:
            if v <= start:
                raise ValueError('计划结束时间必须晚于开始时间')
        return v


class WorkOrderUpdate(BaseModel):
    status: Optional[WorkOrderStatus] = None
    actual_start_time: Optional[datetime] = None
    actual_end_time: Optional[datetime] = None
    description: Optional[str] = None


class WorkOrderResponse(BaseModel):
    id: int
    section: str
    maintenance_type: MaintenanceType
    status: WorkOrderStatus
    plan_start_time: Optional[datetime] = None
    plan_end_time: Optional[datetime] = None
    actual_start_time: Optional[datetime] = None
    actual_end_time: Optional[datetime] = None
    person_in_charge: str
    description: Optional[str] = None
    created_at: datetime
    updated_at: datetime
    is_overdue: bool = False

    class Config:
        from_attributes = True


class PowerSupplyDataCreate(BaseModel):
    section: str
    current: float
    voltage: float
    rated_current: float = 1000.0


class PowerSupplyDataResponse(BaseModel):
    id: int
    section: str
    current: float
    voltage: float
    rated_current: float
    is_power_outage: int
    is_alarm: int
    recorded_at: datetime

    class Config:
        from_attributes = True


class SectionStatisticsResponse(BaseModel):
    id: int
    section: str
    total_inspection_count: int
    total_power_outage_duration_minutes: int
    total_alarm_count: int
    last_updated_at: datetime

    class Config:
        from_attributes = True


class AlarmWorkOrderResponse(BaseModel):
    id: int
    section: str
    severity: AlarmSeverity
    description: Optional[str] = None
    original_work_order_id: Optional[int] = None
    status: WorkOrderStatus
    created_at: datetime

    class Config:
        from_attributes = True
