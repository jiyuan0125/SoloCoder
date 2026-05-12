from datetime import datetime
from sqlalchemy.orm import Session

from app.models import Alarm, AlarmStatus, AlarmLevel


ALARM_LEVEL_COLORS = {
    AlarmLevel.LOW.value: {"color": "blue", "description": "低级别告警"},
    AlarmLevel.MEDIUM.value: {"color": "yellow", "description": "中级别告警"},
    AlarmLevel.HIGH.value: {"color": "red", "description": "高级别告警"}
}


class AlarmService:
    @staticmethod
    def get_all_alarms(db: Session, status: str = None, subsystem: str = None, level: str = None):
        query = db.query(Alarm)
        if status:
            query = query.filter(Alarm.status == status)
        if subsystem:
            query = query.filter(Alarm.subsystem == subsystem)
        if level:
            query = query.filter(Alarm.level == level)
        return query.order_by(Alarm.created_at.desc()).all()

    @staticmethod
    def get_alarm(db: Session, alarm_id: int):
        return db.query(Alarm).filter(Alarm.id == alarm_id).first()

    @staticmethod
    def create_alarm(db: Session, alarm_data: dict):
        alarm = Alarm(**alarm_data)
        db.add(alarm)
        db.commit()
        db.refresh(alarm)
        return alarm

    @staticmethod
    def acknowledge_alarm(db: Session, alarm_id: int):
        alarm = AlarmService.get_alarm(db, alarm_id)
        if not alarm:
            return None
        
        alarm.status = AlarmStatus.ACKNOWLEDGED.value
        alarm.acknowledged_at = datetime.utcnow()
        db.commit()
        db.refresh(alarm)
        return alarm

    @staticmethod
    def resolve_alarm(db: Session, alarm_id: int):
        alarm = AlarmService.get_alarm(db, alarm_id)
        if not alarm:
            return None
        
        alarm.status = AlarmStatus.RESOLVED.value
        alarm.resolved_at = datetime.utcnow()
        db.commit()
        db.refresh(alarm)
        return alarm

    @staticmethod
    def get_active_alarms(db: Session):
        return db.query(Alarm).filter(
            Alarm.status.in_([AlarmStatus.ACTIVE.value, AlarmStatus.ACKNOWLEDGED.value])
        ).order_by(Alarm.created_at.desc()).all()

    @staticmethod
    def get_alarm_color(level: str):
        return ALARM_LEVEL_COLORS.get(level, {"color": "gray", "description": "未知级别"})
