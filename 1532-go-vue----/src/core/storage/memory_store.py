from datetime import datetime
from typing import Dict, List, Optional
from collections import defaultdict

from src.core.models.schemas import (
    Reactor, ReactorCreate, ReactorUpdate, SensorReading, Recipe, RecipeCreate,
    FeedingRecord, ProductionBatch, DailyStatistics, AlarmRecord,
    BatchStatus
)


class MemoryStore:
    def __init__(self):
        self._reactors: Dict[str, Reactor] = {}
        self._recipes: Dict[str, Recipe] = {}
        self._sensor_readings: List[SensorReading] = []
        self._feeding_records: Dict[str, List[FeedingRecord]] = defaultdict(list)
        self._batches: Dict[str, ProductionBatch] = {}
        self._daily_stats: Dict[str, DailyStatistics] = {}
        self._alarm_records: List[AlarmRecord] = []
        self._running_batches_by_reactor: Dict[str, str] = {}

    def create_reactor(self, data: ReactorCreate) -> Reactor:
        from uuid import uuid4
        reactor = Reactor(
            id=str(uuid4()),
            name=data.name,
            description=data.description,
            design_temperature_max=data.design_temperature_max,
            design_pressure_max=data.design_pressure_max
        )
        self._reactors[reactor.id] = reactor
        return reactor

    def get_reactor(self, reactor_id: str) -> Optional[Reactor]:
        return self._reactors.get(reactor_id)

    def list_reactors(self) -> List[Reactor]:
        return list(self._reactors.values())

    def update_reactor(self, reactor_id: str, data: ReactorUpdate) -> Optional[Reactor]:
        reactor = self._reactors.get(reactor_id)
        if not reactor:
            return None
        update_data = data.dict(exclude_unset=True)
        for key, value in update_data.items():
            setattr(reactor, key, value)
        reactor.updated_at = datetime.now()
        return reactor

    def delete_reactor(self, reactor_id: str) -> bool:
        if reactor_id in self._reactors:
            del self._reactors[reactor_id]
            return True
        return False

    def create_recipe(self, data: RecipeCreate) -> Recipe:
        from uuid import uuid4
        recipe = Recipe(
            id=str(uuid4()),
            name=data.name,
            description=data.description,
            materials=data.materials,
            time_window_seconds=data.time_window_seconds
        )
        self._recipes[recipe.id] = recipe
        return recipe

    def get_recipe(self, recipe_id: str) -> Optional[Recipe]:
        return self._recipes.get(recipe_id)

    def list_recipes(self) -> List[Recipe]:
        return list(self._recipes.values())

    def add_sensor_reading(self, reading: SensorReading) -> SensorReading:
        from uuid import uuid4
        if not reading.id:
            reading.id = str(uuid4())
        self._sensor_readings.append(reading)
        batch_id = self._running_batches_by_reactor.get(reading.reactor_id)
        if batch_id:
            batch = self._batches.get(batch_id)
            if batch and batch.status == BatchStatus.RUNNING:
                batch.sensor_readings.append(reading)
        return reading

    def list_sensor_readings(self, reactor_id: Optional[str] = None,
                           start_time: Optional[datetime] = None,
                           end_time: Optional[datetime] = None) -> List[SensorReading]:
        result = self._sensor_readings
        if reactor_id:
            result = [r for r in result if r.reactor_id == reactor_id]
        if start_time:
            result = [r for r in result if r.timestamp >= start_time]
        if end_time:
            result = [r for r in result if r.timestamp <= end_time]
        return result

    def create_batch(self, data) -> ProductionBatch:
        from uuid import uuid4
        batch = ProductionBatch(
            id=str(uuid4()),
            reactor_id=data.reactor_id,
            recipe_id=data.recipe_id
        )
        self._batches[batch.id] = batch
        self._running_batches_by_reactor[batch.reactor_id] = batch.id
        return batch

    def get_batch(self, batch_id: str) -> Optional[ProductionBatch]:
        return self._batches.get(batch_id)

    def list_batches(self, reactor_id: Optional[str] = None,
                     status: Optional[BatchStatus] = None,
                     start_date: Optional[datetime] = None,
                     end_date: Optional[datetime] = None) -> List[ProductionBatch]:
        result = list(self._batches.values())
        if reactor_id:
            result = [b for b in result if b.reactor_id == reactor_id]
        if status:
            result = [b for b in result if b.status == status]
        if start_date:
            result = [b for b in result if b.start_time >= start_date]
        if end_date:
            result = [b for b in result if b.start_time <= end_date]
        return result

    def update_batch(self, batch: ProductionBatch) -> ProductionBatch:
        self._batches[batch.id] = batch
        if batch.status != BatchStatus.RUNNING:
            if batch.reactor_id in self._running_batches_by_reactor:
                if self._running_batches_by_reactor[batch.reactor_id] == batch.id:
                    del self._running_batches_by_reactor[batch.reactor_id]
        return batch

    def add_feeding_record(self, record: FeedingRecord) -> FeedingRecord:
        from uuid import uuid4
        if not record.id:
            record.id = str(uuid4())
        self._feeding_records[record.batch_id].append(record)
        batch = self._batches.get(record.batch_id)
        if batch:
            batch.feeding_records.append(record)
        return record

    def list_feeding_records(self, batch_id: Optional[str] = None) -> List[FeedingRecord]:
        if batch_id:
            return self._feeding_records.get(batch_id, [])
        result = []
        for records in self._feeding_records.values():
            result.extend(records)
        return result

    def get_running_batch_for_reactor(self, reactor_id: str) -> Optional[ProductionBatch]:
        batch_id = self._running_batches_by_reactor.get(reactor_id)
        if batch_id:
            return self._batches.get(batch_id)
        return None

    def add_alarm_record(self, alarm: AlarmRecord) -> AlarmRecord:
        from uuid import uuid4
        if not alarm.id:
            alarm.id = str(uuid4())
        self._alarm_records.append(alarm)
        return alarm

    def list_alarm_records(self, reactor_id: Optional[str] = None,
                          start_time: Optional[datetime] = None,
                          end_time: Optional[datetime] = None,
                          is_high_risk: Optional[bool] = None) -> List[AlarmRecord]:
        result = list(self._alarm_records)
        if reactor_id:
            result = [a for a in result if a.reactor_id == reactor_id]
        if start_time:
            result = [a for a in result if a.timestamp >= start_time]
        if end_time:
            result = [a for a in result if a.timestamp <= end_time]
        if is_high_risk is not None:
            result = [a for a in result if a.is_high_risk == is_high_risk]
        return result

    def save_daily_statistics(self, stats: DailyStatistics) -> DailyStatistics:
        self._daily_stats[stats.date] = stats
        return stats

    def get_daily_statistics(self, date: str) -> Optional[DailyStatistics]:
        return self._daily_stats.get(date)

    def list_daily_statistics(self, start_date: Optional[str] = None,
                            end_date: Optional[str] = None) -> List[DailyStatistics]:
        result = list(self._daily_stats.values())
        if start_date:
            result = [s for s in result if s.date >= start_date]
        if end_date:
            result = [s for s in result if s.date <= end_date]
        return sorted(result, key=lambda x: x.date)


_store_instance = MemoryStore()


def get_store() -> MemoryStore:
    return _store_instance
