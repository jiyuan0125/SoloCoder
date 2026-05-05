from datetime import datetime, timedelta
from typing import Dict, List, Optional

from shared.models import (
    ScrapRecord,
    DefectiveItem,
    ReturnStatus,
    InspectionResult,
    DisposalReason,
    ReturnStatisticsRequest,
    ReturnStatisticsResponse,
    ScrapLedgerRequest,
    ScrapLedgerResponse,
    ScrapLedgerItem,
    QualityAlertRequest,
    QualityAlertResponse,
)
from shared.constants import ReturnConstants
from server.data_store import DataStore, get_data_store
from server.models import ReturnOrder


class StatisticsService:
    def __init__(self, data_store: Optional[DataStore] = None) -> None:
        self._data_store = data_store or get_data_store()

    def get_return_statistics(self, request: ReturnStatisticsRequest) -> ReturnStatisticsResponse:
        returns = self._data_store.list_return_orders()

        filtered_returns: List[ReturnOrder] = []
        for ret in returns:
            if ret.created_at >= request.start_date and ret.created_at <= request.end_date:
                if request.category is None:
                    filtered_returns.append(ret)
                else:
                    for item in ret.items:
                        if item.category == request.category:
                            filtered_returns.append(ret)
                            break

        total_returns = len(filtered_returns)
        total_quantity = 0
        good_count = 0
        good_quantity = 0
        defective_count = 0
        defective_quantity = 0
        scrap_count = 0
        scrap_quantity = 0

        for ret in filtered_returns:
            for item in ret.items:
                if request.category and item.category != request.category:
                    continue

                total_quantity += item.quantity

                if item.inspection_result == InspectionResult.GOOD:
                    good_count += 1
                    good_quantity += item.quantity
                elif item.inspection_result == InspectionResult.MINOR_DEFECT:
                    defective_count += 1
                    defective_quantity += item.quantity
                elif item.inspection_result == InspectionResult.SEVERE_DAMAGE:
                    scrap_count += 1
                    scrap_quantity += item.quantity

        good_rate = 0.0
        scrap_rate = 0.0
        if total_quantity > 0:
            good_rate = good_quantity / total_quantity
            scrap_rate = scrap_quantity / total_quantity

        return ReturnStatisticsResponse(
            total_returns=total_returns,
            total_quantity=total_quantity,
            good_count=good_count,
            good_quantity=good_quantity,
            defective_count=defective_count,
            defective_quantity=defective_quantity,
            scrap_count=scrap_count,
            scrap_quantity=scrap_quantity,
            good_rate=good_rate,
            scrap_rate=scrap_rate,
            period_start=request.start_date,
            period_end=request.end_date,
            category=request.category,
        )

    def get_scrap_ledger(self, request: ScrapLedgerRequest) -> ScrapLedgerResponse:
        scrap_records = self._data_store.list_scrap_records()

        target_month = f"{request.year}-{request.month:02d}"

        filtered_records: List[ScrapRecord] = []
        for record in scrap_records:
            if record.month == target_month:
                if request.reason is None or record.reason == request.reason:
                    filtered_records.append(record)

        total_records = len(filtered_records)
        total_quantity = sum(r.quantity for r in filtered_records)
        total_value = sum(r.quantity * r.original_price for r in filtered_records)

        reason_stats: Dict[DisposalReason, ScrapLedgerItem] = {}
        for record in filtered_records:
            if record.reason not in reason_stats:
                reason_stats[record.reason] = ScrapLedgerItem(
                    reason=record.reason,
                    count=0,
                    quantity=0,
                    total_value=0.0,
                )
            reason_stats[record.reason].count += 1
            reason_stats[record.reason].quantity += record.quantity
            reason_stats[record.reason].total_value += record.quantity * record.original_price

        return ScrapLedgerResponse(
            year=request.year,
            month=request.month,
            total_records=total_records,
            total_quantity=total_quantity,
            total_value=total_value,
            breakdown=list(reason_stats.values()),
            filter_reason=request.reason,
        )

    def check_quality_alert(self, request: QualityAlertRequest) -> QualityAlertResponse:
        threshold = request.threshold or ReturnConstants.DEFAULT_QUALITY_ALERT_THRESHOLD
        now = datetime.now()
        period_start = now - timedelta(days=request.period_days)

        scrap_records = self._data_store.list_scrap_records()

        period_scrap_quality = sum(
            r.quantity for r in scrap_records
            if r.created_at >= period_start and r.reason == DisposalReason.QUALITY_ISSUE
        )

        returns = self._data_store.list_return_orders()
        period_total = sum(
            item.quantity for r in returns
            if r.created_at >= period_start
            for item in r.items
        )

        actual_rate = 0.0
        if period_total > 0:
            actual_rate = period_scrap_quality / period_total

        alert_triggered = actual_rate >= threshold

        message = (
            f"质量问题退货率 {actual_rate:.2%}，{'超过' if alert_triggered else '未超过'}"
            f"阈值 {threshold:.2%}"
        )

        return QualityAlertResponse(
            supplier_id=request.supplier_id,
            period_days=request.period_days,
            threshold=threshold,
            actual_rate=actual_rate,
            alert_triggered=alert_triggered,
            message=message,
        )
