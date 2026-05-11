from datetime import date, datetime
from typing import Optional, List, Dict, Any
from pydantic import BaseModel, Field


class EnterpriseBase(BaseModel):
    name: str
    address: Optional[str] = None
    contact_person: Optional[str] = None
    contact_phone: Optional[str] = None
    license_number: Optional[str] = None


class EnterpriseCreate(EnterpriseBase):
    pass


class EnterpriseUpdate(BaseModel):
    name: Optional[str] = None
    address: Optional[str] = None
    contact_person: Optional[str] = None
    contact_phone: Optional[str] = None
    license_number: Optional[str] = None


class EnterpriseResponse(EnterpriseBase):
    id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class ChemicalBase(BaseModel):
    name: str
    category: Optional[str] = None
    cas_number: Optional[str] = None
    hazard_level: Optional[str] = None
    description: Optional[str] = None


class ChemicalCreate(ChemicalBase):
    pass


class ChemicalUpdate(BaseModel):
    name: Optional[str] = None
    category: Optional[str] = None
    cas_number: Optional[str] = None
    hazard_level: Optional[str] = None
    description: Optional[str] = None


class ChemicalResponse(ChemicalBase):
    id: int
    enterprise_id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class ApprovalBase(BaseModel):
    approval_level: str
    max_quantity: float


class ApprovalCreate(ApprovalBase):
    enterprise_id: int
    chemical_id: int


class ApprovalResponse(BaseModel):
    id: int
    enterprise_id: int
    chemical_id: int
    approval_level: str
    status: str
    max_quantity: float
    issue_date: Optional[date] = None
    expiry_date: Optional[date] = None
    level1_approver: Optional[str] = None
    level1_approved_at: Optional[datetime] = None
    level2_approver: Optional[str] = None
    level2_approved_at: Optional[datetime] = None
    level3_approver: Optional[str] = None
    level3_approved_at: Optional[datetime] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class ApprovalAction(BaseModel):
    level: int
    comment: Optional[str] = None
    user: Optional[str] = None


class EmergencyPlanBase(BaseModel):
    title: str
    version: Optional[str] = None
    content: str


class EmergencyPlanCreate(EmergencyPlanBase):
    enterprise_id: int


class EmergencyPlanUpdate(BaseModel):
    title: Optional[str] = None
    version: Optional[str] = None
    content: Optional[str] = None


class EmergencyPlanResponse(EmergencyPlanBase):
    id: int
    enterprise_id: int
    created_by: Optional[str] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class AccidentRecordBase(BaseModel):
    title: str
    accident_date: datetime
    location: Optional[str] = None
    description: str
    severity: Optional[str] = None
    casualties: int = 0
    financial_loss: float = 0.0
    status: str = "processing"


class AccidentRecordCreate(AccidentRecordBase):
    enterprise_id: int


class AccidentRecordUpdate(BaseModel):
    title: Optional[str] = None
    accident_date: Optional[datetime] = None
    location: Optional[str] = None
    description: Optional[str] = None
    severity: Optional[str] = None
    casualties: Optional[int] = None
    financial_loss: Optional[float] = None
    status: Optional[str] = None


class DispositionAdd(BaseModel):
    disposition: str
    user: Optional[str] = None


class AccidentRecordResponse(AccidentRecordBase):
    id: int
    enterprise_id: int
    disposition: Optional[str] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class LedgerTransaction(BaseModel):
    enterprise_id: int
    chemical_id: int
    transaction_type: str
    quantity: float
    unit: str = "kg"
    transaction_date: Optional[date] = None
    description: Optional[str] = None


class LedgerResponse(BaseModel):
    id: int
    enterprise_id: int
    chemical_id: int
    transaction_type: str
    quantity: float
    balance: float
    unit: str
    transaction_date: date
    operator: Optional[str] = None
    description: Optional[str] = None
    created_at: datetime

    class Config:
        from_attributes = True


class DrillBase(BaseModel):
    title: str
    drill_date: date
    description: Optional[str] = None
    participants: Optional[int] = None
    duration_hours: Optional[float] = None
    evaluation_result: Optional[str] = None
    evaluation_details: Optional[str] = None
    plan_id: Optional[int] = None


class DrillCreate(DrillBase):
    enterprise_id: int


class DrillResponse(DrillBase):
    id: int
    enterprise_id: int
    created_at: datetime

    class Config:
        from_attributes = True


class AuditLogResponse(BaseModel):
    id: int
    action: str
    entity_type: Optional[str] = None
    entity_id: Optional[int] = None
    description: str
    user: Optional[str] = None
    timestamp: datetime
    details: Optional[str] = None

    class Config:
        from_attributes = True


class ReminderResponse(BaseModel):
    id: int
    reminder_type: str
    target_id: int
    target_type: str
    message: str
    due_date: Optional[date] = None
    is_completed: bool
    created_at: datetime

    class Config:
        from_attributes = True


class LedgerCheckResponse(BaseModel):
    id: int
    check_date: date
    enterprise_id: Optional[int] = None
    chemical_id: Optional[int] = None
    is_abnormal: bool
    anomaly_type: Optional[str] = None
    anomaly_description: Optional[str] = None
    created_at: datetime

    class Config:
        from_attributes = True


class CheckResult(BaseModel):
    message: str
    abnormal_count: int
    check_date: date
