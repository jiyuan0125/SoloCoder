from sqlalchemy.orm import Session
from sqlalchemy import func
from datetime import datetime, timedelta
from typing import List, Optional, Dict, Any
from .models import (
    Warehouse,
    StorageArea,
    GrainBatch,
    InboundRecord,
    OutboundRecord,
    TemperatureRecord,
    PestInspection,
    Todo,
)
from .config import (
    calculate_expiry_date,
    get_minute_key,
    is_pest_density_exceeded,
)


class WarehouseService:
    @staticmethod
    def create(db: Session, name: str, location: Optional[str] = None) -> Warehouse:
        warehouse = Warehouse(name=name, location=location)
        db.add(warehouse)
        db.commit()
        db.refresh(warehouse)
        return warehouse

    @staticmethod
    def get_all(db: Session) -> List[Warehouse]:
        return db.query(Warehouse).all()

    @staticmethod
    def get_by_id(db: Session, warehouse_id: int) -> Optional[Warehouse]:
        return db.query(Warehouse).filter(Warehouse.id == warehouse_id).first()

    @staticmethod
    def delete(db: Session, warehouse_id: int) -> bool:
        warehouse = WarehouseService.get_by_id(db, warehouse_id)
        if warehouse:
            db.delete(warehouse)
            db.commit()
            return True
        return False


class StorageAreaService:
    @staticmethod
    def create(
        db: Session,
        warehouse_id: int,
        name: str,
        area_type: str,
        capacity_kg: float,
        min_temp: float = 10.0,
        max_temp: float = 25.0,
    ) -> StorageArea:
        warehouse = WarehouseService.get_by_id(db, warehouse_id)
        if not warehouse:
            raise ValueError(f"仓库 {warehouse_id} 不存在")

        storage_area = StorageArea(
            warehouse_id=warehouse_id,
            name=name,
            area_type=area_type,
            capacity_kg=capacity_kg,
            min_temp=min_temp,
            max_temp=max_temp,
        )
        db.add(storage_area)
        db.commit()
        db.refresh(storage_area)
        return storage_area

    @staticmethod
    def get_all(db: Session) -> List[StorageArea]:
        return db.query(StorageArea).all()

    @staticmethod
    def get_by_id(db: Session, area_id: int) -> Optional[StorageArea]:
        return db.query(StorageArea).filter(StorageArea.id == area_id).first()

    @staticmethod
    def get_by_warehouse(db: Session, warehouse_id: int) -> List[StorageArea]:
        return db.query(StorageArea).filter(StorageArea.warehouse_id == warehouse_id).all()

    @staticmethod
    def get_current_usage(db: Session, area_id: int) -> float:
        result = db.query(func.sum(GrainBatch.remaining_kg)).filter(
            GrainBatch.storage_area_id == area_id
        ).scalar()
        return result or 0.0

    @staticmethod
    def get_available_capacity(db: Session, area_id: int) -> float:
        area = StorageAreaService.get_by_id(db, area_id)
        if not area:
            return 0.0
        used = StorageAreaService.get_current_usage(db, area_id)
        return area.capacity_kg - used

    @staticmethod
    def delete(db: Session, area_id: int) -> bool:
        area = StorageAreaService.get_by_id(db, area_id)
        if area:
            db.delete(area)
            db.commit()
            return True
        return False


class InboundService:
    @staticmethod
    def create(
        db: Session,
        storage_area_id: int,
        grain_variety: str,
        quantity_kg: float,
        source: str,
        inbound_date: Optional[datetime] = None,
    ) -> InboundRecord:
        area = StorageAreaService.get_by_id(db, storage_area_id)
        if not area:
            raise ValueError(f"库区 {storage_area_id} 不存在")

        available = StorageAreaService.get_available_capacity(db, storage_area_id)
        if quantity_kg > available:
            raise ValueError(f"库区容量不足。可用: {available} kg, 请求: {quantity_kg} kg")

        if inbound_date is None:
            inbound_date = datetime.utcnow()

        expiry_date = calculate_expiry_date(inbound_date, grain_variety)

        batch = GrainBatch(
            storage_area_id=storage_area_id,
            grain_variety=grain_variety,
            quantity_kg=quantity_kg,
            remaining_kg=quantity_kg,
            source=source,
            inbound_date=inbound_date,
            expiry_date=expiry_date,
        )
        db.add(batch)
        db.flush()

        record = InboundRecord(
            batch_id=batch.id,
            storage_area_id=storage_area_id,
            grain_variety=grain_variety,
            quantity_kg=quantity_kg,
            source=source,
        )
        db.add(record)
        db.commit()
        db.refresh(record)
        return record

    @staticmethod
    def get_all(db: Session) -> List[InboundRecord]:
        return db.query(InboundRecord).order_by(InboundRecord.created_at.desc()).all()

    @staticmethod
    def get_by_area(db: Session, area_id: int) -> List[InboundRecord]:
        return db.query(InboundRecord).filter(
            InboundRecord.storage_area_id == area_id
        ).order_by(InboundRecord.created_at.desc()).all()


