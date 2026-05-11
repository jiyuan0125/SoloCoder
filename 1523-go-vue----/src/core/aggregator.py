from datetime import date, datetime, time, timedelta
from typing import List

from .models import (
    CrushingRecord,
    DailyMetrics,
    Equipment,
    EquipmentUtilization,
    FlotationRecord,
)


class MetricsAggregator:
    def __init__(
        self,
        crushing_records: List[CrushingRecord],
        flotation_records: List[FlotationRecord],
        equipment_list: List[Equipment],
    ):
        self.crushing_records = crushing_records
        self.flotation_records = flotation_records
        self.equipment_list = equipment_list

    def _get_records_by_date(
        self, target_date: date
    ) -> tuple[List[CrushingRecord], List[FlotationRecord]]:
        crushing_for_date = [
            r for r in self.crushing_records if r.record_date == target_date
        ]
        flotation_for_date = [
            r for r in self.flotation_records if r.date == target_date
        ]
        return crushing_for_date, flotation_for_date

    def calculate_daily_throughput(
        self, crushing_records: List[CrushingRecord]
    ) -> float:
        return sum(r.throughput for r in crushing_records)

    def calculate_average_recovery(
        self, flotation_records: List[FlotationRecord]
    ) -> float:
        if not flotation_records:
            return 0.0
        total_recovery = sum(r.recovery for r in flotation_records)
        return total_recovery / len(flotation_records)

    def calculate_equipment_utilization_for_day(
        self, target_date: date
    ) -> float:
        if not self.equipment_list:
            return 0.0

        total_utilization = 0.0
        counted_equipment = 0

        for equipment in self.equipment_list:
            utilization = self._calculate_single_equipment_utilization(
                equipment, target_date
            )
            if utilization is not None:
                total_utilization += utilization
                counted_equipment += 1

        if counted_equipment == 0:
            return 0.0
        return total_utilization / counted_equipment

    def _calculate_single_equipment_utilization(
        self, equipment: Equipment, target_date: date
    ) -> float:
        start_of_day = datetime.combine(target_date, time.min)
        end_of_day = datetime.combine(target_date, time.max)
        total_day_hours = 24.0

        if equipment.total_running_hours > 0:
            return min(equipment.total_running_hours / total_day_hours, 1.0)

        return 0.0

    def get_daily_metrics(self, target_date: date) -> DailyMetrics:
        crushing, flotation = self._get_records_by_date(target_date)
        throughput = self.calculate_daily_throughput(crushing)
        avg_recovery = self.calculate_average_recovery(flotation)
        utilization = self.calculate_equipment_utilization_for_day(target_date)

        return DailyMetrics(
            date=target_date,
            total_throughput=throughput,
            average_recovery=avg_recovery,
            equipment_utilization=utilization,
        )

    def get_equipment_utilization_list(
        self, target_date: date
    ) -> List[EquipmentUtilization]:
        utilizations = []
        for equipment in self.equipment_list:
            rate = self._calculate_single_equipment_utilization(
                equipment, target_date
            )
            utilizations.append(
                EquipmentUtilization(
                    equipment_id=equipment.id,
                    equipment_name=equipment.name,
                    utilization_rate=rate,
                    running_hours=equipment.total_running_hours,
                )
            )
        return utilizations

    def get_date_range_metrics(
        self, start_date: date, end_date: date
    ) -> List[DailyMetrics]:
        metrics_list = []
        current_date = start_date
        while current_date <= end_date:
            metrics = self.get_daily_metrics(current_date)
            metrics_list.append(metrics)
            current_date += timedelta(days=1)
        return metrics_list
