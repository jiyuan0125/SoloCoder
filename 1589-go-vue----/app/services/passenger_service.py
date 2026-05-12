from datetime import datetime, timedelta
from sqlalchemy.orm import Session
from app.models import GondolaCapacity, PassengerRecord, QueueData
from app.config import CHILD_HEIGHT_LIMIT, QUEUE_CAPACITY_MULTIPLIER


class PassengerService:
    def __init__(self, db: Session):
        self.db = db

    def get_active_capacity(self) -> int:
        capacity = self.db.query(GondolaCapacity).filter(
            GondolaCapacity.is_active == True
        ).order_by(GondolaCapacity.effective_date.desc()).first()
        return capacity.total_capacity if capacity else 0

    def set_capacity(self, total_capacity: int) -> GondolaCapacity:
        self.db.query(GondolaCapacity).update({"is_active": False})
        
        new_capacity = GondolaCapacity(
            total_capacity=total_capacity,
            is_active=True
        )
        self.db.add(new_capacity)
        self.db.commit()
        self.db.refresh(new_capacity)
        return new_capacity

    def record_passengers(self, gondola_id: int, adult_count: int, 
                          child_count: int) -> PassengerRecord:
        total_counted = adult_count
        
        record = PassengerRecord(
            gondola_id=gondola_id,
            adult_count=adult_count,
            child_count=child_count,
            total_counted=total_counted
        )
        self.db.add(record)
        self.db.commit()
        self.db.refresh(record)
        return record

    def update_queue(self, queue_count: int) -> tuple:
        capacity = self.get_active_capacity()
        is_limited = queue_count > capacity * QUEUE_CAPACITY_MULTIPLIER
        
        queue_data = QueueData(
            queue_count=queue_count,
            gondola_capacity=capacity,
            is_limited=is_limited
        )
        self.db.add(queue_data)
        self.db.commit()
        self.db.refresh(queue_data)
        
        return queue_data, is_limited

    def get_current_queue_status(self) -> dict:
        latest = self.db.query(QueueData).order_by(
            QueueData.timestamp.desc()
        ).first()
        
        if not latest:
            return {
                "queue_count": 0,
                "gondola_capacity": self.get_active_capacity(),
                "is_limited": False,
                "threshold": self.get_active_capacity() * QUEUE_CAPACITY_MULTIPLIER
            }
        
        return {
            "queue_count": latest.queue_count,
            "gondola_capacity": latest.gondola_capacity,
            "is_limited": latest.is_limited,
            "threshold": latest.gondola_capacity * QUEUE_CAPACITY_MULTIPLIER
        }

    def get_passenger_summary(self, start_date: datetime, 
                               end_date: datetime) -> dict:
        records = self.db.query(PassengerRecord).filter(
            PassengerRecord.timestamp >= start_date,
            PassengerRecord.timestamp <= end_date
        ).all()
        
        total_adults = sum(r.adult_count for r in records)
        total_children = sum(r.child_count for r in records)
        total_counted = sum(r.total_counted for r in records)
        
        return {
            "total_adults": total_adults,
            "total_children": total_children,
            "total_actual": total_adults + total_children,
            "total_counted": total_counted,
            "child_height_limit": CHILD_HEIGHT_LIMIT
        }
