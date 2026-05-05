import uuid
from datetime import datetime, timedelta, timezone
from typing import Dict, List, Optional, Set
import logging

from shared.models import (
    CrossDockOrder,
    CrossDockOrderCreate,
    CrossDockOrderResponse,
    InboundItem,
    OutboundItem,
    InboundScanRequest,
    OutboundScanRequest,
    CrossDockStatus,
    ItemDifference,
    ValidationResult,
    BatchCrossDockCreate,
    BatchCrossDockResponse,
    BatchStatus,
    EfficiencyStatistics,
    DailyReport,
    ZoneMonitorResponse,
    ExceptionPlan,
)
from server.repositories.in_memory_repository import InMemoryRepository
from server.exceptions import (
    OrderNotFoundException,
    InvalidStatusTransitionException,
    SameOperatorException,
    ItemMismatchException,
    TimeoutException,
    CrossDayException,
    BatchNotReadyException,
    DuplicateOrderException,
    OutboundOrderNotFoundException,
    ExceptionPlanNotFoundException,
)

logger = logging.getLogger(__name__)

TIMEOUT_MINUTES = 240
WAREHOUSE_SUPERVISOR = "warehouse_supervisor"


class CrossDockService:
    def __init__(self, repository: InMemoryRepository) -> None:
        self._repository = repository
        self._exception_plans = self._init_exception_plans()

    def _init_exception_plans(self) -> Dict[str, ExceptionPlan]:
        return {
            "timeout": ExceptionPlan(
                exception_type="timeout",
                severity="high",
                trigger_conditions=[
                    "越库单停留超过4小时",
                    "高峰期超时订单超过5个",
                ],
                response_steps=[
                    "1. 立即通知仓库主管",
                    "2. 检查越库区是否有积压",
                    "3. 安排优先处理超时订单",
                    "4. 记录超时原因",
                ],
                responsible_role="warehouse_operator",
                escalation_path=["warehouse_supervisor", "operations_manager"],
            ),
            "item_mismatch": ExceptionPlan(
                exception_type="item_mismatch",
                severity="medium",
                trigger_conditions=[
                    "入库商品与预期不一致",
                    "出库商品与入库不一致",
                ],
                response_steps=[
                    "1. 停止当前越库流程",
                    "2. 重新核对商品条码",
                    "3. 检查是否有混箱情况",
                    "4. 通知质量检验人员",
                ],
                responsible_role="warehouse_operator",
                escalation_path=["quality_inspector", "warehouse_supervisor"],
            ),
            "operator_insufficient": ExceptionPlan(
                exception_type="operator_insufficient",
                severity="medium",
                trigger_conditions=[
                    "待入库订单超过10个",
                    "平均停留时长超过60分钟",
                    "系统检测到高峰期",
                ],
                response_steps=[
                    "1. 检查当前操作员数量",
                    "2. 评估订单积压情况",
                    "3. 建议增加操作人员",
                    "4. 调度其他区域人员支援",
                ],
                responsible_role="warehouse_supervisor",
                escalation_path=["operations_manager"],
            ),
        }

    def create_order(self, order_create: CrossDockOrderCreate) -> CrossDockOrder:
        if self._repository.exists_by_order_number(order_create.order_number):
            raise DuplicateOrderException(order_create.order_number)

        order = CrossDockOrder(
            id=str(uuid.uuid4()),
            order_number=order_create.order_number,
            outbound_order_number=order_create.outbound_order_number,
            expected_items=order_create.expected_items,
            warehouse_id=order_create.warehouse_id,
            destination=order_create.destination,
            status=CrossDockStatus.PENDING_INBOUND,
            inbound_items=[],
            inbound_operator_id=None,
            outbound_operator_id=None,
            inbound_time=None,
            outbound_time=None,
        )
        return self._repository.save(order)

    def get_order(self, order_id: str) -> Optional[CrossDockOrder]:
        return self._repository.get_by_id(order_id)

    def get_all_orders(self) -> List[CrossDockOrder]:
        return self._repository.get_all()

    def to_response(self, order: CrossDockOrder) -> CrossDockOrderResponse:
        duration_minutes: Optional[float] = None
        if order.inbound_time and order.outbound_time:
            duration = order.outbound_time - order.inbound_time
            duration_minutes = duration.total_seconds() / 60
        elif order.inbound_time and order.status == CrossDockStatus.INBOUND_COMPLETED:
            duration = datetime.now(timezone.utc) - order.inbound_time
            duration_minutes = duration.total_seconds() / 60

        return CrossDockOrderResponse(
            id=order.id,
            order_number=order.order_number,
            outbound_order_number=order.outbound_order_number,
            status=order.status,
            expected_items=order.expected_items,
            inbound_items=order.inbound_items,
            warehouse_id=order.warehouse_id,
            destination=order.destination,
            inbound_operator_id=order.inbound_operator_id,
            outbound_operator_id=order.outbound_operator_id,
            inbound_time=order.inbound_time,
            outbound_time=order.outbound_time,
            created_at=order.created_at,
            is_cross_day=order.is_cross_day,
            alerts=order.alerts,
            duration_minutes=duration_minutes,
        )

    def validate_items(
        self, expected: Dict[str, int], actual: Dict[str, int]
    ) -> ValidationResult:
        differences: List[ItemDifference] = []
        extra_items: List[str] = []
        missing_items: List[str] = []

        all_skus: Set[str] = set(expected.keys()) | set(actual.keys())

        for sku in all_skus:
            exp_qty = expected.get(sku, 0)
            act_qty = actual.get(sku, 0)
            diff = act_qty - exp_qty

            if diff != 0:
                diff_type = "surplus" if diff > 0 else "deficit"
                differences.append(
                    ItemDifference(
                        sku=sku,
                        expected_quantity=exp_qty,
                        actual_quantity=act_qty,
                        difference=diff,
                        difference_type=diff_type,
                    )
                )

                if exp_qty == 0:
                    extra_items.append(sku)
                elif act_qty == 0:
                    missing_items.append(sku)

        success = len(differences) == 0
        message = "商品验证通过" if success else "商品存在差异"

        return ValidationResult(
            success=success,
            differences=differences,
            extra_items=extra_items,
            missing_items=missing_items,
            message=message,
        )

    def check_timeout(self, order: CrossDockOrder, current_time: datetime) -> bool:
        if order.inbound_time is None:
            return False
        elapsed = current_time - order.inbound_time
        return elapsed.total_seconds() > TIMEOUT_MINUTES * 60

    def check_cross_day(self, inbound_time: datetime, outbound_time: datetime) -> bool:
        return inbound_time.date() != outbound_time.date()

    def _items_to_dict(self, items: List[OutboundItem]) -> Dict[str, int]:
        result: Dict[str, int] = {}
        for item in items:
            result[item.sku] = result.get(item.sku, 0) + item.quantity
        return result

    def _inbound_items_to_dict(self, items: List[InboundItem]) -> Dict[str, int]:
        result: Dict[str, int] = {}
        for item in items:
            result[item.sku] = result.get(item.sku, 0) + item.quantity
        return result

    def inbound_scan(self, request: InboundScanRequest) -> CrossDockOrderResponse:
        order = self._repository.get_by_id(request.order_id)
        if order is None:
            raise OrderNotFoundException(request.order_id)

        if order.status != CrossDockStatus.PENDING_INBOUND:
            raise InvalidStatusTransitionException(
                current_status=order.status.value,
                target_status=CrossDockStatus.INBOUND_COMPLETED.value,
                order_id=order.id,
            )

        scan_time = request.scan_time or datetime.now(timezone.utc)

        if self.check_timeout(order, scan_time):
            order.status = CrossDockStatus.TIMEOUT_ALERT
            self._repository.save(order)
            raise TimeoutException(order.id, TIMEOUT_MINUTES)

        expected_dict = self._items_to_dict(order.expected_items)
        actual_dict = self._inbound_items_to_dict(request.items)
        validation = self.validate_items(expected_dict, actual_dict)

        if not validation.success:
            raise ItemMismatchException(order.id, validation)

        order.inbound_items = request.items
        order.inbound_operator_id = request.operator_id
        order.inbound_time = scan_time
        order.status = CrossDockStatus.INBOUND_COMPLETED

        self._repository.save(order)
        return self.to_response(order)

    def outbound_scan(self, request: OutboundScanRequest) -> CrossDockOrderResponse:
        order = self._repository.get_by_id(request.order_id)
        if order is None:
            raise OrderNotFoundException(request.order_id)

        if order.status not in [
            CrossDockStatus.INBOUND_COMPLETED,
            CrossDockStatus.TIMEOUT_ALERT,
        ]:
            raise InvalidStatusTransitionException(
                current_status=order.status.value,
                target_status=CrossDockStatus.OUTBOUND_COMPLETED.value,
                order_id=order.id,
            )

        if order.outbound_order_number:
            batch_status = self.get_batch_status(order.outbound_order_number)
            if batch_status.status != BatchStatus.FULLY_COMPLETED:
                raise BatchNotReadyException(
                    order.outbound_order_number,
                    batch_status.completed_count,
                    batch_status.total_count,
                )

        if order.inbound_operator_id == request.operator_id:
            raise SameOperatorException(request.operator_id, order.id)

        scan_time = request.scan_time or datetime.now(timezone.utc)

        if order.inbound_time and self.check_cross_day(order.inbound_time, scan_time):
            order.is_cross_day = True
            order.alerts.append("跨日越库异常")
            self._repository.save(order)
            raise CrossDayException(
                order.id,
                str(order.inbound_time.date()),
                str(scan_time.date()),
            )

        inbound_dict = self._inbound_items_to_dict(order.inbound_items)
        outbound_dict = self._items_to_dict(request.items)
        validation = self.validate_items(inbound_dict, outbound_dict)

        if not validation.success:
            raise ItemMismatchException(order.id, validation)

        order.outbound_operator_id = request.operator_id
        order.outbound_time = scan_time
        order.status = CrossDockStatus.OUTBOUND_COMPLETED

        self._repository.save(order)
        return self.to_response(order)

    def create_batch_cross_dock(self, batch_create: BatchCrossDockCreate) -> BatchCrossDockResponse:
        inbound_order_ids: List[str] = []

        for order_number in batch_create.inbound_order_numbers:
            if self._repository.exists_by_order_number(order_number):
                raise DuplicateOrderException(order_number)

        for order_number in batch_create.inbound_order_numbers:
            order = CrossDockOrder(
                id=str(uuid.uuid4()),
                order_number=order_number,
                outbound_order_number=batch_create.outbound_order_number,
                expected_items=batch_create.expected_items,
                warehouse_id=batch_create.warehouse_id,
                destination=batch_create.destination,
                status=CrossDockStatus.PENDING_INBOUND,
                inbound_items=[],
                inbound_operator_id=None,
                outbound_operator_id=None,
                inbound_time=None,
                outbound_time=None,
            )
            saved_order = self._repository.save(order)
            inbound_order_ids.append(saved_order.id)

        return BatchCrossDockResponse(
            outbound_order_number=batch_create.outbound_order_number,
            inbound_order_ids=inbound_order_ids,
            status=BatchStatus.PENDING,
            completed_count=0,
            total_count=len(inbound_order_ids),
        )

    def get_batch_status(self, outbound_order_number: str) -> BatchCrossDockResponse:
        orders = self._repository.get_by_outbound_order_number(outbound_order_number)
        if not orders:
            raise OutboundOrderNotFoundException(outbound_order_number)

        inbound_order_ids = [o.id for o in orders]
        completed_count = sum(
            1 for o in orders if o.status == CrossDockStatus.INBOUND_COMPLETED
        )
        total_count = len(orders)

        if completed_count == 0:
            status = BatchStatus.PENDING
        elif completed_count == total_count:
            status = BatchStatus.FULLY_COMPLETED
        else:
            status = BatchStatus.PARTIALLY_COMPLETED

        return BatchCrossDockResponse(
            outbound_order_number=outbound_order_number,
            inbound_order_ids=inbound_order_ids,
            status=status,
            completed_count=completed_count,
            total_count=total_count,
        )

    def get_efficiency_statistics(
        self, start_time: datetime, end_time: datetime
    ) -> EfficiencyStatistics:
        orders = self._repository.get_all()
        completed_orders = [
            o for o in orders
            if o.status == CrossDockStatus.OUTBOUND_COMPLETED
            and o.inbound_time is not None
            and o.outbound_time is not None
            and start_time <= o.outbound_time <= end_time
        ]

        if not completed_orders:
            return EfficiencyStatistics(
                average_duration_minutes=0.0,
                timeout_rate=0.0,
                daily_volume=0,
                period_start=start_time,
                period_end=end_time,
            )

        durations: List[float] = []
        timeout_count = 0
        for order in completed_orders:
            if order.inbound_time and order.outbound_time:
                duration = (order.outbound_time - order.inbound_time).total_seconds() / 60
                durations.append(duration)
                if duration > TIMEOUT_MINUTES:
                    timeout_count += 1

        average_duration = sum(durations) / len(durations) if durations else 0.0
        timeout_rate = (timeout_count / len(completed_orders)) * 100 if completed_orders else 0.0

        days = (end_time - start_time).days or 1
        daily_volume = len(completed_orders) / days

        return EfficiencyStatistics(
            average_duration_minutes=round(average_duration, 2),
            timeout_rate=round(timeout_rate, 2),
            daily_volume=int(daily_volume),
            period_start=start_time,
            period_end=end_time,
        )

    def generate_daily_report(self, report_date: datetime) -> DailyReport:
        from datetime import timezone

        utc_date = report_date.astimezone(timezone.utc) if report_date.tzinfo else report_date.replace(tzinfo=timezone.utc)
        start_of_day = datetime(utc_date.year, utc_date.month, utc_date.day, 0, 0, 0, tzinfo=timezone.utc)
        end_of_day = datetime(utc_date.year, utc_date.month, utc_date.day, 23, 59, 59, tzinfo=timezone.utc)

        orders = self._repository.get_all()
        daily_orders = [
            o for o in orders
            if o.created_at and start_of_day <= o.created_at <= end_of_day
        ]

        total_orders = len(daily_orders)
        completed_orders = sum(
            1 for o in daily_orders if o.status == CrossDockStatus.OUTBOUND_COMPLETED
        )
        pending_orders = sum(
            1 for o in daily_orders
            if o.status in [CrossDockStatus.PENDING_INBOUND, CrossDockStatus.INBOUND_COMPLETED]
        )
        timeout_orders = sum(
            1 for o in daily_orders if o.status == CrossDockStatus.TIMEOUT_ALERT
        )
        cross_day_orders = sum(1 for o in daily_orders if o.is_cross_day)

        efficiency = self.get_efficiency_statistics(start_of_day, end_of_day)

        return DailyReport(
            report_date=report_date,
            total_orders=total_orders,
            completed_orders=completed_orders,
            pending_orders=pending_orders,
            timeout_orders=timeout_orders,
            cross_day_orders=cross_day_orders,
            efficiency=efficiency,
        )

    def get_zone_monitor(self, warehouse_id: str) -> ZoneMonitorResponse:
        orders = self._repository.get_all()
        warehouse_orders = [o for o in orders if o.warehouse_id == warehouse_id]

        pending_inbound_count = sum(
            1 for o in warehouse_orders if o.status == CrossDockStatus.PENDING_INBOUND
        )
        inbound_completed_count = sum(
            1 for o in warehouse_orders if o.status == CrossDockStatus.INBOUND_COMPLETED
        )
        timeout_count = sum(
            1 for o in warehouse_orders if o.status == CrossDockStatus.TIMEOUT_ALERT
        )

        peak_suggestion: Optional[str] = None
        total_pending = pending_inbound_count + inbound_completed_count
        if total_pending > 10:
            peak_suggestion = "当前订单积压较多，建议增加2-3名操作人员支援越库区"
        elif total_pending > 5:
            peak_suggestion = "订单量有所增加，请关注越库区处理进度"

        return ZoneMonitorResponse(
            zone_name=f"越库区-{warehouse_id}",
            pending_inbound_count=pending_inbound_count,
            inbound_completed_count=inbound_completed_count,
            timeout_count=timeout_count,
            peak_hour_suggestion=peak_suggestion,
        )

    def get_exception_plan(self, exception_type: str) -> ExceptionPlan:
        plan = self._exception_plans.get(exception_type)
        if plan is None:
            raise ExceptionPlanNotFoundException(exception_type)
        return plan

    def send_timeout_alert(self, order: CrossDockOrder) -> None:
        alert_message = (
            f"[越库超时告警] 越库单 {order.order_number} 已停留超过{TIMEOUT_MINUTES}分钟。"
            f"入库时间: {order.inbound_time}, 当前状态: {order.status.value}"
        )
        order.alerts.append(alert_message)
        self._repository.save(order)

        logger.warning(alert_message)
        logger.info(f"通知仓库主管: {WAREHOUSE_SUPERVISOR}")
