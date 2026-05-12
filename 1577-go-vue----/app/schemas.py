from pydantic import BaseModel, Field
from typing import Optional
from datetime import datetime
from app.models import MaintenanceType, MaintenanceStatus


class LocomotiveBase(BaseModel):
    locomotive_number: str = Field(..., max_length=50)
    section_cycle_km: float = Field(default=300000.0, gt=0)


class LocomotiveCreate(LocomotiveBase):
    pass


class LocomotiveUpdateKm(BaseModel):
    current_km: float = Field(..., ge=0)


class LocomotiveResponse(LocomotiveBase):
    id: int
    current_km: float
    last_section_maintenance_km: float
    last_factory_maintenance_km: float
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class LocomotiveWithNextMaintenance(BaseModel):
    locomotive: LocomotiveResponse
    next_section_km: Optional[float] = None
    next_factory_km: float
    is_section_overdue: bool = False
    is_factory_overdue: bool = False
    overdue_km: float = 0.0


class MaintenancePlanBase(BaseModel):
    locomotive_id: int
    maintenance_type: MaintenanceType


class MaintenancePlanCreate(MaintenancePlanBase):
    pass


class MaintenancePlanUpdate(BaseModel):
    planned_km: Optional[float] = Field(default=None, gt=0)


class MaintenancePlanStatusUpdate(BaseModel):
    status: MaintenanceStatus


class MaintenanceContentUpdate(BaseModel):
    maintenance_content: str


class PartReplacement(BaseModel):
    part_code: str
    quantity: int = Field(..., gt=0)


class ReplacementPartResponse(BaseModel):
    id: int
    part_code: str
    part_name: str
    quantity: int
    created_at: datetime

    class Config:
        from_attributes = True


class MaintenancePlanResponse(BaseModel):
    id: int
    plan_number: str
    locomotive_id: int
    locomotive_number: str
    maintenance_type: MaintenanceType
    planned_km: float
    status: MaintenanceStatus
    is_cancelled: bool
    maintenance_content: Optional[str] = None
    replacement_parts: list[ReplacementPartResponse] = []
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class PartBase(BaseModel):
    part_code: str = Field(..., max_length=50)
    part_name: str = Field(..., max_length=100)
    stock: int = Field(default=0, ge=0)
    warning_threshold: int = Field(default=10, ge=0)
    unit: str = Field(default="个", max_length=20)


class PartCreate(PartBase):
    pass


class PartUpdate(BaseModel):
    part_name: Optional[str] = None
    stock: Optional[int] = Field(default=None, ge=0)
    warning_threshold: Optional[int] = Field(default=None, ge=0)
    unit: Optional[str] = None


class PartResponse(PartBase):
    id: int
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class PurchaseAlertResponse(BaseModel):
    id: int
    part_code: str
    part_name: str
    current_stock: int
    threshold: int
    is_active: bool
    created_at: datetime

    class Config:
        from_attributes = True


class TechnicalManualBase(BaseModel):
    manual_code: str = Field(..., max_length=50)
    title: str = Field(..., max_length=200)
    author: Optional[str] = Field(default=None, max_length=100)
    version: Optional[str] = Field(default=None, max_length=50)


class TechnicalManualCreate(TechnicalManualBase):
    pass


class TechnicalManualResponse(TechnicalManualBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class BorrowRecordBase(BaseModel):
    manual_id: int
    borrower: str = Field(..., max_length=100)


class BorrowRecordCreate(BorrowRecordBase):
    pass


class BorrowRecordResponse(BaseModel):
    id: int
    manual_id: int
    manual_code: str
    manual_title: str
    borrower: str
    borrow_date: datetime
    due_date: datetime
    return_date: Optional[datetime] = None
    is_overdue: bool
    created_at: datetime

    class Config:
        from_attributes = True


class MessageResponse(BaseModel):
    message: str
