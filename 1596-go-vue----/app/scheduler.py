from datetime import datetime, timedelta
from apscheduler.schedulers.background import BackgroundScheduler
from apscheduler.triggers.interval import IntervalTrigger
from sqlalchemy import and_

from app.database import SessionLocal
from app.models import (
    Booking, TrainingClass, ClassRegistration, Maintenance, Notification,
    BookingStatus, PaymentStatus, ClassStatus
)
from app.config import settings


def cancel_expired_payments():
    db = SessionLocal()
    try:
        now = datetime.now()
        timeout = timedelta(minutes=settings.PAYMENT_TIMEOUT_MINUTES)
        cutoff = now - timeout
        
        expired = db.query(Booking).filter(
            Booking.status == BookingStatus.PENDING_PAYMENT,
            Booking.created_at < cutoff,
            Booking.booking_type == "individual",
        ).all()
        
        for booking in expired:
            booking.status = BookingStatus.CANCELLED
            booking.cancelled_at = now
            
            if booking.payment:
                booking.payment.status = PaymentStatus.CANCELLED
        
        db.commit()
    finally:
        db.close()


def check_training_classes():
    db = SessionLocal()
    try:
        now = datetime.now()
        check_date = now.date() + timedelta(days=settings.CLASS_CANCEL_DAYS)
        
        pending_classes = db.query(TrainingClass).filter(
            TrainingClass.status == ClassStatus.PENDING,
            TrainingClass.start_date == check_date,
        ).all()
        
        for training_class in pending_classes:
            reg_count = db.query(ClassRegistration).filter(
                ClassRegistration.class_id == training_class.id,
            ).count()
            
            if reg_count < training_class.min_students:
                training_class.status = ClassStatus.CANCELLED
                
                registrations = db.query(ClassRegistration).filter(
                    ClassRegistration.class_id == training_class.id,
                    ClassRegistration.is_refunded == False,
                ).all()
                
                for reg in registrations:
                    reg.is_refunded = True
                    reg.refund_amount = reg.amount_paid
                
                if training_class.booking_id:
                    parent_booking = db.query(Booking).filter(
                        Booking.id == training_class.booking_id
                    ).first()
                    if parent_booking:
                        parent_booking.status = BookingStatus.CANCELLED
                        parent_booking.cancelled_at = now
                        
                        from sqlalchemy import or_
                        all_bookings = db.query(Booking).filter(
                            or_(
                                Booking.id == training_class.booking_id,
                                Booking.parent_booking_id == training_class.booking_id,
                            )
                        ).all()
                        
                        for booking in all_bookings:
                            if booking.status not in [BookingStatus.CANCELLED, BookingStatus.COMPLETED]:
                                booking.status = BookingStatus.CANCELLED
                                booking.cancelled_at = now
                
                notification = Notification(
                    message_type="class_cancellation",
                    title=f"培训班已取消：{training_class.name}",
                    content=f"因报名人数不足（{reg_count}/{training_class.min_students}），该培训班已取消，报名费将全额退还。",
                )
                db.add(notification)
        
        db.commit()
    finally:
        db.close()


def check_maintenance_notifications():
    db = SessionLocal()
    try:
        now = datetime.now()
        next_24h = now + timedelta(hours=24)
        
        maintenances = db.query(Maintenance).filter(
            Maintenance.start_time > now,
            Maintenance.start_time <= next_24h,
        ).all()
        
        for m in maintenances:
            affected = db.query(Booking).filter(
                Booking.venue_id == m.venue_id,
                Booking.booking_date == m.start_time.date(),
                Booking.status.in_([BookingStatus.CONFIRMED, BookingStatus.PENDING_PAYMENT]),
            ).all()
            
            for booking in affected:
                from datetime import time as dt_time
                booking_start = datetime.combine(booking.booking_date, booking.start_time)
                booking_end = datetime.combine(booking.booking_date, booking.end_time)
                
                if booking_start < m.end_time and booking_end > m.start_time:
                    notification = Notification(
                        user_id=booking.user_id,
                        booking_id=booking.id,
                        message_type="maintenance_notice",
                        title=f"场地[{booking.venue.name}]维修通知",
                        content=f"您预约的场地时段{booking.booking_date} {booking.start_time}-{booking.end_time} 因场地维修需要改期，请联系管理员。",
                    )
                    db.add(notification)
        
        db.commit()
    finally:
        db.close()


def complete_past_bookings():
    db = SessionLocal()
    try:
        now = datetime.now()
        today = now.date()
        
        past = db.query(Booking).filter(
            Booking.status == BookingStatus.CONFIRMED,
            Booking.booking_date <= today,
        ).all()
        
        for booking in past:
            booking_end = datetime.combine(booking.booking_date, booking.end_time)
            if booking_end < now:
                booking.status = BookingStatus.COMPLETED
        
        db.commit()
    finally:
        db.close()


def start_scheduler():
    scheduler = BackgroundScheduler()
    
    scheduler.add_job(
        cancel_expired_payments,
        IntervalTrigger(minutes=1),
        id="cancel_expired_payments",
        replace_existing=True,
    )
    
    scheduler.add_job(
        check_training_classes,
        IntervalTrigger(hours=6),
        id="check_training_classes",
        replace_existing=True,
    )
    
    scheduler.add_job(
        check_maintenance_notifications,
        IntervalTrigger(hours=12),
        id="check_maintenance_notifications",
        replace_existing=True,
    )
    
    scheduler.add_job(
        complete_past_bookings,
        IntervalTrigger(hours=1),
        id="complete_past_bookings",
        replace_existing=True,
    )
    
    scheduler.start()
    return scheduler


def run_scheduled_tasks_manual():
    print("运行定时任务...")
    cancel_expired_payments()
    print("已处理超时支付取消")
    
    check_training_classes()
    print("已检查培训班开班情况")
    
    check_maintenance_notifications()
    print("已检查维修通知")
    
    complete_past_bookings()
    print("已完成过期预约")


if __name__ == "__main__":
    run_scheduled_tasks_manual()
