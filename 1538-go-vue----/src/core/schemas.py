from datetime import datetime, date
from typing import Optional, List
from pydantic import BaseModel, Field

from src.core.enums import WasteType, WaybillStatus, UnitType


class UnitBase(BaseModel):
    name: str
    unit_type: UnitType
    address: Optional[str] = None
    contact_person: Optional[str] = None
    contact_phone: Optional[str] = None


class UnitCreate(UnitBase):
    pass


class UnitUpdate(BaseModel):
    name: Optional[str] = None
    address: Optional[str] = None
    contact_person: Optional[str] = None
    contact_phone: Optional[str] = None


class UnitResponse(UnitBase):
    id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class QualificationBase(BaseModel):
    disposer_id: int
    hw_codes: str
    valid_from: date
    valid_until: date
    is_active: bool = True


class QualificationCreate(QualificationBase):
    pass


class QualificationUpdate(BaseModel):
    hw_codes: Optional[str] = None
    valid_from: Optional[date] = None
    valid_until: Optional[date] = None
    is_active: Optional[bool] = None


class QualificationResponse(QualificationBase):
    id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class WasteLedgerBase(BaseModel):
    producer_id: int
    waste_type: WasteType
    hw_code: Optional[str] = None
    waste_name: str
    quantity: float = Field(gt=0)
    unit: str = "吨"
    production_date: date
    remarks: Optional[str] = None


class WasteLedgerCreate(WasteLedgerBase):
    pass


class WasteLedgerUpdate(BaseModel):
    waste_name: Optional[str] = None
    quantity: Optional[float] = Field(default=None, gt=0)
    unit: Optional[str] = None
    remarks: Optional[str] = None


class WasteLedgerResponse(WasteLedgerBase):
    id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class WaybillBase(BaseModel):
    producer_id: int
    disposer_id: int
    waste_type: WasteType
    hw_code: Optional[str] = None
    waste_name: str
    quantity: float = Field(gt=0)
    unit: str = "吨"


class WaybillCreate(WaybillBase):
    pass


class WaybillStatusUpdate(BaseModel):
    new_status: WaybillStatus
    reject_reason: Optional[str] = None


class WaybillResponse(WaybillBase):
    id: int
    waybill_code: str
    status: WaybillStatus
    reject_reason: Optional[str] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class MonthlyLedgerBase(BaseModel):
    unit_id: int
    year: int
    month: int
    closing_balance: Optional[float] = None


class MonthlyLedgerCalculate(BaseModel):
    unit_id: int
    year: int
    month: int


class MonthlyLedgerResponse(BaseModel):
    id: int
    unit_id: int
    year: int
    month: int
    opening_balance: float
    production: float
    transfer_out: float
    closing_balance: float
    calculated_closing: float
    difference: float
    is_approved: bool
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class AlertResponse(BaseModel):
    id: int
    unit_id: int
    alert_type: str
    message: str
    is_resolved: bool
    created_at: datetime

    class Config:
        from_attributes = True


class CSVExportResponse(BaseModel):
    content: str
    filename: str
