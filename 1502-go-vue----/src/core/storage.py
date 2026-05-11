from typing import Dict, List, Optional, Any
from .models import (
    Pond, WaterQualityThreshold, WaterQualityRecord,
    FeedingPlan, FeedingRecord, HarvestRecord
)


class InMemoryStorage:
    def __init__(self):
        self._ponds: Dict[int, Pond] = {}
        self._pond_next_id = 1
        
        self._thresholds: Dict[int, WaterQualityThreshold] = {}
        self._threshold_next_id = 1
        
        self._quality_records: Dict[int, WaterQualityRecord] = {}
        self._quality_next_id = 1
        
        self._feeding_plans: Dict[int, FeedingPlan] = {}
        self._feeding_plan_next_id = 1
        
        self._feeding_records: Dict[int, FeedingRecord] = {}
        self._feeding_record_next_id = 1
        
        self._harvest_records: Dict[int, HarvestRecord] = {}
        self._harvest_next_id = 1

    # Ponds
    def add_pond(self, pond: Pond) -> Pond:
        if pond.id is None:
            pond.id = self._pond_next_id
            self._pond_next_id += 1
        self._ponds[pond.id] = pond
        return pond

    def get_pond(self, pond_id: int) -> Optional[Pond]:
        return self._ponds.get(pond_id)

    def list_ponds(self) -> List[Pond]:
        return list(self._ponds.values())

    def update_pond(self, pond: Pond) -> Pond:
        self._ponds[pond.id] = pond
        return pond

    def delete_pond(self, pond_id: int) -> bool:
        if pond_id in self._ponds:
            del self._ponds[pond_id]
            return True
        return False

    # Thresholds
    def add_threshold(self, threshold: WaterQualityThreshold) -> WaterQualityThreshold:
        if threshold.id is None:
            threshold.id = self._threshold_next_id
            self._threshold_next_id += 1
        self._thresholds[threshold.id] = threshold
        return threshold

    def get_threshold_by_pond_and_param(self, pond_id: int, parameter: str) -> Optional[WaterQualityThreshold]:
        for t in self._thresholds.values():
            if t.pond_id == pond_id and t.parameter == parameter:
                return t
        return None

    def list_thresholds(self, pond_id: Optional[int] = None) -> List[WaterQualityThreshold]:
        thresholds = list(self._thresholds.values())
        if pond_id is not None:
            thresholds = [t for t in thresholds if t.pond_id == pond_id]
        return thresholds

    # Water Quality Records
    def add_quality_record(self, record: WaterQualityRecord) -> WaterQualityRecord:
        if record.id is None:
            record.id = self._quality_next_id
            self._quality_next_id += 1
        self._quality_records[record.id] = record
        return record

    def list_quality_records(self, pond_id: Optional[int] = None,
                            start_time: Optional[Any] = None,
                            end_time: Optional[Any] = None) -> List[WaterQualityRecord]:
        records = list(self._quality_records.values())
        if pond_id is not None:
            records = [r for r in records if r.pond_id == pond_id]
        if start_time is not None:
            records = [r for r in records if r.recorded_at >= start_time]
        if end_time is not None:
            records = [r for r in records if r.recorded_at <= end_time]
        return records

    # Feeding Plans
    def add_feeding_plan(self, plan: FeedingPlan) -> FeedingPlan:
        if plan.id is None:
            plan.id = self._feeding_plan_next_id
            self._feeding_plan_next_id += 1
        self._feeding_plans[plan.id] = plan
        return plan

    def get_feeding_plan(self, plan_id: int) -> Optional[FeedingPlan]:
        return self._feeding_plans.get(plan_id)

    def list_feeding_plans(self, pond_id: Optional[int] = None) -> List[FeedingPlan]:
        plans = list(self._feeding_plans.values())
        if pond_id is not None:
            plans = [p for p in plans if p.pond_id == pond_id]
        return plans

    def delete_feeding_plan(self, plan_id: int) -> bool:
        if plan_id in self._feeding_plans:
            del self._feeding_plans[plan_id]
            return True
        return False

    # Feeding Records
    def add_feeding_record(self, record: FeedingRecord) -> FeedingRecord:
        if record.id is None:
            record.id = self._feeding_record_next_id
            self._feeding_record_next_id += 1
        self._feeding_records[record.id] = record
        return record

    def list_feeding_records(self, pond_id: Optional[int] = None) -> List[FeedingRecord]:
        records = list(self._feeding_records.values())
        if pond_id is not None:
            records = [r for r in records if r.pond_id == pond_id]
        return records

    # Harvest Records
    def add_harvest_record(self, record: HarvestRecord) -> HarvestRecord:
        if record.id is None:
            record.id = self._harvest_next_id
            self._harvest_next_id += 1
        self._harvest_records[record.id] = record
        return record

    def list_harvest_records(self, pond_id: Optional[int] = None) -> List[HarvestRecord]:
        records = list(self._harvest_records.values())
        if pond_id is not None:
            records = [r for r in records if r.pond_id == pond_id]
        return records


storage = InMemoryStorage()
