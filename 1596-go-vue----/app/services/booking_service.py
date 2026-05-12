from datetime import datetime, date, time, timedelta
from typing import List, Tuple, Optional, Dict, Any
from sqlalchemy.orm import Session
from sqlalchemy import and_, or_

from app.models import (
    Booking, Venue, VenueType, User, Payment, Refund, Maintenance,
    BookingStatus, PaymentStatus, ChargeType, BookingType
)
from app.schemas import BookingCreate
from app.utils import (
    generate_booking_no, generate_payment_no, generate_refund_no,
    calculate_hours, is_time_overlap, is_continuous,
    calculate_refund_rate, generate_qr_code
)
from app.config import settings


def check_venue_availability(db: Session, venue_id: int, booking_date: date,
                              start_time: time, end_time: time,
                              exclude_booking_id: Optional[int] = None) -> bool:
    if start_time >= end_time:
        return False
    
    existing = db.query(Booking).filter(
        Booking.venue_id == venue_id,
        Booking.booking_date == booking_date,
        Booking.status.in_([BookingStatus.PENDING_PAYMENT, BookingStatus.CONFIRMED]),
    )
    if exclude_booking_id:
        existing = existing.filter(Booking.id != exclude_booking_id)
    
    for booking in existing.all():
        if is_time_overlap(start_time, end_time, booking.start_time, booking.end_time):
            return False
    
    venue = db.query(Venue).filter(Venue.id == venue_id).first()
    if not venue or not venue.is_active:
        return False
    
    start_dt = datetime.combine(booking_date, start_time)
    end_dt = datetime.combine(booking_date, end_time)
    
    maintenance = db.query(Maintenance).filter(
        Maintenance.venue_id == venue_id,
        Maintenance.start_time < end_dt,
        Maintenance.end_time > start_dt,
    ).first()
    
    if maintenance:
        return False
    
    return True


def get_available_slots(db: Session, venue_id: int, booking_date: date,
                         start_of_day: time = time(8, 0),
                         end_of_day: time = time(22, 0),
                         slot_duration_minutes: int = 60) -> List[Dict[str, Any]]:
    venue = db.query(Venue).filter(Venue.id == venue_id, Venue.is_active == True).first()
    if not venue:
        return []
    
    bookings = db.query(Booking).filter(
        Booking.venue_id == venue_id,
        Booking.booking_date == booking_date,
        Booking.status.in_([BookingStatus.PENDING_PAYMENT, BookingStatus.CONFIRMED]),
    ).all()
    
    slots = []
    current = datetime.combine(booking_date, start_of_day)
    end_dt = datetime.combine(booking_date, end_of_day)
    slot_delta = timedelta(minutes=slot_duration_minutes)
    
    while current + slot_delta <= end_dt:
        slot_start = current.time()
        slot_end = (current + slot_delta).time()
        
        is_available = True
        for booking in bookings:
            if is_time_overlap(slot_start, slot_end, booking.start_time, booking.end_time):
                is_available = False
                break
        
        if is_available:
            start_dt = datetime.combine(booking_date, slot_start)
            end_check = datetime.combine(booking_date, slot_end)
            maintenance = db.query(Maintenance).filter(
                Maintenance.venue_id == venue_id,
                Maintenance.start_time < end_check,
                Maintenance.end_time > start_dt,
            ).first()
            if maintenance:
                is_available = False
        
        slots.append({
            "start_time": slot_start,
            "end_time": slot_end,
            "available": is_available,
        })
        
        current += slot_delta
    
    return slots


def calculate_booking_price(db: Session, venue_id: int,
                             start_time: time, end_time: time,
                             is_continuous_discount: bool = False) -> Tuple[float, float, float, float]:
    venue = db.query(Venue).filter(Venue.id == venue_id).first()
    if not venue:
        raise ValueError("场地不存在")
    
    venue_type = db.query(VenueType).filter(VenueType.id == venue.venue_type_id).first()
    if not venue_type:
        raise ValueError("场地类型不存在")
    
    hours = calculate_hours(start_time, end_time)
    
    if venue_type.charge_type == ChargeType.PER_HOUR:
        original = venue_type.price * hours
    else:
        original = venue_type.price
    
    discount_rate = settings.CONTINUOUS_BOOKING_DISCOUNT if is_continuous_discount else 1.0
    discount_amount = original * (1 - discount_rate) if is_continuous_discount else 0.0
    final = original * discount_rate
    
    return hours, original, discount_amount, final


