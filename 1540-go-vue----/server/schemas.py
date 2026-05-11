from datetime import datetime, date
from typing import Optional, List
from pydantic import BaseModel, Field, validator

from .models import SourceStatus, TodoStatus, ApprovalStatus, AlertLevel


class UnitBase(BaseModel):
    name: str = Field(..., description="涉源单位名称")
    address: str = Field(..., description="单位地址")
    contact_person: Optional[str] = Field(None, description="联系人")
    contact_phone: Optional[str] = Field(None, description="联系电话")


class UnitCreate(UnitBase):
    pass


class UnitUpdate(BaseModel):
    name: Optional[str] = None
    address: Optional[str] = None
    contact_person: Optional[str] = None
    contact_phone: Optional[str] = None


class Unit(UnitBase):
    id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class SourceBase(BaseModel):
    source_code: str = Field(..., min_length=12, max_length=12, description="12位源编码")
    name: str = Field(..., description="放射源名称")
    source_type: str = Field(..., description="放射源类型")
    source_class: int = Field(..., ge=1, le=5, description="放射源类别(1-5类)")
    activity: float = Field(..., gt=0, description="活度")
    unit_id: int = Field(..., description="所属单位ID")
    manufacture_date: date = Field(..., description="生产日期")
    expire_date: date = Field(..., description="到期日期")
    storage_location: Optional[str] = Field(None, description="存放位置")
    description: Optional[str] = Field(None, description="描述")


class SourceCreate(SourceBase):
    @validator("source_code")
    def validate_source_code(cls, v):
        if len(v) != 12:
            raise ValueError("源编码必须为12位")
        return v

    @validator("expire_date")
    def validate_expire_date(cls, v, values):
        if "manufacture_date" in values and v <= values["manufacture_date"]:
            raise ValueError("到期日期必须晚于生产日期")
        return v


class SourceUpdate(BaseModel):
    name: Optional[str] = None
    source_type: Optional[str] = None
    source_class: Optional[int] = None
    activity: Optional[float] = None
    unit_id: Optional[int] = None
    status: Optional[SourceStatus] = None
    manufacture_date: Optional[date] = None
    expire_date: Optional[date] = None
    storage_location: Optional[str] = None
    description: Optional[str] = None


class SourceStatusUpdate(BaseModel):
    status: SourceStatus = Field(..., description="目标状态")
    operator: str = Field(..., description="操作人")
    remarks: Optional[str] = None


class Source(SourceBase):
    id: int
    status: SourceStatus
    created_at: datetime
    updated_at: datetime
    unit: Optional[Unit] = None

    class Config:
        from_attributes = True


class SourceWithAlert(Source):
    alert_level: Optional[AlertLevel] = None
    days_to_expire: Optional[int] = None


class InspectionBase(BaseModel):
    inspector: str = Field(..., description="巡检人")
    radiation_dose: float = Field(..., description="辐射剂量")
    remarks: Optional[str] = None


class InspectionCreate(InspectionBase):
    source_id: int = Field(..., description="放射源ID")


class Inspection(InspectionBase):
    id: int
    source_id: int
    inspection_time: datetime
    is_abnormal: int
    result: str
    created_at: datetime

    class Config:
        from_attributes = True


class TodoBase(BaseModel):
    pass


class TodoUpdate(BaseModel):
    status: TodoStatus = Field(..., description="状态")
    handler: str = Field(..., description="处理人")
    handle_result: str = Field(..., description="处理结果")


class Todo(TodoBase):
    id: int
    inspection_id: int
    source_id: int
    source_code: str
    dose_value: float
    status: TodoStatus
    handler: Optional[str] = None
    handle_time: Optional[datetime] = None
    handle_result: Optional[str] = None
    created_at: datetime

    class Config:
        from_attributes = True


class ApprovalBase(BaseModel):
    plan_content: str = Field(..., description="退役方案内容")
    applicant: str = Field(..., description="申请人")


class ApprovalCreate(ApprovalBase):
    source_id: int = Field(..., description="放射源ID")


class ApprovalUpdate(BaseModel):
    plan_content: Optional[str] = None
    applicant: Optional[str] = None


class ApprovalSubmit(BaseModel):
    operator: str = Field(..., description="提交人")


class ApprovalApprove(BaseModel):
    operator: str = Field(..., description="审批人")
    approved: bool = Field(..., description="是否通过")
    remarks: Optional[str] = None


class Approval(ApprovalBase):
    id: int
    source_id: int
    source_code: str
    apply_time: datetime
    status: ApprovalStatus
    approver: Optional[str] = None
    approval_time: Optional[datetime] = None
    approval_remarks: Optional[str] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class AuditLogBase(BaseModel):
    operation_type: str = Field(..., description="操作类型")
    operator: str = Field(..., description="操作人")
    target_type: Optional[str] = None
    target_id: Optional[int] = None
    description: Optional[str] = None


class AuditLogCreate(AuditLogBase):
    pass


class AuditLog(AuditLogBase):
    id: int
    operation_time: datetime
    created_at: datetime

    class Config:
        from_attributes = True


class UnitDetail(Unit):
    sources: List[Source] = []
    inspections: List[Inspection] = []


class AlertInfo(BaseModel):
    source_id: int
    source_code: str
    source_name: str
    unit_name: str
    expire_date: date
    days_to_expire: int
    alert_level: AlertLevel


class PaginatedResponse(BaseModel):
    total: int
    items: List
    page: int
    size: int
