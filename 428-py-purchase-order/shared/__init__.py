"""
采购订单管理系统 - 共享模块

包含 Pydantic 数据模型、错误码常量和协议定义。
"""

from shared.models import (
    PurchaseOrderStatus,
    ApprovalType,
    QualityResult,
    DiscrepancyAction,
    SupplierRating,
    OrderItem,
    ApprovalRecord,
    ReceiptRecord,
    QualityRecord,
    DiscrepancyRecord,
    PaymentRecord,
    TimelineEvent,
    SupplierStats,
    Supplier,
    PurchaseOrder,
    ArchivedOrder,
)
from shared.requests import (
    CreateOrderRequest,
    ApproveRequest,
    RejectRequest,
    PlaceOrderRequest,
    ReceiveRequest,
    QualityInspectRequest,
    ResolveDiscrepancyRequest,