def create_booking(db: Session, data: BookingCreate, 
                   booking_type: BookingType = BookingType.INDIVIDUAL) -> Tuple[Booking, str]:
    user = db.query(User).filter(User.id == data.user_id).first()
    if not user:
        raise ValueError("用户不存在")
    
    venue = db.query(Venue).filter(Venue.id == data.venue_id).first()
    if not venue:
        raise ValueError("场地不存在")
    
    slots = [(slot.start_time, slot.end_time) for slot in data.slots]
    
    for start, end in slots:
        if not check_venue_availability(db, data.venue_id, data.booking_date, start, end):
            raise ValueError(f"时段 {start} - {end} 不可用")
    
    is_continuous_booking = is_continuous(slots) if len(slots) > 1 else False
    
    parent_booking = None
    first_booking = None
    qr_base64 = None
    
    for i, (start, end) in enumerate(slots):
        hours, original, discount, final = calculate_booking_price(
            db, data.venue_id, start, end,
            is_continuous_discount=is_continuous_booking and i > 0
        )
        
        booking_no = generate_booking_no()
        
        booking = Booking(
            booking_no=booking_no,
            user_id=data.user_id,
            venue_id=data.venue_id,
            booking_type=booking_type,
            booking_date=data.booking_date,
            start_time=start,
            end_time=end,
            hours=hours,
            original_amount=original,
            discount_amount=discount,
            final_amount=final,
            status=BookingStatus.PENDING_PAYMENT,
            is_continuous=is_continuous_booking,
            parent_booking_id=parent_booking.id if parent_booking else None,
            notes=data.notes,
        )
        
        db.add(booking)
        db.flush()
        
        if i == 0:
            first_booking = booking
            qr_token, qr_base64 = generate_qr_code(
                booking_no=booking_no,
                user_id=data.user_id,
                venue_id=data.venue_id,
                booking_date=data.booking_date,
                start_time=start,
                end_time=end,
            )
            booking.qr_token = qr_token
        else:
            booking.parent_booking_id = first_booking.id
        
        parent_booking = booking
    
    db.commit()
    db.refresh(first_booking)
    
    return first_booking, qr_base64


def process_payment(db: Session, booking_id: int, payment_method: str = "cash",
                     transaction_id: Optional[str] = None) -> Payment:
    booking = db.query(Booking).filter(Booking.id == booking_id).first()
    if not booking:
        raise ValueError("预约不存在")
    
    if booking.status != BookingStatus.PENDING_PAYMENT:
        raise ValueError("预约状态不支持支付")
    
    if booking.payment and booking.payment.status == PaymentStatus.PAID:
        raise ValueError("预约已支付")
    
    now = datetime.now()
    
    if booking.payment:
        payment = booking.payment
    else:
        payment = Payment(
            payment_no=generate_payment_no(),
            user_id=booking.user_id,
            booking_id=booking.id,
            amount=booking.final_amount,
        )
        db.add(payment)
    
    payment.status = PaymentStatus.PAID
    payment.payment_method = payment_method
    payment.transaction_id = transaction_id
    payment.paid_at = now
    
    booking.status = BookingStatus.CONFIRMED
    booking.paid_at = now
    
    db.commit()
    db.refresh(payment)
    db.refresh(booking)
    
    return payment


def cancel_booking(db: Session, booking_id: int, reason: Optional[str] = None,
                    force_refund: bool = False) -> Tuple[Booking, Optional[Refund]]:
    booking = db.query(Booking).filter(Booking.id == booking_id).first()
    if not booking:
        raise ValueError("预约不存在")
    
    if booking.status in [BookingStatus.CANCELLED, BookingStatus.COMPLETED]:
        raise ValueError("预约状态不支持取消")
    
    refund = None
    refund_amount = 0.0
    refund_rate = 0.0
    
    if booking.payment and booking.payment.status == PaymentStatus.PAID:
        if force_refund:
            refund_rate = 1.0
            reason_msg = "系统强制退款"
        else:
            refund_rate, reason_msg = calculate_refund_rate(booking.booking_date, booking.start_time)
            if reason:
                reason_msg = reason
        
        refund_amount = booking.payment.amount * refund_rate
        
        if refund_amount > 0:
            refund = Refund(
                refund_no=generate_refund_no(),
                booking_id=booking.id,
                payment_id=booking.payment.id,
                refund_amount=refund_amount,
                refund_rate=refund_rate,
                reason=reason_msg,
                status=PaymentStatus.PAID,
                processed_at=datetime.now(),
            )
            db.add(refund)
            
            booking.payment.status = PaymentStatus.REFUNDED
    
    booking.status = BookingStatus.CANCELLED
    booking.cancelled_at = datetime.now()
    
    if booking.child_bookings:
        for child in booking.child_bookings:
            if child.status not in [BookingStatus.CANCELLED, BookingStatus.COMPLETED]:
                child.status = BookingStatus.CANCELLED
                child.cancelled_at = datetime.now()
    
    db.commit()
    db.refresh(booking)
    if refund:
        db.refresh(refund)
    
    return booking, refund


def complete_booking(db: Session, booking_id: int) -> Booking:
    booking = db.query(Booking).filter(Booking.id == booking_id).first()
    if not booking:
        raise ValueError("预约不存在")
    
    if booking.status != BookingStatus.CONFIRMED:
        raise ValueError("预约状态不支持完成")
    
    booking.status = BookingStatus.COMPLETED
    db.commit()
    db.refresh(booking)
    
    return booking
