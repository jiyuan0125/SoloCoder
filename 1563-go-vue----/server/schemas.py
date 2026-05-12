from typing import Optional, List
from datetime import datetime
from pydantic import BaseModel, Field


class PartBase(BaseModel):
    part_number: str = Field(..., description="件号")
    serial_number: str = Field(..., description="件序号")
    name: str = Field(..., description="名称")
    unit_price: float = Field(..., ge=0, description="单价")
    is_controllable: bool = Field(default=False, description="是否可控件")
    minimum_stock: int = Field(default=0, ge=0, description="最低库存量")
    available_quantity: int = Field(default=0, ge=0, description="可用库存")


class PartCreate(PartBase):
    pass


class PartUpdate(BaseModel):
    name: Optional[str] = None
    unit_price: Optional[float] = None
    is_controllable: Optional[bool] = None
    minimum_stock: Optional[int] = None


class PartResponse(PartBase):
    id: int
    status: str
    stock_status: str
    is_out_of_stock: bool
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class RequisitionBase(BaseModel):
    requester: str = Field(..., description="申请人")
    quantity: int = Field(..., gt=0, description="领用数量")
    reason: Optional[str] = None


class RequisitionCreate(RequisitionBase):
    part_id: int


class RequisitionUpdate(BaseModel):
    approver: Optional[str] = None


class RequisitionResponse(BaseModel):
    id: int
    part_id: int
    requester: str
    quantity: int
    reason: Optional[str]
    status: str
    level1_approver: Optional[str]
    level1_approved_at: Optional[datetime]
    level2_approver: Optional[str]
    level2_approved_at: Optional[datetime]
    needs_level2_approval: bool
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class RepairBase(BaseModel):
    quantity: int = Field(..., gt=0, description="送修数量")
    repair_vendor: Optional[str] = None
    repair_cost: float = Field(default=0.0, ge=0, description="修复费用")


class RepairCreate(RepairBase):
    part_id: int


class RepairUpdate(BaseModel):
    repair_vendor: Optional[str] = None
    repair_cost: Optional[float] = None
    status: Optional[str] = None


class RepairResponse(BaseModel):
    id: int
    part_id: int
    quantity: int
    repair_vendor: Optional[str]
    repair_cost: float
    status: str
    is_overdue: bool
    should_scrap: bool
    last_updated_at: datetime
    created_at: datetime
    completed_at: Optional[datetime]

    class Config:
        from_attributes = True


class PurchaseBase(BaseModel):
    quantity: int = Field(..., gt=0, description="采购数量")
    unit_price: float = Field(..., ge=0, description="采购单价")
    reason: Optional[str] = None


class PurchaseCreate(PurchaseBase):
    part_id: int


class PurchaseUpdate(BaseModel):
    approver: Optional[str] = None
    status: Optional[str] = None


class PurchaseResponse(BaseModel):
    id: int
    part_id: int
    quantity: int
    unit_price: float
    urgency: str
    status: str
    approver: Optional[str]
    approved_at: Optional[datetime]
    order_date: Optional[datetime]
    receive_date: Optional[datetime]
    reason: Optional[str]
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class PurchaseHistoryResponse(BaseModel):
    id: int
    part_id: int
    purchase_id: Optional[int]
    quantity: int
    unit_price: float
    total_price: float
    received_at: datetime

    class Config:
        from_attributes = True


class InventoryCountBase(BaseModel):
    counted_quantity: int = Field(..., description="盘点数量")
    reason: Optional[str] = None
    adjuster: str = Field(..., description="调整人")


class InventoryCountCreate(InventoryCountBase):
    part_id: int


class InventoryCountResponse(BaseModel):
    id: int
    part_id: int
    counted_quantity: int
    system_quantity: int
    difference: int
    reason: Optional[str]
    adjuster: str
    adjusted_at: datetime
    created_at: datetime

    class Config:
        from_attributes = True


class PartDetailResponse(PartResponse):
    requisitions: List[RequisitionResponse] = []
    repairs: List[RepairResponse] = []
    purchases: List[PurchaseResponse] = []
    purchase_histories: List[PurchaseHistoryResponse] = []
