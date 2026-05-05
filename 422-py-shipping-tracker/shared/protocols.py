from datetime import datetime
from typing import Optional, Protocol

from shared.models import (
    AlertResponse,
    BatchQueryResponse,
    CustomStatus,
    TrackingNode,
    TrackingNodeCreate,
    Waybill,
    WaybillCreate,
    WaybillQueryResponse,
    WaybillSplitRequest,
    WaybillSplitResponse,
)


class ShippingServiceProtocol(Protocol):
    def create_waybill(self, waybill_create: WaybillCreate) -> Waybill:
        ...

    def get_waybill(self, waybill_number: str) -> Optional[Waybill]:
        ...

    def add_tracking_node(
        self, waybill_number: str, node_create: TrackingNodeCreate
    ) -> TrackingNode:
        ...

    def query_waybill(self, waybill_number: str) -> WaybillQueryResponse:
        ...

    def batch_query_waybills(self, waybill_numbers: list[str]) -> BatchQueryResponse:
        ...

    def split_waybill(self, split_request: WaybillSplitRequest) -> WaybillSplitResponse:
        ...

    def update_custom_status(
        self, waybill_number: str, custom_status: CustomStatus, operator_id: str,
        operator_name: Optional[str]
    ) -> Waybill:
        ...

    def check_abnormal_waybills(self, current_time: datetime) -> list[AlertResponse]:
        ...

    def generate_alerts(self, current_time: datetime) -> list[AlertResponse]:
        ...

    def get_sub_waybills(self, parent_waybill_number: str) -> list[Waybill]:
        ...