class OutboundService:
    @staticmethod
    def recommend_batches(
        db: Session,
        storage_area_id: int,
        grain_variety: str,
        quantity_kg: float,
    ) -> List[Dict[str, Any]]:
        batches = db.query(GrainBatch).filter(
            GrainBatch.storage_area_id == storage_area_id,
            GrainBatch.grain_variety == grain_variety,
            GrainBatch.remaining_kg > 0,
        ).order_by(GrainBatch.inbound_date.asc()).all()

        total_available = sum(b.remaining_kg for b in batches)
        if total_available < quantity_kg:
            raise ValueError(f"库存不足。可用: {total_available} kg, 请求: {quantity_kg} kg")

        recommendations = []
        remaining_needed = quantity_kg

        for batch in batches:
            if remaining_needed <= 0:
                break
            take = min(batch.remaining_kg, remaining_needed)
            recommendations.append({
                "batch_id": batch.id,
                "inbound_date": batch.inbound_date,
                "remaining_kg": batch.remaining_kg,
                "take_kg": take,
            })
            remaining_needed -= take

        return recommendations

    @staticmethod
    def create(
        db: Session,
        storage_area_id: int,
        grain_variety: str,
        quantity_kg: float,
        destination: str,
        purpose: str,
    ) -> List[OutboundRecord]:
        area = StorageAreaService.get_by_id(db, storage_area_id)
        if not area:
            raise ValueError(f"库区 {storage_area_id} 不存在")

        recommendations = OutboundService.recommend_batches(
            db, storage_area_id, grain_variety, quantity_kg
        )

        records = []
        for rec in recommendations:
            batch = db.query(GrainBatch).filter(GrainBatch.id == rec["batch_id"]).first()
            if not batch:
                continue

            if rec["take_kg"] > batch.remaining_kg:
                raise ValueError(f"批次 {batch.id} 剩余数量不足")

            batch.remaining_kg -= rec["take_kg"]

            record = OutboundRecord(
                batch_id=batch.id,
                storage_area_id=storage_area_id,
                grain_variety=grain_variety,
                quantity_kg=rec["take_kg"],
                destination=destination,
                purpose=purpose,
            )
            db.add(record)
            records.append(record)

        db.commit()
        for record in records:
            db.refresh(record)
        return records

    @staticmethod
    def get_all(db: Session) -> List[OutboundRecord]:
        return db.query(OutboundRecord).order_by(OutboundRecord.created_at.desc()).all()

    @staticmethod
    def get_by_area(db: Session, area_id: int) -> List[OutboundRecord]:
        return db.query(OutboundRecord).filter(
            OutboundRecord.storage_area_id == area_id
        ).order_by(OutboundRecord.created_at.desc()).all()


class TemperatureService:
    @staticmethod
    def record(
        db: Session,
        storage_area_id: int,
        temperature: float,
        recorded_at: Optional[datetime] = None,
    ) -> TemperatureRecord:
        area = StorageAreaService.get_by_id(db, storage_area_id)
        if not area:
            raise ValueError(f"库区 {storage_area_id} 不存在")

        if recorded_at is None:
            recorded_at = datetime.utcnow()

        minute_key = get_minute_key(recorded_at)

        existing = db.query(TemperatureRecord).filter(
            TemperatureRecord.storage_area_id == storage_area_id,
            TemperatureRecord.minute_key == minute_key,
        ).first()

        is_alert = temperature < area.min_temp or temperature > area.max_temp

        if existing:
            existing.temperature = temperature
            existing.is_alert = is_alert
            existing.recorded_at = recorded_at
            db.commit()
            db.refresh(existing)
            return existing
        else:
            record = TemperatureRecord(
                storage_area_id=storage_area_id,
                temperature=temperature,
                is_alert=is_alert,
                recorded_at=recorded_at,
                minute_key=minute_key,
            )
            db.add(record)
            db.commit()
            db.refresh(record)
            return record

    @staticmethod
    def get_all(db: Session) -> List[TemperatureRecord]:
        return db.query(TemperatureRecord).order_by(TemperatureRecord.recorded_at.desc()).all()

    @staticmethod
    def get_by_area(db: Session, area_id: int, limit: int = 100) -> List[TemperatureRecord]:
        return db.query(TemperatureRecord).filter(
            TemperatureRecord.storage_area_id == area_id
        ).order_by(TemperatureRecord.recorded_at.desc()).limit(limit).all()

    @staticmethod
    def get_alerts(db: Session) -> List[TemperatureRecord]:
        return db.query(TemperatureRecord).filter(
            TemperatureRecord.is_alert == True
        ).order_by(TemperatureRecord.recorded_at.desc()).all()


