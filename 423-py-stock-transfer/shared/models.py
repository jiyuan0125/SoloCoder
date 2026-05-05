from datetime import date, datetime
from decimal import Decimal
from typing import Optional
from uuid import UUID

from pydantic import BaseModel, Field, computed_field, field_validator

from shared.enums import (
    ApprovalLevel,
    ApprovalStatus,
    TransactionType,
    TransferStatus,
    TransferType,
)


class Warehouse(BaseModel):
    warehouse_id: UUID
    name: str
    location: str
    company_id: UUID
    created_at: datetime = Field(default_factory=datetime.now)
    updated_at: datetime = Field(default_factory=datetime.now)


class Product(BaseModel):
    product_id: UUID
    sku: str
    name: str
    unit_price: Decimal = Field(decimal_places=2, gt=Decimal("0"))
    unit: str = "件"
    created_at: datetime = Field(default_factory=datetime.now)


class Inventory(BaseModel):
    inventory_id: UUID
    warehouse_id: UUID
    product_id: UUID
    available_quantity: int = Field(ge=0, default=0)
    frozen_quantity: int = Field(ge=0, default=0)
    last_updated: datetime = Field(default_factory=datetime.now)

    @property
    def total_quantity(self) -> int:
        return self.available_quantity + self.frozen_quantity


class TransferItem(BaseModel):
    item_id: UUID
    product_id: UUID
    requested_quantity: int = Field(gt=0)
    shipped_quantity: Optional[int] = Field(ge=0, default=None)
    arrived_quantity: Optional[int] = Field(ge=0, default=None)
    loss_quantity: int = Field(ge=0, default=0)
    unit_price: Decimal = Field(decimal_places=2, gt=Decimal("0"))

    @field_validator("shipped_quantity", "arrived_quantity")
    @classmethod
    def check_quantity_ge_zero(cls, v: Optional[int]) -> Optional[int]:
        if v is not None and v < 0:
            raise ValueError("Quantity cannot be negative")
        return v

    @computed_field
    @property
    def requested_amount(self) -> Decimal:
        return self.requested_quantity * self.unit_price

    @computed_field
    @property
    def shipped_amount(self) -> Decimal:
        if self.shipped_quantity is None:
            return Decimal("0")
        return self.shipped_quantity * self.unit_price

    @computed_field
    @property
    def arrived_amount(self) -> Decimal:
        if self.arrived_quantity is None:
            return Decimal("0")
        return self.arrived_quantity * self.unit_price

    @computed_field
    @property
    def loss_amount(self) -> Decimal:
        return self.loss_quantity * self.unit_price


class ApprovalRecord(BaseModel):
    approval_id: UUID
    transfer_id: UUID
    approval_level: ApprovalLevel
    approval_status: ApprovalStatus
    approver_id: Optional[UUID] = None
    approver_name: Optional[str] = None
    comment: Optional[str] = None
    approved_at: Optional[datetime] = None
    created_at: datetime = Field(default_factory=datetime.now)


class TransferOrder(BaseModel):
    transfer_id: UUID
    transfer_no: str
    source_warehouse_id: UUID
    target_warehouse_id: UUID
    status: TransferStatus = TransferStatus.PENDING_CONFIRM
    items: list[TransferItem] = Field(default_factory=list)
    transfer_type: TransferType = TransferType.INTRA_COMPANY
    approval_records: list[ApprovalRecord] = Field(default_factory=list)
    required_approval_level: ApprovalLevel = ApprovalLevel.NONE
    inter_company_approved: bool = False
    created_by: UUID
    created_at: datetime = Field(default_factory=datetime.now)
    confirmed_at: Optional[datetime] = None
    shipped_at: Optional[datetime] = None
    arrived_at: Optional[datetime] = None
    stocked_at: Optional[datetime] = None
    cancelled_at: Optional[datetime] = None
    cancelled_by: Optional[UUID] = None
    remark: Optional[str] = None

    @computed_field
    @property
    def total_requested_amount(self) -> Decimal:
        total = Decimal("0")
        for item in self.items:
            total += item.requested_amount
        return total

    @computed_field
    @property
    def total_shipped_amount(self) -> Decimal:
        total = Decimal("0")
        for item in self.items:
            total += item.shipped_amount
        return total

    @computed_field
    @property
    def total_arrived_amount(self) -> Decimal:
        total = Decimal("0")
        for item in self.items:
            total += item.arrived_amount
        return total

    @computed_field
    @property
    def total_loss_amount(self) -> Decimal:
        total = Decimal("0")
        for item in self.items:
            total += item.loss_amount
        return total

    @computed_field
    @property
    def is_approved(self) -> bool:
        if self.required_approval_level == ApprovalLevel.NONE:
            return True
        for record in self.approval_records:
            if record.approval_level == self.required_approval_level:
                return record.approval_status == ApprovalStatus.APPROVED
        return False


class LossRecord(BaseModel):
    loss_id: UUID
    transfer_id: UUID
    item_id: UUID
    product_id: UUID
    warehouse_id: UUID
    loss_quantity: int = Field(gt=0)
    unit_price: Decimal = Field(decimal_places=2, gt=Decimal("0"))
    loss_amount: Decimal = Field(decimal_places=2, gt=Decimal("0"))
    loss_date: date
    created_at: datetime = Field(default_factory=datetime.now)
    remark: Optional[str] = None


class InventoryTransaction(BaseModel):
    transaction_id: UUID
    warehouse_id: UUID
    product_id: UUID
    transaction_type: TransactionType
    quantity: int
    unit_price: Decimal = Field(decimal_places=2)
    amount: Decimal = Field(decimal_places=2)
    reference_type: str = "transfer"
    reference_id: UUID
    transaction_time: datetime = Field(default_factory=datetime.now)
    remark: Optional[str] = None


class LossMonthlySummary(BaseModel):
    year: int
    month: int
    warehouse_id: UUID
    total_transfer_amount: Decimal = Field(decimal_places=2, default=Decimal("0"))
    total_loss_amount: Decimal = Field(decimal_places=2, default=Decimal("0"))
    total_transfer_quantity: int = 0
    total_loss_quantity: int = 0

    @computed_field
    @property
    def loss_rate(self) -> Decimal:
        if self.total_transfer_quantity == 0:
            return Decimal("0")
        rate = (self.total_loss_quantity / self.total_transfer_quantity) * 100
        return Decimal(str(round(rate, 4)))
