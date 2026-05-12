from typing import List, Optional
from datetime import date, time
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from app.database import get_db
from app.models import (
    Seat,
    SeatReservation,
    SeatReservationStatus,
    SEAT_TIME_SLOTS,
)
from app.schemas import (
    SeatCreate,
    SeatUpdate,
    SeatReservationCreate,
    SeatCheckIn,
    Seat as SeatSchema,
    SeatReservation as SeatReservationSchema,
)
from app.services import (
    create_seat_reservation,
    check_in_seat,
    cancel_seat_reservation,
    check_expired_seat_reservations,
    is_seat_available,
)


router = APIRouter(prefix="/seats", tags=["seats"])


@router.post("", response_model=SeatSchema)
def create_seat(seat_data: SeatCreate, db: Session = Depends(get_db)):
    existing = db.query(Seat).filter(Seat.seat_number == seat_data.seat_number).first()
    if existing:
        raise HTTPException(status_code=400, detail="座位编号已存在")
    
    seat = Seat(**seat_data.model_dump())
    db.add(seat)
    db.commit()
    db.refresh(seat)
    return seat


@router.get("", response_model=List[SeatSchema])
def list_seats(
    floor: Optional[str] = None,
    area: Optional[str] = None,
    is_active: Optional[bool] = None,
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db),
):
    query = db.query(Seat)
    if floor:
        query = query.filter(Seat.floor == floor)
    if area:
        query = query.filter(Seat.area == area)
    if is_active is not None:
        query = query.filter(Seat.is_active == is_active)
    return query.offset(skip).limit(limit).all()


@router.get("/{seat_id}", response_model=SeatSchema)
def get_seat(seat_id: int, db: Session = Depends(get_db)):
    seat = db.query(Seat).filter(Seat.id == seat_id).first()
    if not seat:
        raise HTTPException(status_code=404, detail="座位不存在")
    return seat


@router.put("/{seat_id}", response_model=SeatSchema)
def update_seat(seat_id: int, update_data: SeatUpdate, db: Session = Depends(get_db)):
    seat = db.query(Seat).filter(Seat.id == seat_id).first()
    if not seat:
        raise HTTPException(status_code=404, detail="座位不存在")
    
    for key, value in update_data.model_dump(exclude_unset=True).items():
        setattr(seat, key, value)
    
    db.commit()
    db.refresh(seat)
    return seat


@router.delete("/{seat_id}")
def delete_seat(seat_id: int, db: Session = Depends(get_db)):
    seat = db.query(Seat).filter(Seat.id == seat_id).first()
    if not seat:
        raise HTTPException(status_code=404, detail="座位不存在")
    
    active_reservations = db.query(SeatReservation).filter(
        SeatReservation.seat_id == seat_id,
        SeatReservation.status.in_([
            SeatReservationStatus.RESERVED.value,
            SeatReservationStatus.CHECKED_IN.value,
        ]),
    ).first()
    if active_reservations:
        raise HTTPException(status_code=400, detail="存在活跃的座位预约")
    
    db.delete(seat)
    db.commit()
    return {"message": "删除成功"}


@router.get("/time-slots")
def get_time_slots():
    return {
        "slots": [
            {"start": start.strftime("%H:%M"), "end": end.strftime("%H:%M")}
            for start, end in SEAT_TIME_SLOTS
        ],
        "note": "每档2小时，8:00-21:00",
    }


@router.post("/reserve", response_model=SeatReservationSchema)
def reserve_seat(data: SeatReservationCreate, db: Session = Depends(get_db)):
    check_expired_seat_reservations(db)
    reservation, msg = create_seat_reservation(db, data)
    if not reservation:
        raise HTTPException(status_code=400, detail=msg)
    return reservation


@router.post("/check-in", response_model=SeatReservationSchema)
def check_in(data: SeatCheckIn, db: Session = Depends(get_db)):
    check_expired_seat_reservations(db)
    reservation, msg = check_in_seat(db, data.reservation_id)
    if not reservation:
        raise HTTPException(status_code=400, detail=msg)
    return reservation


@router.put("/reservations/{reservation_id}/cancel")
def cancel_reservation(reservation_id: int, db: Session = Depends(get_db)):
    reservation, msg = cancel_seat_reservation(db, reservation_id)
    if not reservation:
        raise HTTPException(status_code=400, detail=msg)
    return {"message": "取消成功", "reservation": reservation}


@router.get("/reservations/", response_model=List[SeatReservationSchema])
def list_reservations(
    seat_id: Optional[int] = None,
    reader_id: Optional[int] = None,
    reservation_date: Optional[date] = None,
    status: Optional[SeatReservationStatus] = None,
    db: Session = Depends(get_db),
):
    query = db.query(SeatReservation)
    if seat_id:
        query = query.filter(SeatReservation.seat_id == seat_id)
    if reader_id:
        query = query.filter(SeatReservation.reader_id == reader_id)
    if reservation_date:
        query = query.filter(SeatReservation.reservation_date == reservation_date)
    if status:
        query = query.filter(SeatReservation.status == status.value)
    return query.order_by(SeatReservation.reservation_date.desc(), SeatReservation.start_time.asc()).all()


@router.get("/{seat_id}/availability")
def check_seat_availability(
    seat_id: int,
    check_date: date,
    db: Session = Depends(get_db),
):
    seat = db.query(Seat).filter(Seat.id == seat_id).first()
    if not seat:
        raise HTTPException(status_code=404, detail="座位不存在")
    
    check_expired_seat_reservations(db)
    
    availability = []
    for start, end in SEAT_TIME_SLOTS:
        available = is_seat_available(db, seat_id, check_date, start, end)
        availability.append({
            "start": start.strftime("%H:%M"),
            "end": end.strftime("%H:%M"),
            "available": available,
        })
    
    return {
        "seat_id": seat_id,
        "date": check_date.isoformat(),
        "availability": availability,
    }


@router.post("/check-expired")
def trigger_check_expired(db: Session = Depends(get_db)):
    check_expired_seat_reservations(db)
    return {"message": "过期预约检查完成"}