class PestService:
    @staticmethod
    def create(
        db: Session,
        storage_area_id: int,
        grain_variety: str,
        pest_count_per_kg: float,
        inspection_result: str,
    ) -> PestInspection:
        area = StorageAreaService.get_by_id(db, storage_area_id)
        if not area:
            raise ValueError(f"库区 {storage_area_id} 不存在")

        inspection = PestInspection(
            storage_area_id=storage_area_id,
            grain_variety=grain_variety,
            pest_count_per_kg=pest_count_per_kg,
            inspection_result=inspection_result,
        )
        db.add(inspection)
        db.flush()

        if is_pest_density_exceeded(pest_count_per_kg):
            todo = Todo(
                title=f"虫害告警 - 库区 {area.name}",
                description=f"品种: {grain_variety}, 虫害密度: {pest_count_per_kg} 头/公斤, 超过预警阈值 10 头/公斤",
            )
            db.add(todo)

        db.commit()
        db.refresh(inspection)
        return inspection

    @staticmethod
    def get_all(db: Session) -> List[PestInspection]:
        return db.query(PestInspection).order_by(PestInspection.created_at.desc()).all()

    @staticmethod
    def get_by_area(db: Session, area_id: int) -> List[PestInspection]:
        return db.query(PestInspection).filter(
            PestInspection.storage_area_id == area_id
        ).order_by(PestInspection.created_at.desc()).all()


class MetricsService:
    @staticmethod
    def get_total_inventory(db: Session) -> float:
        result = db.query(func.sum(GrainBatch.remaining_kg)).filter(
            GrainBatch.remaining_kg > 0
        ).scalar()
        return result or 0.0

    @staticmethod
    def get_variety_percentage(db: Session) -> Dict[str, Dict[str, float]]:
        total = MetricsService.get_total_inventory(db)
        if total == 0:
            return {}

        results = db.query(
            GrainBatch.grain_variety,
            func.sum(GrainBatch.remaining_kg).label("quantity")
        ).filter(
            GrainBatch.remaining_kg > 0
        ).group_by(GrainBatch.grain_variety).all()

        return {
            r.grain_variety: {
                "quantity": r.quantity,
                "percentage": (r.quantity / total * 100) if total > 0 else 0
            }
            for r in results
        }

    @staticmethod
    def get_pest_detection_rate(db: Session, days: int = 30) -> float:
        since = datetime.utcnow() - timedelta(days=days)
        total = db.query(func.count(PestInspection.id)).filter(
            PestInspection.created_at >= since
        ).scalar() or 0

        if total == 0:
            return 0.0

        exceeded = db.query(func.count(PestInspection.id)).filter(
            PestInspection.created_at >= since,
            PestInspection.pest_count_per_kg > 10
        ).scalar() or 0

        return (exceeded / total * 100) if total > 0 else 0

    @staticmethod
    def get_temperature_compliance_rate(db: Session, days: int = 30) -> float:
        since = datetime.utcnow() - timedelta(days=days)
        total = db.query(func.count(TemperatureRecord.id)).filter(
            TemperatureRecord.recorded_at >= since
        ).scalar() or 0

        if total == 0:
            return 100.0

        alerts = db.query(func.count(TemperatureRecord.id)).filter(
            TemperatureRecord.recorded_at >= since,
            TemperatureRecord.is_alert == True
        ).scalar() or 0

        return ((total - alerts) / total * 100) if total > 0 else 100

    @staticmethod
    def get_expiring_batches(db: Session, days: int = 30) -> List[GrainBatch]:
        threshold = datetime.utcnow() + timedelta(days=days)
        return db.query(GrainBatch).filter(
            GrainBatch.remaining_kg > 0,
            GrainBatch.expiry_date <= threshold,
            GrainBatch.is_expired == False,
        ).order_by(GrainBatch.expiry_date.asc()).all()

    @staticmethod
    def get_todos(db: Session, include_completed: bool = False) -> List[Todo]:
        query = db.query(Todo)
        if not include_completed:
            query = query.filter(Todo.is_completed == False)
        return query.order_by(Todo.created_at.desc()).all()

    @staticmethod
    def complete_todo(db: Session, todo_id: int) -> bool:
        todo = db.query(Todo).filter(Todo.id == todo_id).first()
        if todo:
            todo.is_completed = True
            db.commit()
            return True
        return False

    @staticmethod
    def get_all_metrics(db: Session) -> Dict[str, Any]:
        return {
            "total_inventory_kg": MetricsService.get_total_inventory(db),
            "variety_percentage": MetricsService.get_variety_percentage(db),
            "pest_detection_rate_30d": MetricsService.get_pest_detection_rate(db),
            "temperature_compliance_rate_30d": MetricsService.get_temperature_compliance_rate(db),
            "expiring_batches_count": len(MetricsService.get_expiring_batches(db)),
            "pending_todos_count": len(MetricsService.get_todos(db, include_completed=False)),
        }
