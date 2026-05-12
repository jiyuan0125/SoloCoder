from typing import List, Optional
from datetime import datetime, date, time, timedelta
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from sqlalchemy import func
from app.database import get_db
from app.models import VisitReservation, TimeSlotConfig
from app.schemas import (
    VisitReservationCreate, VisitReservationResponse,
    TimeSlotConfigCreate, TimeSlotConfigResponse
)

router = APIRouter()


@router.get("/time-slots", response_model=List[TimeSlotConfigResponse])
def list_time_slots(db: Session = Depends(get_db)):
    slots = db.query(TimeSlotConfig).filter(TimeSlotConfig.is_active == True).all()
    if not slots:
        default_slots = [
            {"time_slot": "09:00-10:00", "max_visitors": 50},
            {"time_slot": "10:00-11:00", "max_visitors": 50},
            {"time_slot": "11:00-12:00", "max_visitors": 50},
            {"time_slot": "13:00-14:00", "max_visitors": 50},
            {"time_slot": "14:00-15:00", "max_visitors": 50},
            {"time_slot": "15:00-16:00", "max_visitors": 50},
            {"time_slot": "16:00-17:00", "max_visitors": 50},
        ]
        for slot in default_slots:
            db.add(TimeSlotConfig(**slot))
        db.commit()
        slots = db.query(TimeSlotConfig).filter(TimeSlotConfig.is_active == True).all()
    return slots


@router.post("/time-slots", response_model=TimeSlotConfigResponse, status_code=status.HTTP_201_CREATED)
def create_time_slot(slot: TimeSlotConfigCreate, db: Session = Depends(get_db)):
    existing = db.query(TimeSlotConfig).filter(TimeSlotConfig.time_slot == slot.time_slot).first()
    if existing:
        existing.max_visitors = slot.max_visitors
        existing.is_active = True
        db.commit()
        db.refresh(existing)
        return existing
    db_slot = TimeSlotConfig(**slot.dict())
    db.add(db_slot)
    db.commit()
    db.refresh(db_slot)
    return db_slot


@router.get("/", response_model=List[VisitReservationResponse])
def list_reservations(
    visit_date: Optional[date] = None,
    id_number: Optional[str] = None,
    status: Optional[str] = None,
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db)
):
    query = db.query(VisitReservation)
    if visit_date:
        query = query.filter(VisitReservation.visit_date == visit_date)
    if id_number:
        query = query.filter(VisitReservation.id_number == id_number)
    if status:
        query = query.filter(VisitReservation.status == status)
    return query.order_by(VisitReservation.created_at.desc()).offset(skip).limit(limit).all()


@router.post("/", response_model=VisitReservationResponse, status_code=status.HTTP_201_CREATED)
def create_reservation(reservation: VisitReservationCreate, db: Session = Depends(get_db)):
    today = date.today()
    if reservation.visit_date < today:
        raise HTTPException(status_code=400, detail="不能预约过去的日期")

    time_slot = db.query(TimeSlotConfig).filter(
        TimeSlotConfig.time_slot == reservation.time_slot,
        TimeSlotConfig.is_active == True
    ).first()
    if not time_slot:
        raise HTTPException(status_code=400, detail=f"无效的时间段: {reservation.time_slot}")

    existing = db.query(VisitReservation).filter(
        VisitReservation.id_number == reservation.id_number,
        VisitReservation.visit_date == reservation.visit_date,
        VisitReservation.status == "已预约"
    ).first()
    if existing:
        raise HTTPException(status_code=400, detail="该身份证号当天已预约，一天只能预约一次")

    total_visitors = db.query(func.sum(VisitReservation.visitor_count)).filter(
        VisitReservation.visit_date == reservation.visit_date,
        VisitReservation.time_slot == reservation.time_slot,
        VisitReservation.status == "已预约"
    ).scalar() or 0

    if total_visitors + reservation.visitor_count > time_slot.max_visitors:
        raise HTTPException(
            status_code=400,
            detail=f"该时间段已预约 {total_visitors} 人，最多 {time_slot.max_visitors} 人，剩余名额不足"
        )

    db_reservation = VisitReservation(**reservation.dict())
    db.add(db_reservation)
    db.commit()
    db.refresh(db_reservation)
    return db_reservation


@router.get("/{reservation_id}", response_model=VisitReservationResponse)
def get_reservation(reservation_id: int, db: Session = Depends(get_db)):
    reservation = db.query(VisitReservation).filter(VisitReservation.id == reservation_id).first()
    if not reservation:
        raise HTTPException(status_code=404, detail="预约不存在")
    return reservation


@router.post("/{reservation_id}/cancel")
def cancel_reservation(reservation_id: int, id_number: str, db: Session = Depends(get_db)):
    reservation = db.query(VisitReservation).filter(VisitReservation.id == reservation_id).first()
    if not reservation:
        raise HTTPException(status_code=404, detail="预约不存在")

    if reservation.id_number != id_number:
        raise HTTPException(status_code=403, detail="身份证号不匹配")

    if reservation.status != "已预约":
        raise HTTPException(status_code=400, detail=f"该预约状态为 '{reservation.status}'，无法取消")

    time_slot_parts = reservation.time_slot.split("-")
    slot_start_time = time.fromisoformat(time_slot_parts[0]) if len(time_slot_parts) > 0 else time(9, 0)
    visit_datetime = datetime.combine(reservation.visit_date, slot_start_time)

    now = datetime.now()
    time_diff = visit_datetime - now

    if time_diff < timedelta(hours=2):
        raise HTTPException(status_code=400, detail="预约开始前2小时内无法取消预约")

    reservation.status = "已取消"
    reservation.cancelled_at = now
    db.commit()
    db.refresh(reservation)
    return {"message": "预约已取消", "reservation_id": reservation_id}


@router.get("/availability/{visit_date}")
def check_availability(visit_date: date, db: Session = Depends(get_db)):
    slots = db.query(TimeSlotConfig).filter(TimeSlotConfig.is_active == True).all()
    availability = []

    for slot in slots:
        total_visitors = db.query(func.sum(VisitReservation.visitor_count)).filter(
            VisitReservation.visit_date == visit_date,
            VisitReservation.time_slot == slot.time_slot,
            VisitReservation.status == "已预约"
        ).scalar() or 0
        available = slot.max_visitors - total_visitors
        availability.append({
            "time_slot": slot.time_slot,
            "max_visitors": slot.max_visitors,
            "booked_visitors": total_visitors,
            "available": available,
            "is_full": available <= 0
        })

    return {"visit_date": visit_date, "slots": availability}
