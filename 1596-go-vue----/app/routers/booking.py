from datetime import date, datetime
from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from app.database import get_db
from app.models import Booking, Payment, Venue
from app.schemas import (
    BookingCreate, BookingResponse, PaymentCreate, PaymentResponse,
    RefundResponse, QRVerifyResponse
)
from app.services.booking_service import (
    create_booking, process_payment, cancel_booking
)
from app.utils import verify_qr_token

router = APIRouter(prefix="/api/bookings", tags=["预约管理"])


@router.post("/")
def create_new_booking(data: BookingCreate, db: Session = Depends(get_db)):
    try:
        booking, qr_base64 = create_booking(db, data)
        return {
            "booking": BookingResponse.model_validate(booking),
            "qr_code_base64": qr_base64,
        }
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/", response_model=List[BookingResponse])
def list_bookings(
    user_id: Optional[int] = None,
    venue_id: Optional[int] = None,
    booking_date: Optional[date] = None,
    status: Optional[str] = None,
    db: Session = Depends(get_db),
):
    query = db.query(Booking)
    if user_id:
        query = query.filter(Booking.user_id == user_id)
    if venue_id:
        query = query.filter(Booking.venue_id == venue_id)
    if booking_date:
        query = query.filter(Booking.booking_date == booking_date)
    if status:
        query = query.filter(Booking.status == status)
    
    return query.order_by(Booking.created_at.desc()).all()


@router.get("/{booking_id}", response_model=BookingResponse)
def get_booking(booking_id: int, db: Session = Depends(get_db)):
    booking = db.query(Booking).filter(Booking.id == booking_id).first()
    if not booking:
        raise HTTPException(status_code=404, detail="预约不存在")
    return booking


@router.post("/pay", response_model=PaymentResponse)
def pay_booking(data: PaymentCreate, db: Session = Depends(get_db)):
    try:
        return process_payment(
            db, data.booking_id, data.payment_method, data.transaction_id
        )
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.post("/{booking_id}/cancel")
def cancel_booking_endpoint(booking_id: int, reason: Optional[str] = None, 
                             db: Session = Depends(get_db)):
    try:
        booking, refund = cancel_booking(db, booking_id, reason)
        return {
            "booking": BookingResponse.model_validate(booking),
            "refund": RefundResponse.model_validate(refund) if refund else None,
        }
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.post("/qr/verify", response_model=QRVerifyResponse)
def verify_qr_endpoint(token: str, db: Session = Depends(get_db)):
    verified = verify_qr_token(token)
    if not verified:
        return QRVerifyResponse(
            valid=False,
            message="二维码无效或已过期"
        )
    
    booking = db.query(Booking).filter(
        Booking.booking_no == verified["booking_no"],
        Booking.qr_token == token,
    ).first()
    
    if not booking:
        return QRVerifyResponse(
            valid=False,
            message="预约记录不存在"
        )
    
    venue = db.query(Venue).filter(Venue.id == booking.venue_id).first()
    
    return QRVerifyResponse(
        valid=True,
        booking_no=booking.booking_no,
        venue_name=venue.name if venue else None,
        booking_date=booking.booking_date,
        start_time=booking.start_time,
        end_time=booking.end_time,
        status=booking.status.value,
        message="验证通过"
    )


@router.get("/{booking_id}/qr")
def get_booking_qr(booking_id: int, db: Session = Depends(get_db)):
    booking = db.query(Booking).filter(Booking.id == booking_id).first()
    if not booking:
        raise HTTPException(status_code=404, detail="预约不存在")
    
    if not booking.qr_token:
        raise HTTPException(status_code=400, detail="该预约没有二维码")
    
    return {
        "booking_no": booking.booking_no,
        "qr_token": booking.qr_token,
    }
