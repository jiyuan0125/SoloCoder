from decimal import Decimal
from datetime import datetime
from typing import Optional, List
from threading import Lock
from collections import defaultdict

from shared.models import (
    TransportMode,
    CalculationResult,
    MonthlyStatistics,
)


class CalculationRecord:
    def __init__(
        self,
        result: CalculationResult,
        transport_mode: TransportMode,
        customer_id: Optional[str],
    ) -> None:
        self.result = result
        self.transport_mode = transport_mode
        self.customer_id = customer_id
        self.created_at = datetime.now()

    def get_year_month(self) -> str:
        return self.created_at.strftime("%Y-%m")


class DataStore:
    def __init__(self) -> None:
        self._records: List[CalculationRecord] = []
        self._customer_cumulative: dict[str, Decimal] = defaultdict(lambda: Decimal("0"))
        self._lock: Lock = Lock()

    def add_record(
        self,
        result: CalculationResult,
        transport_mode: TransportMode,
        customer_id: Optional[str],
    ) -> None:
        with self._lock:
            record = CalculationRecord(result, transport_mode, customer_id)
            self._records.append(record)
            
            if customer_id:
                self._customer_cumulative[customer_id] += result.total_final_amount

    def get_customer_cumulative(self, customer_id: Optional[str]) -> Decimal:
        if customer_id is None:
            return Decimal("0")
        
        with self._lock:
            return self._customer_cumulative.get(customer_id, Decimal("0"))

    def get_monthly_statistics(self, year_month: str) -> Optional[MonthlyStatistics]:
        total_packages = 0
        total_base_freight = Decimal("0")
        total_surcharge = Decimal("0")
        total_insurance_fee = Decimal("0")
        total_discount = Decimal("0")
        total_final_amount = Decimal("0")
        
        mode_breakdown: dict[TransportMode, int] = defaultdict(int)
        
        with self._lock:
            found = False
            for record in self._records:
                if record.get_year_month() == year_month:
                    found = True
                    total_packages += len(record.result.package_details)
                    total_base_freight += record.result.total_base_freight
                    total_surcharge += record.result.total_surcharge
                    total_insurance_fee += record.result.total_insurance_fee
                    total_discount += record.result.total_discount
                    total_final_amount += record.result.total_final_amount
                    mode_breakdown[record.transport_mode] += 1
            
            if not found:
                return None
        
        return MonthlyStatistics(
            year_month=year_month,
            total_packages=total_packages,
            total_base_freight=total_base_freight,
            total_surcharge=total_surcharge,
            total_insurance_fee=total_insurance_fee,
            total_discount=total_discount,
            total_final_amount=total_final_amount,
            transport_mode_breakdown=dict(mode_breakdown),
        )

    def list_available_months(self) -> List[str]:
        with self._lock:
            months = set()
            for record in self._records:
                months.add(record.get_year_month())
        
        return sorted(list(months))

    def get_all_records_count(self) -> int:
        with self._lock:
            return len(self._records)

    def get_total_packages_count(self) -> int:
        total = 0
        with self._lock:
            for record in self._records:
                total += len(record.result.package_details)
        return total


data_store: DataStore = DataStore()
