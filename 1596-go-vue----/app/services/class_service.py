from datetime import datetime, date, time, timedelta
from typing import List, Optional
from sqlalchemy.orm import Session
from sqlalchemy import and_

from app.models import (
    TrainingClass, ClassRegistration, Booking, User, Venue, VenueType, Notification,
    ClassStatus, BookingStatus, BookingType, PaymentStatus
)
from app.utils import generate_class_no, generate_registration_no, generate_booking_no
from app.config import settings


def create_training_class(db: Session, name: str, instructor: Optional[str],
                          venue_type_id: int, start_date: date, end_date: date,
                          class_time: time, duration_hours: float = 2.0,
                          min_students: int = 5, max_students: int = 20,
                          price_per_student: float = 100.0,
                          description: Optional[str] = None) -> TrainingClass:
    if start_date > end_date:
        raise ValueError("开始日期不能晚于结束日期")
    
    if min_students < 1:
        raise ValueError("最少人数至少为1")
    
    if max_students < min_students:
        raise ValueError("最大人数不能小于最少人数")
    
    training_class = TrainingClass(
        class_no=generate_class_no(),
        name=name,
        instructor=instructor,
        venue_type_id=venue_type_id,
        start_date=start_date,
        end_date=end_date,
        class_time=class_time,
        duration_hours=duration_hours,
        min_students=min_students,
        max_students=max_students,
        price_per_student=price_per_student,
        description=description,
        status=ClassStatus.PENDING,
    )
    db.add(training_class)
    db.commit()
    db.refresh(training_class)
    
    return training_class


def book_class_venue(db: Session, class_id: int, venue_id: int,
                     admin_user_id: int) -> List[Booking]:
    training_class = db.query(TrainingClass).filter(TrainingClass.id == class_id).first()
    if not training_class:
        raise ValueError("培训班不存在")
    
    if training_class.booking_id:
        raise ValueError("该培训班已有场地预约")
    
    from app.services.booking_service import check_venue_availability
    
    bookings = []
    current_date = training_class.start_date
    end_dt = datetime.combine(training_class.end_date, training_class.class_time) + \
             timedelta(hours=training_class.duration_hours)
    
    end_date = end_dt.date()
    
    parent_booking = None
    
    while current_date <= end_date:
        class_end_time = (datetime.combine(current_date, training_class.class_time) + 
                         timedelta(hours=training_class.duration_hours)).time()
        
        if not check_venue_availability(db, venue_id, current_date,
                                         training_class.class_time, class_end_time):
            raise ValueError(f"{current_date} 时段场地不可用")
        
        booking = Booking(
            booking_no=generate_booking_no(),
            user_id=admin_user_id,
            venue_id=venue_id,
            booking_type=BookingType.CLASS,
            booking_date=current_date,
            start_time=training_class.class_time,
            end_time=class_end_time,
            hours=training_class.duration_hours,
            original_amount=0,
            discount_amount=0,
            final_amount=0,
            status=BookingStatus.CONFIRMED,
            paid_at=datetime.now(),
            notes=f"培训班: {training_class.name}",
            parent_booking_id=parent_booking.id if parent_booking else None,
        )
        db.add(booking)
        db.flush()
        
        if not parent_booking:
            parent_booking = booking
        
        bookings.append(booking)
        
        current_date += timedelta(days=7)
    
    training_class.booking_id = parent_booking.id if parent_booking else None
    db.commit()
    
    for booking in bookings:
        db.refresh(booking)
    db.refresh(training_class)
    
    return bookings


def register_for_class(db: Session, class_id: int, user_id: int,
                       student_name: str, phone: Optional[str] = None,
                       payment_method: str = "cash") -> ClassRegistration:
    training_class = db.query(TrainingClass).filter(TrainingClass.id == class_id).first()
    if not training_class:
        raise ValueError("培训班不存在")
    
    if training_class.status in [ClassStatus.CANCELLED, ClassStatus.COMPLETED]:
        raise ValueError("培训班不接受报名")
    
    user = db.query(User).filter(User.id == user_id).first()
    if not user:
        raise ValueError("用户不存在")
    
    existing = db.query(ClassRegistration).filter(
        ClassRegistration.class_id == class_id,
        ClassRegistration.user_id == user_id,
    ).first()
    if existing:
        raise ValueError("已报名该培训班")
    
    registrations_count = db.query(ClassRegistration).filter(
        ClassRegistration.class_id == class_id,
    ).count()
    
    if registrations_count >= training_class.max_students:
        raise ValueError("培训班已满")
    
    registration = ClassRegistration(
        registration_no=generate_registration_no(),
        class_id=class_id,
        user_id=user_id,
        student_name=student_name,
        phone=phone,
        amount_paid=training_class.price_per_student,
    )
    
    db.add(registration)
    db.commit()
    db.refresh(registration)
    
    new_count = registrations_count + 1
    if new_count >= training_class.min_students and training_class.status == ClassStatus.PENDING:
        training_class.status = ClassStatus.CONFIRMED
        db.commit()
    
    return registration


def cancel_training_class(db: Session, class_id: int, reason: str = "人数不足") -> TrainingClass:
    training_class = db.query(TrainingClass).filter(TrainingClass.id == class_id).first()
    if not training_class:
        raise ValueError("培训班不存在")
    
    if training_class.status in [ClassStatus.CANCELLED, ClassStatus.COMPLETED]:
        return training_class
    
    training_class.status = ClassStatus.CANCELLED
    
    registrations = db.query(ClassRegistration).filter(
        ClassRegistration.class_id == class_id,
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
            parent_booking.cancelled_at = datetime.now()
            
            all_bookings = db.query(Booking).filter(
                or_(
                    Booking.id == training_class.booking_id,
                    Booking.parent_booking_id == training_class.booking_id,
                )
            ).all()
            
            for booking in all_bookings:
                if booking.status not in [BookingStatus.CANCELLED, BookingStatus.COMPLETED]:
                    booking.status = BookingStatus.CANCELLED
                    booking.cancelled_at = datetime.now()
    
    db.commit()
    db.refresh(training_class)
    
    return training_class
