from datetime import datetime, date
from sqlalchemy.orm import Session
from typing import List, Tuple
from app.models import MaintenanceTask, MaintenanceStatus, MaintenanceType


class MaintenanceChecker:
    def __init__(self, db: Session):
        self.db = db

    def check_overdue_tasks(self) -> List[MaintenanceTask]:
        today = date.today()
        overdue_tasks = self.db.query(MaintenanceTask).filter(
            MaintenanceTask.scheduled_date < today,
            MaintenanceTask.status.in_([MaintenanceStatus.PENDING, MaintenanceStatus.IN_PROGRESS])
        ).all()
        
        for task in overdue_tasks:
            if task.status == MaintenanceStatus.PENDING:
                task.status = MaintenanceStatus.OVERDUE
        self.db.commit()
        
        return overdue_tasks

    def has_overdue_maintenance(self, cabin_id: int = None) -> bool:
        self.check_overdue_tasks()
        
        query = self.db.query(MaintenanceTask).filter(
            MaintenanceTask.status == MaintenanceStatus.OVERDUE
        )
        
        if cabin_id:
            query = query.filter(MaintenanceTask.cabin_id == cabin_id)
        
        return query.count() > 0

    def can_operate(self, cabin_id: int = None) -> Tuple[bool, List[str]]:
        overdue_tasks = self.check_overdue_tasks()
        reasons = []
        
        query = self.db.query(MaintenanceTask).filter(
            MaintenanceTask.status == MaintenanceStatus.OVERDUE
        )
        
        if cabin_id:
            query = query.filter(MaintenanceTask.cabin_id == cabin_id)
        
        cabin_overdue = query.all()
        
        for task in cabin_overdue:
            cabin_info = f"轿厢 {task.cabin_id}" if task.cabin_id else "系统"
            reasons.append(f"{cabin_info}: {task.task_type.value} 维护超期（原定于 {task.scheduled_date}）")
        
        return len(reasons) == 0, reasons
