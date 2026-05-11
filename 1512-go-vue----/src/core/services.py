from datetime import datetime, date, timedelta
from typing import List, Optional, Dict, Tuple
from uuid import UUID
import random

from .models import (
    Storehouse, WarehouseArea, StockInRecord, StockOutRecord,
    TemperatureRecord, PestInspectionRecord, TodoItem,
    GrainVariety, get_shelf_life_days
)
from .repository import uow


class BusinessError(Exception):
    pass


class StorehouseService:
    def create(self, name: str, location: str) -> Storehouse:
        storehouse = Storehouse(name=name, location=location)
        return uow.storehouses.add(storehouse)

    def list(self) -> List[Storehouse]:
        return uow.storehouses.list()

    def get(self, storehouse_id: UUID) -> Optional[Storehouse]:
        return uow.storehouses.get(storehouse_id)


class WarehouseAreaService:
    def create(
        self,
        storehouse_id: UUID,
        name: str,
        area_type: str,
        capacity_kg: float,
        min_temp: float = 10.0,
        max_temp: float = 25.0,
    ) -> WarehouseArea:
        if not uow.storehouses.get(storehouse_id):
            raise BusinessError("仓库不存在")
        if capacity_kg <= 0:
            raise BusinessError("容量必须大于0")
        area = WarehouseArea(
            storehouse_id=storehouse_id,
            name=name,
            area_type=area_type,
            capacity_kg=capacity_kg,
            min_temp=min_temp,
            max_temp=max_temp,
        )
        return uow.areas.add(area)

    def list(self, storehouse_id: Optional[UUID] = None) -> List[WarehouseArea]:
        if storehouse_id:
            return uow.areas.list_by_storehouse(storehouse_id)
        return uow.areas.list()

    def get(self, area_id: UUID) -> Optional[WarehouseArea]:
        return uow.areas.get(area_id)


class StockService:
    def _generate_batch_no(self) -> str:
        date_str = datetime.now().strftime("%Y%m%d")
        random_str = ''.join([str(random.randint(0, 9)) for _ in range(6)])
        return f"BATCH-{date_str}-{random_str}"

    def stock_in(
        self,
        area_id: UUID,
        variety: str,
        quantity_kg: float,
        source: str,
        in_date: Optional[date] = None,
    ) -> StockInRecord:
        if quantity_kg <= 0:
            raise BusinessError("入库数量必须大于0")

        area = uow.areas.get(area_id)
        if not area:
            raise BusinessError("库区不存在")

        available = area.capacity_kg - area.current_used_kg
        if quantity_kg > available:
            raise BusinessError(f"库区容量不足，可用: {available}kg")

        variety_enum = GrainVariety(variety)
        shelf_life_days = get_shelf_life_days(variety_enum)
        actual_in_date = in_date or date.today()
        expiry_date = actual_in_date + timedelta(days=shelf_life_days)

        stock_in = StockInRecord(
            area_id=area_id,
            variety=variety_enum,
            quantity_kg=quantity_kg,
            remaining_kg=quantity_kg,
            source=source,
            batch_no=self._generate_batch_no(),
            in_date=actual_in_date,
            expiry_date=expiry_date,
        )

        area.current_used_kg += quantity_kg
        uow.areas.update(area)

        return uow.stock_ins.add(stock_in)

    def get_fifo_recommendation(
        self, area_id: UUID, variety: str, quantity_kg: float
    ) -> List[StockInRecord]:
        return uow.stock_ins.list_for_fifo(area_id, variety)

    def stock_out(
        self,
        area_id: UUID,
        variety: str,
        quantity_kg: float,
        destination: str,
        purpose: str,
        out_date: Optional[date] = None,
    ) -> List[StockOutRecord]:
        if quantity_kg <= 0:
            raise BusinessError("出库数量必须大于0")

        area = uow.areas.get(area_id)
        if not area:
            raise BusinessError("库区不存在")

        fifo_batches = uow.stock_ins.list_for_fifo(area_id, variety)
        total_available = sum(r.remaining_kg for r in fifo_batches)

        if total_available < quantity_kg:
            raise BusinessError(f"库存不足，可用: {total_available}kg")

        remaining_to_out = quantity_kg
        out_records: List[StockOutRecord] = []

        for batch in fifo_batches:
            if remaining_to_out <= 0:
                break

            take_amount = min(batch.remaining_kg, remaining_to_out)

            out_record = StockOutRecord(
                stock_in_id=batch.id,
                area_id=area_id,
                variety=GrainVariety(variety),
                quantity_kg=take_amount,
                destination=destination,
                purpose=purpose,
                out_date=out_date or date.today(),
            )
            out_records.append(uow.stock_outs.add(out_record))

            batch.remaining_kg -= take_amount
            uow.stock_ins.update(batch)

            remaining_to_out -= take_amount

        area.current_used_kg -= quantity_kg
        uow.areas.update(area)

        return out_records

    def list_stock_ins(self, area_id: Optional[UUID] = None) -> List[StockInRecord]:
        if area_id:
            return uow.stock_ins.list_by_area(area_id)
        return uow.stock_ins.list()

    def list_stock_outs(self, area_id: Optional[UUID] = None) -> List[StockOutRecord]:
        if area_id:
            return uow.stock_outs.list_by_area(area_id)
        return uow.stock_outs.list()

    def get_expiring_soon(self, days: int = 30) -> List[StockInRecord]:
        today = date.today()
        cutoff = today + timedelta(days=days)
        expiring = []
        for record in uow.stock_ins.list_active():
            if record.expiry_date <= cutoff:
                expiring.append(record)
        return sorted(expiring, key=lambda x: x.expiry_date)


