from datetime import datetime
from sqlalchemy.orm import Session
from ..models import Farmer, Notification, RedTideEvent, Station
import json


class NotificationService:
    @staticmethod
    def create_farmer(db: Session, name: str, station_ids: list, email: str = None, phone: str = None):
        farmer = Farmer(
            name=name,
            email=email,
            phone=phone,
            station_ids=json.dumps(station_ids) if station_ids else None
        )
        db.add(farmer)
        db.commit()
        db.refresh(farmer)
        return farmer

    @staticmethod
    def get_farmers_for_station(db: Session, station_id: int):
        all_farmers = db.query(Farmer).all()
        farmers = []
        for farmer in all_farmers:
            if farmer.station_ids:
                try:
                    ids = json.loads(farmer.station_ids)
                    if str(station_id) in ids or station_id in ids:
                        farmers.append(farmer)
                except:
                    continue
        return farmers

    @staticmethod
    def create_notification(
        db: Session, 
        message: str, 
        notification_type: str,
        farmer_id: int = None,
        station_id: int = None,
        event_id: int = None
    ):
        notification = Notification(
            message=message,
            notification_type=notification_type,
            farmer_id=farmer_id,
            station_id=station_id,
            event_id=event_id
        )
        db.add(notification)
        db.commit()
        db.refresh(notification)
        return notification

    @staticmethod
    def notify_farmers(db: Session, event: RedTideEvent):
        farmers = NotificationService.get_farmers_for_station(db, event.station_id)
        
        station = db.query(Station).filter(Station.id == event.station_id).first()
        station_name = station.name if station else f"监测站 #{event.station_id}"
        
        for farmer in farmers:
            message = f"赤潮预警通知：监测站 '{station_name}' 已确认赤潮事件，请密切关注海域情况并采取相应措施。"
            NotificationService.create_notification(
                db=db,
                message=message,
                notification_type="red_tide_published",
                farmer_id=farmer.id,
                station_id=event.station_id,
                event_id=event.id
            )
        
        return len(farmers)

    @staticmethod
    def get_notifications_for_farmer(db: Session, farmer_id: int, limit: int = 50):
        return db.query(Notification).filter(
            Notification.farmer_id == farmer_id
        ).order_by(Notification.sent_at.desc()).limit(limit).all()

    @staticmethod
    def mark_notification_read(db: Session, notification_id: int):
        notification = db.query(Notification).filter(
            Notification.id == notification_id
        ).first()
        if notification:
            notification.is_read = True
            db.commit()
            db.refresh(notification)
        return notification
