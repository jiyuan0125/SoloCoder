from datetime import datetime
from enum import Enum
from typing import Optional, List, Dict, Any

from pydantic import BaseModel, Field, field_validator


class OrderStatus(str, Enum):
    PENDING = "pending"
    SHIPPED = "shipped"
    DELIVERED = "delivered"
    COMPLETED = "completed"


class ReturnStatus(str, Enum):
    APPLIED = "applied"
    WAREHOUSE_RECEIVED = "warehouse_received"
    INSPECTED = "inspected"
    STOCKED_IN = "stocked_in"
    RETURNED_TO_CUSTOMER = "returned_to_customer"
    EXPIRED = "expired"
    CANCELLED = "cancelled"


class InspectionResult(str, Enum):
    GOOD = "good"
    MINOR_DEFECT = "minor_defect"
    SEVERE_DAMAGE = "severe_damage"


class DisposalReason(str, Enum):
    QUALITY_ISSUE = "quality_issue"
    SIZE_MISMATCH = "size_mismatch"
    DISLIKE = "dislike"
    WRONG_ITEM = "wrong_item"
    OTHER = "other"


class OutboundOrderItem(BaseModel):
    sku: str
    product_name: str
    quantity: int
    unit_price: float
    category: str = ""


class OutboundOrder(BaseModel):
    order_id: str
    customer_id: str
    items: List[OutboundOrderItem]
    total_amount: float
    status: OrderStatus
    created_at: datetime
    updated_at: datetime
    shipped_at: Optional[datetime] = None
    delivered_at: Optional[datetime] = None


class ReturnOrderItem(BaseModel):
    sku: str
    product_name: str
    quantity: int
    original_unit_price: float
    category: str = ""
    inspection_result: Optional[InspectionResult] = None
    disposal_price: Optional[float] = None


class ReturnOrder(BaseModel):
    return_order_id: str
    outbound_order_id: str
    status: ReturnStatus
    items: List[ReturnOrderItem]
    reason: DisposalReason
    customer_note: str
    created_at: datetime
    expiry_date: datetime
    warehouse_received_at: Optional[datetime] = None
    warehouse_received_by: Optional[str] = None
    inspected_at: Optional[datetime] = None
    inspected_by: Optional[str] = None
    stocked_in_at: Optional[datetime] = None
    stocked_in_by: Optional[str] = None
    stockin_deadline: Optional[datetime] = None


class ScrapRecord(BaseModel):
    scrap_id: str
    return_order_id: str
    sku: str
    product_name: str
    quantity: int
    original_price: float
    reason: DisposalReason
    detailed_reason: str = ""
    approver: str
    created_at: datetime
    month: str


class DefectiveItem(BaseModel):
    defective_id: str
    sku: str
    product_name: str
    quantity: int
    unit_price: float
    defective_price: float
    category: str
    return_order_id: str
    created_at: datetime
    converted_to_normal: bool = False
    converted_at: Optional[datetime] = None
    converted_by: Optional[str] = None


class OutboundOrderCreate(BaseModel):
    order_id: str
    customer_id: str
    items: List[OutboundOrderItem]

    @field_validator("items")
    @classmethod
    def validate_items(cls, v: List[OutboundOrderItem]) -> List[OutboundOrderItem]:
        if len(v) == 0:
            raise ValueError("订单至少需要一个商品")
        for item in v:
            if item.quantity <= 0:
                raise ValueError(f"商品 {item.sku} 数量必须大于0")
            if item.unit_price < 0:
                raise ValueError(f"商品 {item.sku} 单价不能为负")
        return v


class ReturnApplyRequest(BaseModel):
    return_order_id: str
    outbound_order_id: str
    items: List[ReturnOrderItem]
    reason: DisposalReason
    customer_note: str = ""

    @field_validator("items")
    @classmethod
    def validate_items(cls, v: List[ReturnOrderItem]) -> List[ReturnOrderItem]:
        if len(v) == 0:
            raise ValueError("退货单至少需要一个商品")
        for item in v:
            if item.quantity <= 0:
                raise ValueError(f"商品 {item.sku} 数量必须大于0")
        return v


class ReturnApplyResponse(BaseModel):
    return_order_id: str
    outbound_order_id: str
    status: ReturnStatus
    created_at: datetime
    expiry_date: datetime
    message: str = "退货申请已提交"


class WarehouseReceiveRequest(BaseModel):
    return_order_id: str
    received_by: str
    received_at: Optional[datetime] = None


class InspectionRequest(BaseModel):
    return_order_id: str
    inspector: str
    item_results: Dict[str, InspectionResult]
    inspection_date: Optional[datetime] = None


class StockInRequest(BaseModel):
    return_order_id: str
    stock_in_by: str
    stock_in_at: Optional[datetime] = None


class DefectiveToNormalRequest(BaseModel):
    defective_id: str
    inspector: str
    inspection_date: Optional[datetime] = None


class ReturnStatisticsRequest(BaseModel):
    start_date: datetime
    end_date: datetime
    category: Optional[str] = None

    @field_validator("end_date")
    @classmethod
    def validate_date_range(cls, v: datetime, info: Any) -> datetime:
        start_date = info.data.get("start_date")
        if start_date and v < start_date:
            raise ValueError("结束日期必须大于开始日期")
        return v


class ReturnStatisticsResponse(BaseModel):
    total_returns: int
    total_quantity: int
    good_count: int
    good_quantity: int
    defective_count: int
    defective_quantity: int
    scrap_count: int
    scrap_quantity: int
    good_rate: float
    scrap_rate: float
    period_start: datetime
    period_end: datetime
    category: Optional[str] = None


class ScrapLedgerRequest(BaseModel):
    year: int
    month: int
    reason: Optional[DisposalReason] = None


class ScrapLedgerItem(BaseModel):
    reason: DisposalReason
    count: int
    quantity: int
    total_value: float


class ScrapLedgerResponse(BaseModel):
    year: int
    month: int
    total_records: int
    total_quantity: int
    total_value: float
    breakdown: List[ScrapLedgerItem]
    filter_reason: Optional[DisposalReason] = None


class QualityAlertRequest(BaseModel):
    supplier_id: str
    period_days: int = 30
    threshold: Optional[float] = None


class QualityAlertResponse(BaseModel):
    supplier_id: str
    period_days: int
    threshold: float
    actual_rate: float
    alert_triggered: bool
    message: str


class ErrorResponse(BaseModel):
    code: int
    message: str
    detail: Optional[str] = None


class OutboundOrderResponse(BaseModel):
    order_id: str
    customer_id: str
    items: List[OutboundOrderItem]
    total_amount: float
    status: OrderStatus
    created_at: datetime
    updated_at: datetime
    shipped_at: Optional[datetime] = None
    delivered_at: Optional[datetime] = None


class ReturnOrderResponse(BaseModel):
    return_order_id: str
    outbound_order_id: str
    status: ReturnStatus
    items: List[ReturnOrderItem]
    reason: DisposalReason
    customer_note: str
    created_at: datetime
    expiry_date: datetime
    warehouse_received_at: Optional[datetime] = None
    warehouse_received_by: Optional[str] = None
    inspected_at: Optional[datetime] = None
    inspected_by: Optional[str] = None
    stocked_in_at: Optional[datetime] = None
    stocked_in_by: Optional[str] = None


class DefectiveItemResponse(BaseModel):
    defective_id: str
    sku: str
    product_name: str
    quantity: int
    unit_price: float
    defective_price: float
    category: str
    return_order_id: str
    created_at: datetime
    converted_to_normal: bool
    converted_at: Optional[datetime] = None
    converted_by: Optional[str] = None
