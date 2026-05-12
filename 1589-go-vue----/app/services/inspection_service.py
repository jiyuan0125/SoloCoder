from datetime import datetime, date
from sqlalchemy.orm import Session
from app.models import Inspection, InspectionType, InspectionStatus


class InspectionService:
    def __init__(self, db: Session):
        self.db = db

    def create_inspection(self, inspection_type: InspectionType, 
                          scheduled_date: datetime, notes: str = None) -> Inspection:
        inspection = Inspection(
            inspection_type=inspection_type,
            scheduled_date=scheduled_date,
            notes=notes,
            status=InspectionStatus.PENDING,
            resolved=True
        )
        self.db.add(inspection)
        self.db.commit()
        self.db.refresh(inspection)
        return inspection

    def update_inspection(self, inspection_id: int, **kwargs) -> Inspection:
        inspection = self.db.query(Inspection).filter(
            Inspection.id == inspection_id
        ).first()
        
        if not inspection:
            return None
        
        for key, value in kwargs.items():
            if value is not None:
                setattr(inspection, key, value)
        
        if 'issues_found' in kwargs and kwargs['issues_found']:
            inspection.status = InspectionStatus.URGENT
            inspection.is_urgent = True
            inspection.resolved = False
        
        if 'completed_date' in kwargs and kwargs['completed_date']:
            if inspection.resolved:
                inspection.status = InspectionStatus.COMPLETED
        
        self.db.commit()
        self.db.refresh(inspection)
        return inspection

    def can_operate_tomorrow(self) -> bool:
        today = date.today()
        
        urgent_inspections = self.db.query(Inspection).filter(
            Inspection.is_urgent == True,
            Inspection.resolved == False
        ).all()
        
        for inspection in urgent_inspections:
            if inspection.scheduled_date.date() <= today:
                return False
        
        return True

    def get_pending_inspections(self) -> list:
        return self.db.query(Inspection).filter(
            Inspection.status != InspectionStatus.COMPLETED
        ).order_by(Inspection.scheduled_date).all()

    def get_inspections_by_type(self, inspection_type: InspectionType) -> list:
        return self.db.query(Inspection).filter(
            Inspection.inspection_type == inspection_type
        ).order_by(Inspection.scheduled_date.desc()).all()

    def get_urgent_inspections(self) -> list:
        return self.db.query(Inspection).filter(
            Inspection.is_urgent == True,
            Inspection.resolved == False
        ).all()
