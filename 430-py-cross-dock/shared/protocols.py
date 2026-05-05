from typing import Protocol, Optional, List
from datetime import datetime

from shared.models import (
    CrossDockOrder,
    CrossDockOrderCreate,
    CrossDockOrderResponse,
    InboundScanRequest,
    OutboundScanRequest,
    ValidationResult,
    BatchCrossDockCreate,
    BatchCrossDockResponse,
    EfficiencyStatistics,
    DailyReport,
    ZoneMonitorResponse,
    ExceptionPlan,
)


class CrossDockServiceProtocol(Protocol):
    def create_order(self, order_create: CrossDockOrderCreate) -> CrossDockOrder: ...

    def get_order(self, order_id: str) -> Optional[CrossDockOrder]: ...

    def get_all_orders(self) -> List[CrossDockOrder]: ...

    def inbound_scan(self, request: InboundScanRequest) -> CrossDockOrderResponse: ...

    def outbound_scan(self, request: OutboundScanRequest) -> CrossDockOrderResponse: ...

    def validate_items(
        self, expected: dict[str, int], actual: dict[str, int]
    ) -> ValidationResult: ...

    def check_timeout(self, order: CrossDockOrder, current_time: datetime) -> bool: ...

    def check_cross_day(self, inbound_time: datetime, outbound_time: datetime) -> bool: ...

    def create_batch_cross_dock(self, batch_create: BatchCrossDockCreate) -> BatchCrossDockResponse: ...

    def get_batch_status(self, outbound_order_number: str) -> BatchCrossDockResponse: ...

    def get_efficiency_statistics(
        self, start_time: datetime, end_time: datetime
    ) -> EfficiencyStatistics: ...

    def generate_daily_report(self, report_date: datetime) -> DailyReport: ...

    def get_zone_monitor(self, warehouse_id: str) -> ZoneMonitorResponse: ...

    def get_exception_plan(self, exception_type: str) -> ExceptionPlan: ...

    def send_timeout_alert(self, order: CrossDockOrder) -> None: ...
