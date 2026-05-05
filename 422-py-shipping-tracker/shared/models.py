from datetime import datetime
from enum import Enum, auto
from typing import Optional

from pydantic import BaseModel, Field, field_validator
from typing_extensions import Annotated


class NodeType(str, Enum):
    PICKED_UP = "已揽收"
    IN_TRANSIT = "运输中"
    ARRIVED_AT_TRANSIT = "到达中转站"
    OUT_FOR_DELIVERY = "派送中"
    SIGNED = "已签收"


class WaybillStatus(str, Enum):
    CREATED = "已创建"
    IN_TRANSIT = "运输中"
    OUT_FOR_DELIVERY = "派送中"
    SIGNED = "已签收"
    ABNORMAL = "异常"


class CustomStatus(str, Enum):
    PENDING_DECLARATION = "待报关"
    DECLARING = "报关中"
    DECLARED = "已报关"
    CUSTOMS_CLEARED = "已清关"
    CUSTOMS_HELD = "海关扣留"


class OperatorType(str, Enum):
    SYSTEM = "system"
    COURIER = "courier"


class Location(BaseModel):
    address: str = Field(..., min_length=1, description="地点地址")
    latitude: Optional[float] = Field(None, ge=-90, le=90, description="纬度")
    longitude: Optional[float] = Field(None, ge=-180, le=180, description="经度")

    @field_validator("latitude")
    @classmethod
    def validate_latitude(cls, v: Optional[float]) -> Optional[float]:
        if v is not None and (v < -90 or v > 90):
            raise ValueError("纬度必须在 -90 到 90 之间")
        return v

    @field_validator("longitude")
    @classmethod
    def validate_longitude(cls, v: Optional[float]) -> Optional[float]:
        if v is not None and (v < -180 or v > 180):
            raise ValueError("经度必须在 -180 到 180 之间")
        return v


class SignedBy(BaseModel):
    name: str = Field(..., min_length=1, description="签收人姓名")
    is_authorized: bool = Field(False, description="是否代签收")
    authorized_relation: Optional[str] = Field(None, description="与收件人关系")

    @field_validator("authorized_relation")
    @classmethod
    def validate_authorized_relation(cls, v: Optional[str], info: object) -> Optional[str]:
        data = getattr(info, "data", {}) if hasattr(info, "data") else {}
        is_authorized = data.get("is_authorized") if isinstance(data, dict) else False
        if is_authorized and not v:
            raise ValueError("代签收时必须填写与收件人关系")
        return v


class Operator(BaseModel):
    operator_type: OperatorType
    operator_id: str = Field(..., min_length=1, description="操作人ID")
    operator_name: Optional[str] = Field(None, description="操作人名称")


class TrackingNodeCreate(BaseModel):
    node_type: NodeType
    location: Location
    timestamp: datetime = Field(..., description="UTC 时间")
    operator: Operator
    signed_by: Optional[SignedBy] = Field(None, description="签收信息，签收节点必填")

    @field_validator("timestamp")
    @classmethod
    def validate_utc(cls, v: datetime) -> datetime:
        if v.tzinfo is not None:
            from datetime import timezone
            if v.utcoffset() != timezone.utc.utcoffset(v):
                raise ValueError("时间必须是 UTC 时区")
        return v

    @field_validator("signed_by")
    @classmethod
    def validate_signed_by(cls, v: Optional[SignedBy], info: object) -> Optional[SignedBy]:
        data = getattr(info, "data", {}) if hasattr(info, "data") else {}
        node_type = data.get("node_type") if isinstance(data, dict) else None
        if node_type == NodeType.SIGNED and v is None:
            raise ValueError("签收节点必须填写签收人信息")
        return v


class TrackingNode(TrackingNodeCreate):
    node_id: str = Field(..., description="节点唯一标识")


class StatusChangeHistory(BaseModel):
    from_status: WaybillStatus
    to_status: WaybillStatus
    changed_at: datetime
    operator: Operator
    reason: Optional[str] = None


class Waybill(BaseModel):
    waybill_number: str = Field(..., min_length=1, description="运单号")
    parent_waybill_number: Optional[str] = Field(None, description="父运单号")
    is_international: bool = Field(False, description="是否国际运单")
    sender: str = Field(..., min_length=1, description="发件人")
    receiver: str = Field(..., min_length=1, description="收件人")
    receiver_address: str = Field(..., min_length=1, description="收件地址")
    status: WaybillStatus = WaybillStatus.CREATED
    nodes: list[TrackingNode] = Field(default_factory=list)
    custom_status: Optional[CustomStatus] = Field(None, description="海关状态")
    is_abnormal: bool = Field(False, description="是否异常")
    last_alert_time: Optional[datetime] = Field(None, description="上次告警时间")
    created_at: datetime = Field(..., description="创建时间 UTC")
    updated_at: datetime = Field(..., description="更新时间 UTC")
    status_history: list[StatusChangeHistory] = Field(default_factory=list)

    @property
    def last_node_time(self) -> Optional[datetime]:
        if not self.nodes:
            return None
        return max(node.timestamp for node in self.nodes)

    @property
    def is_signed(self) -> bool:
        return self.status == WaybillStatus.SIGNED


class WaybillCreate(BaseModel):
    sender: str = Field(..., min_length=1, description="发件人")
    receiver: str = Field(..., min_length=1, description="收件人")
    receiver_address: str = Field(..., min_length=1, description="收件地址")
    is_international: bool = Field(False, description="是否国际运单")
    operator: Operator


class WaybillQueryResponse(BaseModel):
    waybill: Waybill
    tracking_history: list[TrackingNode] = Field(
        default_factory=list, description="按时间正序排列的物流轨迹"
    )


class BatchQueryRequest(BaseModel):
    waybill_numbers: list[str] = Field(
        ..., min_length=1, max_length=50, description="运单号列表，最多50个"
    )


class BatchQueryResponse(BaseModel):
    results: dict[str, WaybillQueryResponse] = Field(
        default_factory=dict, description="运单号到查询结果的映射"
    )
    not_found: list[str] = Field(default_factory=list, description="未找到的运单号列表")


class WaybillSplitRequest(BaseModel):
    parent_waybill_number: str
    sub_waybills: list[WaybillCreate] = Field(..., min_length=1, description="子运单列表")


class WaybillSplitResponse(BaseModel):
    parent_waybill_number: str
    sub_waybill_numbers: list[str] = Field(default_factory=list)


class CustomStatusUpdateRequest(BaseModel):
    custom_status: CustomStatus
    operator: Operator


class AlertResponse(BaseModel):
    waybill_number: str
    current_status: WaybillStatus
    last_node_time: Optional[datetime]
    hours_since_last_node: float
    alert_time: datetime