class TemperatureService:
    def report_temperature(
        self, area_id: UUID, temperature: float, record_time: Optional[datetime] = None
    ) -> TemperatureRecord:
        area = uow.areas.get(area_id)
        if not area:
            raise BusinessError("库区不存在")

        actual_time = record_time or datetime.utcnow()

        existing = uow.temperatures.get_by_area_and_minute(area_id, actual_time)
        if existing:
            uow.temperatures.delete(existing.id)

        is_alarm = temperature < area.min_temp or temperature > area.max_temp

        record = TemperatureRecord(
            area_id=area_id,
            temperature=temperature,
            record_time=actual_time,
            is_alarm=is_alarm,
        )
        return uow.temperatures.add(record)

    def list_records(self, area_id: Optional[UUID] = None) -> List[TemperatureRecord]:
        if area_id:
            return uow.temperatures.list_by_area(area_id)
        return uow.temperatures.list()

    def get_alarming_records(self) -> List[TemperatureRecord]:
        return [r for r in uow.temperatures.list() if r.is_alarm]


class PestService:
    PEST_THRESHOLD = 10.0

    def record_inspection(
        self,
        area_id: UUID,
        pest_count_per_kg: float,
        variety: Optional[str] = None,
        inspection_date: Optional[date] = None,
        notes: Optional[str] = None,
    ) -> PestInspectionRecord:
        if pest_count_per_kg < 0:
            raise BusinessError("虫害数量不能为负")

        if not uow.areas.get(area_id):
            raise BusinessError("库区不存在")

        variety_enum = GrainVariety(variety) if variety else None

        inspection = PestInspectionRecord(
            area_id=area_id,
            variety=variety_enum,
            pest_count_per_kg=pest_count_per_kg,
            inspection_date=inspection_date or date.today(),
            notes=notes,
        )
        inspection = uow.pest_inspections.add(inspection)

        if pest_count_per_kg >= self.PEST_THRESHOLD:
            todo = TodoItem(
                area_id=area_id,
                title="虫害密度超标处理",
                description=f"检测到虫害密度 {pest_count_per_kg} 头/公斤，超过阈值 {self.PEST_THRESHOLD} 头/公斤，请立即处理。",
                source_type="pest_inspection",
                source_id=inspection.id,
            )
            uow.todos.add(todo)

        return inspection

    def list_inspections(self, area_id: Optional[UUID] = None) -> List[PestInspectionRecord]:
        if area_id:
            return uow.pest_inspections.list_by_area(area_id)
        return uow.pest_inspections.list()


class TodoService:
    def list_todos(self, area_id: Optional[UUID] = None, pending_only: bool = False) -> List[TodoItem]:
        if pending_only:
            todos = uow.todos.list_pending()
            if area_id:
                return [t for t in todos if t.area_id == area_id]
            return todos
        if area_id:
            return uow.todos.list_by_area(area_id)
        return uow.todos.list()

    def complete_todo(self, todo_id: UUID) -> Optional[TodoItem]:
        todo = uow.todos.get(todo_id)
        if todo:
            todo.is_completed = True
            return uow.todos.update(todo)
        return None

    def get(self, todo_id: UUID) -> Optional[TodoItem]:
        return uow.todos.get(todo_id)


class MetricsService:
    def get_total_stock(self) -> float:
        return sum(r.remaining_kg for r in uow.stock_ins.list_active())

    def get_variety_distribution(self) -> Dict[str, float]:
        distribution: Dict[str, float] = {}
        for record in uow.stock_ins.list_active():
            key = record.variety.value
            distribution[key] = distribution.get(key, 0) + record.remaining_kg
        return distribution

    def get_variety_percentage(self) -> Dict[str, float]:
        total = self.get_total_stock()
        distribution = self.get_variety_distribution()
        if total == 0:
            return {k: 0.0 for k in distribution}
        return {k: round(v / total * 100, 2) for k, v in distribution.items()}

    def get_pest_detection_rate(self) -> float:
        inspections = uow.pest_inspections.list()
        if not inspections:
            return 0.0
        detected = sum(1 for i in inspections if i.pest_count_per_kg > 0)
        return round(detected / len(inspections) * 100, 2)

    def get_temperature_compliance_rate(self) -> float:
        records = uow.temperatures.list()
        if not records:
            return 100.0
        compliant = sum(1 for r in records if not r.is_alarm)
        return round(compliant / len(records) * 100, 2)

    def get_all_metrics(self) -> Dict:
        return {
            "total_stock_kg": self.get_total_stock(),
            "variety_distribution_kg": self.get_variety_distribution(),
            "variety_percentage": self.get_variety_percentage(),
            "pest_detection_rate": self.get_pest_detection_rate(),
            "temperature_compliance_rate": self.get_temperature_compliance_rate(),
            "pending_todos_count": len(uow.todos.list_pending()),
            "expiring_soon_count": len(StockService().get_expiring_soon(days=30)),
        }


storehouse_service = StorehouseService()
area_service = WarehouseAreaService()
stock_service = StockService()
temperature_service = TemperatureService()
pest_service = PestService()
todo_service = TodoService()
metrics_service = MetricsService()
