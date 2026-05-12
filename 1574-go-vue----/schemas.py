from datetime import datetime
from typing import Optional, List
from pydantic import BaseModel, Field
from models import InspectionType, DefectLevel, TaskStatus, TeamStatus


class LineSectionBase(BaseModel):
    name: str
    start_km: float
    end_km: float
    description: Optional[str] = None


class LineSectionCreate(LineSectionBase):
    pass


class LineSectionUpdate(BaseModel):
    name: Optional[str] = None
    start_km: Optional[float] = None
    end_km: Optional[float] = None
    description: Optional[str] = None


class LineSectionResponse(LineSectionBase):
    id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class MaintenanceTeamBase(BaseModel):
    name: str
    leader: str
    phone: Optional[str] = None


class MaintenanceTeamCreate(MaintenanceTeamBase):
    pass


class MaintenanceTeamUpdate(BaseModel):
    name: Optional[str] = None
    leader: Optional[str] = None
    phone: Optional[str] = None


class MaintenanceTeamResponse(MaintenanceTeamBase):
    id: int
    status: str
    current_task_id: Optional[int] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class InspectionBase(BaseModel):
    section_id: int
    inspection_type: InspectionType
    inspection_date: Optional[datetime] = None
    inspector: Optional[str] = None
    notes: Optional[str] = None


class InspectionCreate(InspectionBase):
    pass


class InspectionResponse(BaseModel):
    id: int
    section_id: int
    team_id: int
    inspection_type: str
    inspection_date: datetime
    inspector: Optional[str] = None
    notes: Optional[str] = None
    created_at: datetime

    class Config:
        from_attributes = True


class DefectBase(BaseModel):
    section_id: int
    defect_type: str
    location_km: float
    level: DefectLevel
    description: Optional[str] = None


class DefectCreate(DefectBase):
    pass


class DefectUpdate(BaseModel):
    defect_type: Optional[str] = None
    location_km: Optional[float] = None
    level: Optional[DefectLevel] = None
    description: Optional[str] = None
    is_resolved: Optional[bool] = None


class DefectResponse(BaseModel):
    id: int
    inspection_id: int
    section_id: int
    defect_type: str
    location_km: float
    level: str
    description: Optional[str] = None
    is_resolved: bool
    resolved_at: Optional[datetime] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class TaskBase(BaseModel):
    task_type: str
    notes: Optional[str] = None


class TaskCreate(TaskBase):
    defect_id: int


class TaskUpdate(BaseModel):
    status: Optional[TaskStatus] = None
    notes: Optional[str] = None


class TaskResponse(BaseModel):
    id: int
    defect_id: int
    team_id: Optional[int] = None
    task_type: str
    status: str
    priority: int
    due_date: datetime
    started_at: Optional[datetime] = None
    completed_at: Optional[datetime] = None
    notes: Optional[str] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class RailBase(BaseModel):
    section_id: int
    rail_number: str
    start_km: float
    end_km: float
    max_total_weight: float
    current_total_weight: Optional[float] = 0
    wear_mm: Optional[float] = 0


class RailCreate(RailBase):
    pass


class RailUpdate(BaseModel):
    rail_number: Optional[str] = None
    current_total_weight: Optional[float] = None
    wear_mm: Optional[float] = None
    is_replaced: Optional[bool] = None


class RailResponse(BaseModel):
    id: int
    section_id: int
    rail_number: str
    start_km: float
    end_km: float
    max_total_weight: float
    current_total_weight: float
    wear_mm: float
    installation_date: datetime
    is_replaced: bool
    replaced_at: Optional[datetime] = None
    created_at: datetime
    updated_at: datetime

    @property
    def remaining_life_percent(self) -> float:
        if self.max_total_weight <= 0:
            return 0
        remaining = max(0, self.max_total_weight - self.current_total_weight)
        return (remaining / self.max_total_weight) * 100

    class Config:
        from_attributes = True


class RailMaintenanceBase(BaseModel):
    maintenance_type: str
    before_wear_mm: Optional[float] = None
    after_wear_mm: Optional[float] = None
    notes: Optional[str] = None


class RailMaintenanceCreate(RailMaintenanceBase):
    rail_id: int


class RailMaintenanceResponse(BaseModel):
    id: int
    rail_id: int
    maintenance_type: str
    maintenance_date: datetime
    before_wear_mm: Optional[float] = None
    after_wear_mm: Optional[float] = None
    notes: Optional[str] = None
    created_at: datetime

    class Config:
        from_attributes = True


class RailReplacementBase(BaseModel):
    old_rail_number: str
    new_rail_number: str
    replacement_reason: str
    notes: Optional[str] = None


class RailReplacementCreate(RailReplacementBase):
    rail_id: int


class RailReplacementResponse(BaseModel):
    id: int
    rail_id: int
    old_rail_number: str
    new_rail_number: str
    replacement_reason: str
    replacement_date: datetime
    notes: Optional[str] = None
    created_at: datetime

    class Config:
        from_attributes = True


class WarningResponse(BaseModel):
    id: int
    warning_type: str
    target_type: str
    target_id: int
    message: str
    is_acknowledged: bool
    created_at: datetime

    class Config:
        from_attributes = True
