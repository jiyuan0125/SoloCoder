from datetime import datetime, timezone
from enum import Enum, auto
from typing import Optional, List, Dict, Any
from pydantic import BaseModel, Field


def _utc_now() -> datetime:
    return datetime.now(timezone.utc)


class CrossDockStatus(str, Enum):
    PENDING_INBOUND = "pending_inbound"
    INBOUND_COMPLETED = "inbound_completed"
    OUTBOUND_COMPLETED = "outbound_completed"
    TIMEOUT_ALERT = "timeout_alert"


class BatchStatus(str, Enum):
    PENDING = "pending"
    PARTIALLY_COMPLETED = "partially_completed"
    FULLY_COMPLETED = "fully_completed"


class InboundItem(BaseModel):
    sku: str = Field(..., description="商品SKU")
    quantity: int = Field(..., gt=0, description="数量")
    batch_number: Optional[str] = Field(None, description="批次号")
    expiry_date: Optional[datetime] = Field(None, description="有效期")


class OutboundItem(BaseModel):
    sku: str = Field(..., description="商品SKU")
    quantity: int = Field(..., gt=0, description="数量")


class ItemDifference(BaseModel):
    sku: str = Field(..., description="商品SKU")
    expected_quantity: int = Field(..., description="期望数量")
    actual_quantity: int = Field(..., description="实际数量")
    difference: int = Field(..., description="差异数量")
    difference_type: str = Field(..., description="差异类型：surplus(多) / deficit(少)")


class ValidationResult(BaseModel):
    success: bool = Field(..., description="验证是否成功")
    differences: List[ItemDifference] = Field(default_factory=list, description="差异明细")
    extra_items: List[str] = Field(default_factory=list, description="多余的商品SKU")
    missing_items: List[str] = Field(default_factory=list, description="缺失的商品SKU")
    message: Optional[str] = Field(None, description="验证消息")


class CrossDockOrderCreate(BaseModel):
    order_number: str = Field(..., description="越库单号")
    outbound_order_number: Optional[str] = Field(None, description="出库单号（批量越库时使用）")
    expected_items: List[OutboundItem] = Field(..., description="期望出库的商品列表")
    warehouse_id: str = Field(..., description="仓库ID")
    destination: str = Field(..., description="目的地")


class CrossDockOrder(BaseModel):
    id: str = Field(..., description="越库单唯一ID")
    order_number: str = Field(..., description="越库单号")
    outbound_order_number: Optional[str] = Field(None, description="出库单号（批量越库时使用）")
    status: CrossDockStatus = Field(default=CrossDockStatus.PENDING_INBOUND, description="状态")
    expected_items: List[OutboundItem] = Field(..., description="期望出库的商品列表")
    inbound_items: List[InboundItem] = Field(default_factory=list, description="实际入库的商品列表")
    warehouse_id: str = Field(..., description="仓库ID")
    destination: str = Field(..., description="目的地")
    inbound_operator_id: Optional[str] = Field(None, description="入库操作员ID")
    outbound_operator_id: Optional[str] = Field(None, description="出库操作员ID")
    inbound_time: Optional[datetime] = Field(None, description="入库时间")
    outbound_time: Optional[datetime] = Field(None, description="出库时间")
    created_at: datetime = Field(default_factory=_utc_now, description="创建时间")
    is_cross_day: bool = Field(default=False, description="是否跨日")
    alerts: List[str] = Field(default_factory=list, description="告警列表")


class CrossDockOrderResponse(BaseModel):
    id: str = Field(..., description="越库单唯一ID")
    order_number: str = Field(..., description="越库单号")
    outbound_order_number: Optional[str] = Field(None, description="出库单号")
    status: CrossDockStatus = Field(..., description="状态")
    expected_items: List[OutboundItem] = Field(..., description="期望商品列表")
    inbound_items: List[InboundItem] = Field(default_factory=list, description="入库商品列表")
    warehouse_id: str = Field(..., description="仓库ID")
    destination: str = Field(..., description="目的地")
    inbound_operator_id: Optional[str] = Field(None, description="入库操作员ID")
    outbound_operator_id: Optional[str] = Field(None, description="出库操作员ID")
    inbound_time: Optional[datetime] = Field(None, description="入库时间")
    outbound_time: Optional[datetime] = Field(None, description="出库时间")
    created_at: datetime = Field(..., description="创建时间")
    is_cross_day: bool = Field(default=False, description="是否跨日")
    alerts: List[str] = Field(default_factory=list, description="告警列表")
    duration_minutes: Optional[float] = Field(None, description="停留时长（分钟）")


class InboundScanRequest(BaseModel):
    order_id: str = Field(..., description="越库单ID")
    operator_id: str = Field(..., description="操作员ID")
    items: List[InboundItem] = Field(..., description="扫码入库的商品列表")
    scan_time: Optional[datetime] = Field(None, description="扫码时间")


class OutboundScanRequest(BaseModel):
    order_id: str = Field(..., description="越库单ID")
    operator_id: str = Field(..., description="操作员ID")
    items: List[OutboundItem] = Field(..., description="扫码出库的商品列表")
    scan_time: Optional[datetime] = Field(None, description="扫码时间")


class BatchCrossDockCreate(BaseModel):
    outbound_order_number: str = Field(..., description="出库单号")
    inbound_order_numbers: List[str] = Field(..., description="入库单号列表")
    expected_items: List[OutboundItem] = Field(..., description="期望出库的商品列表")
    warehouse_id: str = Field(..., description="仓库ID")
    destination: str = Field(..., description="目的地")


class BatchCrossDockResponse(BaseModel):
    outbound_order_number: str = Field(..., description="出库单号")
    inbound_order_ids: List[str] = Field(..., description="入库单ID列表")
    status: BatchStatus = Field(..., description="批量状态")
    completed_count: int = Field(..., description="已完成入库的数量")
    total_count: int = Field(..., description="总数量")


class EfficiencyStatistics(BaseModel):
    average_duration_minutes: float = Field(..., description="平均停留时长（分钟）")
    timeout_rate: float = Field(..., description="超时率（百分比）")
    daily_volume: int = Field(..., description="日处理量")
    period_start: datetime = Field(..., description="统计周期开始时间")
    period_end: datetime = Field(..., description="统计周期结束时间")


class DailyReport(BaseModel):
    report_date: datetime = Field(..., description="报告日期")
    total_orders: int = Field(..., description="总越库单数")
    completed_orders: int = Field(..., description="已完成单数")
    pending_orders: int = Field(..., description="待处理单数")
    timeout_orders: int = Field(..., description="超时单数")
    cross_day_orders: int = Field(..., description="跨日单数")
    efficiency: EfficiencyStatistics = Field(..., description="效率统计")


class ZoneMonitorResponse(BaseModel):
    zone_name: str = Field(..., description="区域名称")
    pending_inbound_count: int = Field(..., description="待入库数量")
    inbound_completed_count: int = Field(..., description="已入库待出库数量")
    timeout_count: int = Field(..., description="超时数量")
    peak_hour_suggestion: Optional[str] = Field(None, description="高峰期建议")


class ExceptionPlan(BaseModel):
    exception_type: str = Field(..., description="异常类型")
    severity: str = Field(..., description="严重程度")
    trigger_conditions: List[str] = Field(..., description="触发条件")
    response_steps: List[str] = Field(..., description="响应步骤")
    responsible_role: str = Field(..., description="负责角色")
    escalation_path: List[str] = Field(..., description="升级路径")
