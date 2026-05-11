from datetime import datetime, timedelta
from typing import List, Optional
from src.core.models.schemas import DailyStatistics, BatchStatus, ProductionBatch


class AggregationService:
    def aggregate_daily_statistics(
        self,
        date_str: str,
        batches: List[ProductionBatch]
    ) -> DailyStatistics:
        stats = DailyStatistics(date=date_str)

        stats.total_batches = len(batches)
        
        completed_or_abnormal = [
            b for b in batches
            if b.status in (BatchStatus.COMPLETED, BatchStatus.ABNORMAL)
        ]
        stats.completed_batches = len(completed_or_abnormal)
        
        abnormal_batches = [b for b in batches if b.is_abnormal]
        stats.abnormal_batches = len(abnormal_batches)
        
        if stats.completed_batches > 0:
            stats.abnormal_rate = round(
                (stats.abnormal_batches / stats.completed_batches) * 100, 2
            )

        return stats

    def _date_only_str(self, dt: datetime) -> str:
        return dt.strftime("%Y-%m-%d")

    def get_date_range(
        self,
        start_date: Optional[str] = None,
        end_date: Optional[str] = None
    ) -> List[str]:
        if end_date is None:
            end_dt = datetime.now()
        else:
            end_dt = datetime.strptime(end_date, "%Y-%m-%d")

        if start_date is None:
            start_dt = end_dt
        else:
            start_dt = datetime.strptime(start_date, "%Y-%m-%d")

        dates = []
        current = start_dt
        while current <= end_dt:
            dates.append(self._date_only_str(current))
            current += timedelta(days=1)

        return dates
