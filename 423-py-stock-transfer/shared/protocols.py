from decimal import Decimal
from typing import Any, Optional
from uuid import UUID

from pydantic import BaseModel, Field, field_validator

from shared.enums import (
    ApprovalLevel,
    ApprovalStatus,
    TransferStatus,
    TransferType,
)
from shared.models import (
    ApprovalRecord,
    Inventory,
    InventoryTransaction,
    LossMonthlySummary,
    LossRecord,
    TransferItem,
    TransferOrder,
    Warehouse,
)


class TransferItemCreate(BaseModel):
    product_id: UUID
    requested_quantity: int = Field(gt=0)


class TransferCreateRequest(BaseModel):
    source_warehouse_id: UUID
    target_warehouse_id: UUID
    items: list[TransferItemCreate] = Field(min_length=1)
    transfer_type: TransferType = TransferType.INTRA_COMPANY
    created_by: UUID
    remark: Optional[str] = None

    @field_validator("items")
    @classmethod
    def validate_items_unique(cls, v: list[TransferItemCreate]) -> list[TransferItemCreate]:
        product_ids = {item.product_id for item in v}
        if len(product_ids) != len(v):
            raise ValueError("Duplicate product_id in items")
        return v


class TransferUpdateStatusRequest(BaseModel):
    operator_id: UUID
    remark: Optional[str] = None


class TransferConfirmRequest(TransferUpdateStatusRequest):
    pass


class TransferShipRequest(TransferUpdateStatusRequest):
    pass


class TransferArrivalItem(BaseModel):
    item_id: UUID
    arrived_quantity: int = Field(ge=0)


class TransferArriveRequest(TransferUpdateStatusRequest):
    arrival_items: list[TransferArrivalItem] = Field(min_length=1)


class TransferStockRequest(TransferUpdateStatusRequest):
    pass


class TransferCancelRequest(BaseModel):
    operator_id: UUID
    reason: str


class TransferApproveRequest(BaseModel):
    approval_level: ApprovalLevel
    approval_status: ApprovalStatus
    approver_id: UUID
    approver_name: str
    comment: Optional[str] = None


class InterCompanyApproveRequest(BaseModel):
    operator_id: UUID
    operator_name: str
    approved: bool
    comment: Optional[str] = None


class TransferPrintType(str):
    OUTBOUND = "outbound"
    INBOUND = "inbound"


class TransferListFilter(BaseModel):
    status: Optional[TransferStatus] = None
    source_warehouse_id: Optional[UUID] = None
    target_warehouse_id: Optional[UUID] = None
    transfer_type: Optional[TransferType] = None
    created_by: Optional[UUID] = None
    start_date: Optional[str] = None
    end_date: Optional[str] = None
    page: int = Field(ge=1, default=1)
    page_size: int = Field(ge=1, le=100, default=20)


class InventoryListFilter(BaseModel):
    warehouse_id: Optional[UUID] = None
    product_id: Optional[UUID] = None
    page: int = Field(ge=1, default=1)
    page_size: int = Field(ge=1, le=100, default=100)


class LossListFilter(BaseModel):
    warehouse_id: Optional[UUID] = None
    transfer_id: Optional[UUID] = None
    start_date: Optional[str] = None
    end_date: Optional[str] = None
    page: int = Field(ge=1, default=1)
    page_size: int = Field(ge=1, le=100, default=20)


class LossSummaryFilter(BaseModel):
    warehouse_id: Optional[UUID] = None
    year: Optional[int] = None
    month: Optional[int] = None


class TransactionListFilter(BaseModel):
    warehouse_id: Optional[UUID] = None
    product_id: Optional[UUID] = None
    reference_id: Optional[UUID] = None
    start_date: Optional[str] = None
    end_date: Optional[str] = None
    page: int = Field(ge=1, default=1)
    page_size: int = Field(ge=1, le=100, default=20)


class ApiResponse(BaseModel):
    code: int
    message: str
    data: Optional[dict[str, Any]] = None


class TransferDetailResponse(BaseModel):
    transfer_id: UUID
    transfer_no: str
    source_warehouse_id: UUID
    target_warehouse_id: UUID
    source_warehouse_name: Optional[str] = None
    target_warehouse_name: Optional[str] = None
    status: TransferStatus
    items: list[TransferItem]
    transfer_type: TransferType
    approval_records: list[ApprovalRecord]
    required_approval_level: ApprovalLevel
    inter_company_approved: bool
    total_requested_amount: Decimal = Field(decimal_places=2)
    total_shipped_amount: Decimal = Field(decimal_places=2)
    total_arrived_amount: Decimal = Field(decimal_places=2)
    total_loss_amount: Decimal = Field(decimal_places=2)
    is_approved: bool
    created_by: UUID
    created_at: str
    confirmed_at: Optional[str] = None
    shipped_at: Optional[str] = None
    arrived_at: Optional[str] = None
    stocked_at: Optional[str] = None
    cancelled_at: Optional[str] = None
    cancelled_by: Optional[UUID] = None
    remark: Optional[str] = None


class TransferListResponse(BaseModel):
    total: int
    page: int
    page_size: int
    transfers: list[TransferDetailResponse]


class InventoryDetailResponse(BaseModel):
    inventory_id: UUID
    warehouse_id: UUID
    warehouse_name: Optional[str] = None
    product_id: UUID
    product_sku: Optional[str] = None
    product_name: Optional[str] = None
    product_unit: str = "件"
    product_unit_price: Optional[Decimal] = None
    available_quantity: int
    frozen_quantity: int
    total_quantity: int
    available_amount: Optional[Decimal] = None
    frozen_amount: Optional[Decimal] = None
    total_amount: Optional[Decimal] = None
    last_updated: str


class InventoryOverviewResponse(BaseModel):
    total: int
    page: int
    page_size: int
    inventories: list[InventoryDetailResponse]


class LossDetailResponse(BaseModel):
    loss_id: UUID
    transfer_id: UUID
    transfer_no: Optional[str] = None
    item_id: UUID
    product_id: UUID
    product_sku: Optional[str] = None
    product_name: Optional[str] = None
    warehouse_id: UUID
    warehouse_name: Optional[str] = None
    loss_quantity: int
    unit_price: Decimal = Field(decimal_places=2)
    loss_amount: Decimal = Field(decimal_places=2)
    loss_date: str
    created_at: str
    remark: Optional[str] = None


class LossListResponse(BaseModel):
    total: int
    page: int
    page_size: int
    losses: list[LossDetailResponse]


class LossSummaryResponse(BaseModel):
    summaries: list[LossMonthlySummary]


class TransactionDetailResponse(BaseModel):
    transaction_id: UUID
    warehouse_id: UUID
    warehouse_name: Optional[str] = None
    product_id: UUID
    product_sku: Optional[str] = None
    product_name: Optional[str] = None
    transaction_type: str
    quantity: int
    unit_price: Decimal = Field(decimal_places=2)
    amount: Decimal = Field(decimal_places=2)
    reference_type: str
    reference_id: UUID
    transaction_time: str
    remark: Optional[str] = None


class TransactionListResponse(BaseModel):
    total: int
    page: int
    page_size: int
    transactions: list[TransactionDetailResponse]


class PrintDocumentResponse(BaseModel):
    document_type: str
    transfer_no: str
    content: str
    generated_at: str


class WarehouseResponse(BaseModel):
    warehouse_id: UUID
    name: str
    location: str
    company_id: UUID
    created_at: str
    updated_at: str


class ProductResponse(BaseModel):
    product_id: UUID
    sku: str
    name: str
    unit_price: Decimal = Field(decimal_places=2)
    unit: str
    created_at: str
