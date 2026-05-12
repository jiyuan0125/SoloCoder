from datetime import date, datetime
from typing import Optional, List
from pydantic import BaseModel, Field


class VesselBase(BaseModel):
    name: str
    imo_number: Optional[str] = None
    flag: Optional[str] = None
    gross_tonnage: Optional[int] = None
    built_year: Optional[int] = None
    vessel_type: Optional[str] = None


class VesselCreate(VesselBase):
    pass


class VesselUpdate(VesselBase):
    pass


class Vessel(VesselBase):
    id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class InspectionBase(BaseModel):
    inspection_type: str
    inspection_date: date
    inspector: Optional[str] = None
    result: Optional[str] = None
    remarks: Optional[str] = None


class InspectionCreate(InspectionBase):
    vessel_id: int


class InspectionUpdate(BaseModel):
    inspector: Optional[str] = None
    result: Optional[str] = None
    remarks: Optional[str] = None


class Inspection(InspectionBase):
    id: int
    vessel_id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class CertificateBase(BaseModel):
    certificate_type: str
    certificate_number: Optional[str] = None
    issue_date: date
    expiry_date: date
    issued_by: Optional[str] = None
    status: str = "valid"
    remarks: Optional[str] = None


class CertificateCreate(CertificateBase):
    vessel_id: int
    inspection_id: Optional[int] = None


class CertificateUpdate(BaseModel):
    certificate_number: Optional[str] = None
    expiry_date: Optional[date] = None
    issued_by: Optional[str] = None
    status: Optional[str] = None
    remarks: Optional[str] = None


class Certificate(CertificateBase):
    id: int
    vessel_id: int
    inspection_id: Optional[int] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class DockBase(BaseModel):
    name: str
    location: Optional[str] = None
    capacity: Optional[str] = None


class DockCreate(DockBase):
    pass


class Dock(DockBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class DockingBase(BaseModel):
    vessel_id: int
    dock_id: int
    start_date: date
    end_date: date
    purpose: Optional[str] = None
    status: str = "scheduled"


class DockingCreate(DockingBase):
    pass


class DockingUpdate(BaseModel):
    start_date: Optional[date] = None
    end_date: Optional[date] = None
    purpose: Optional[str] = None
    status: Optional[str] = None


class Docking(DockingBase):
    id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class TodoItemBase(BaseModel):
    vessel_id: int
    todo_type: str
    title: str
    inspection_type: Optional[str] = None
    due_date: date
    arrange_deadline: Optional[date] = None
    status: str = "pending"
    reminder_level: Optional[int] = None
    related_certificate_id: Optional[int] = None
    remarks: Optional[str] = None


class TodoItemCreate(TodoItemBase):
    pass


class TodoItemUpdate(BaseModel):
    status: Optional[str] = None
    remarks: Optional[str] = None


class TodoItem(TodoItemBase):
    id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class ReminderBase(BaseModel):
    vessel_id: int
    certificate_id: int
    reminder_type: str
    days_before_expiry: int
    message: Optional[str] = None
    sent_date: Optional[date] = None
    is_read: bool = False


class ReminderCreate(ReminderBase):
    pass


class Reminder(ReminderBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class OperationRecordBase(BaseModel):
    vessel_id: int
    start_date: date
    end_date: Optional[date] = None
    route: Optional[str] = None
    status: str = "active"
    is_illegal: bool = False
    remarks: Optional[str] = None


class OperationRecordCreate(OperationRecordBase):
    pass


class OperationRecordUpdate(BaseModel):
    end_date: Optional[date] = None
    route: Optional[str] = None
    status: Optional[str] = None
    is_illegal: Optional[bool] = None
    remarks: Optional[str] = None


class OperationRecord(OperationRecordBase):
    id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True
