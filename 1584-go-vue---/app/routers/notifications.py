from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from app.database import get_db
from app.models import Notification, Staff
from app.schemas import NotificationResponse
from app.services import check_and_send_reminders, check_and_send_timeout_reminders

router = APIRouter()


@router.get("/staff/{staff_id}", response_model=List[NotificationResponse])
def get_staff_notifications(staff_id: int, unread_only: bool = True, db: Session = Depends(get_db)):
    staff = db.query(Staff).filter(Staff.id == staff_id).first()
    if not staff:
        raise HTTPException(status_code=404, detail="人员不存在")
    
    query = db.query(Notification).filter(Notification.staff_id == staff_id)
    if unread_only:
        query = query.filter(Notification.read == False)
    
    return query.order_by(Notification.sent_at.desc()).all()


@router.post("/{notification_id}/read")
def mark_notification_read(notification_id: int, db: Session = Depends(get_db)):
    notification = db.query(Notification).filter(Notification.id == notification_id).first()
    if not notification:
        raise HTTPException(status_code=404, detail="通知不存在")
    
    notification.read = True
    db.commit()
    return {"message": "通知已标记为已读"}


@router.post("/trigger-checks")
def trigger_reminder_checks(db: Session = Depends(get_db)):
    check_and_send_reminders(db)
    check_and_send_timeout_reminders(db)
    return {"message": "提醒检查已执行"}
