from datetime import datetime
from sqlalchemy.orm import Session
from typing import List, Tuple
from app.models import QueueItem, Cabin, CabinStatus
from app.config import get_settings

settings = get_settings()


class QueueManager:
    def __init__(self, db: Session):
        self.db = db

    def get_total_queue_count(self) -> int:
        unserved_items = self.db.query(QueueItem).filter(
            QueueItem.is_served == False
        ).all()
        return sum(item.passenger_count for item in unserved_items)

    def get_upward_passengers(self) -> int:
        upward_cabins = self.db.query(Cabin).filter(
            Cabin.current_status == CabinStatus.UPWARD
        ).all()
        return sum(cabin.passenger_count for cabin in upward_cabins)

    def get_queue_stats(self) -> dict:
        total_queue = self.get_total_queue_count()
        upward_passengers = self.get_upward_passengers()
        effective_queue = total_queue + upward_passengers
        
        unserved_items = self.db.query(QueueItem).filter(
            QueueItem.is_served == False
        ).order_by(QueueItem.join_time).all()
        
        first_join_time = unserved_items[0].join_time if unserved_items else None
        avg_wait_seconds = 0
        if first_join_time:
            avg_wait_seconds = (datetime.utcnow() - first_join_time).total_seconds()
        
        return {
            "total_queue": total_queue,
            "upward_passengers": upward_passengers,
            "effective_queue": effective_queue,
            "first_join_time": first_join_time,
            "average_wait_seconds": avg_wait_seconds,
            "queue_groups": len(unserved_items)
        }

    def add_to_queue(self, passenger_count: int, average_weight: float = 70.0) -> QueueItem:
        item = QueueItem(
            passenger_count=passenger_count,
            average_weight=average_weight,
            join_time=datetime.utcnow()
        )
        self.db.add(item)
        self.db.commit()
        self.db.refresh(item)
        return item

    def get_next_queue_items(self, max_weight: float) -> Tuple[List[QueueItem], int, float]:
        unserved_items = self.db.query(QueueItem).filter(
            QueueItem.is_served == False
        ).order_by(QueueItem.join_time).all()
        
        selected_items = []
        total_passengers = 0
        total_weight = 0.0
        
        for item in unserved_items:
            item_weight = item.passenger_count * (item.average_weight + settings.default_luggage_weight)
            if total_weight + item_weight <= max_weight:
                selected_items.append(item)
                total_passengers += item.passenger_count
                total_weight += item_weight
            else:
                break
        
        return selected_items, total_passengers, total_weight

    def mark_items_served(self, items: List[QueueItem], cabin_id: int) -> None:
        now = datetime.utcnow()
        for item in items:
            item.is_served = True
            item.assigned_cabin_id = cabin_id
            item.served_time = now
        self.db.commit()
