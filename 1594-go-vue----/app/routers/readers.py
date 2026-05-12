from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from sqlalchemy import or_

from app.database import get_db
from app.models import Reader, Borrowing, SeatReservation
from app.schemas import (
    ReaderCreate,
    ReaderUpdate,
    Reader as ReaderSchema,
    Borrowing as BorrowingSchema,
    SeatReservation as SeatReservationSchema,
)


router = APIRouter(prefix="/readers", tags=["readers"])


@router.post("", response_model=ReaderSchema)
def create_reader(reader_data: ReaderCreate, db: Session = Depends(get_db)):
    existing = db.query(Reader).filter(Reader.card_number == reader_data.card_number).first()
    if existing:
        raise HTTPException(status_code=400, detail="读者证号已存在")
    
    reader = Reader(**reader_data.model_dump())
    db.add(reader)
    db.commit()
    db.refresh(reader)
    return reader


@router.get("", response_model=List[ReaderSchema])
def list_readers(
    skip: int = 0,
    limit: int = 100,
    search: Optional[str] = None,
    db: Session = Depends(get_db),
):
    query = db.query(Reader)
    if search:
        query = query.filter(
            or_(
                Reader.name.contains(search),
                Reader.card_number.contains(search),
                Reader.phone.contains(search),
                Reader.email.contains(search),
            )
        )
    return query.offset(skip).limit(limit).all()


@router.get("/{reader_id}", response_model=ReaderSchema)
def get_reader(reader_id: int, db: Session = Depends(get_db)):
    reader = db.query(Reader).filter(Reader.id == reader_id).first()
    if not reader:
        raise HTTPException(status_code=404, detail="读者不存在")
    return reader


@router.put("/{reader_id}", response_model=ReaderSchema)
def update_reader(reader_id: int, update_data: ReaderUpdate, db: Session = Depends(get_db)):
    reader = db.query(Reader).filter(Reader.id == reader_id).first()
    if not reader:
        raise HTTPException(status_code=404, detail="读者不存在")
    
    for key, value in update_data.model_dump(exclude_unset=True).items():
        if key == "reader_type":
            value = value.value
        setattr(reader, key, value)
    
    db.commit()
    db.refresh(reader)
    return reader


@router.delete("/{reader_id}")
def delete_reader(reader_id: int, db: Session = Depends(get_db)):
    reader = db.query(Reader).filter(Reader.id == reader_id).first()
    if not reader:
        raise HTTPException(status_code=404, detail="读者不存在")
    
    active_borrowings = db.query(Borrowing).filter(
        Borrowing.reader_id == reader_id,
        Borrowing.is_returned == False,
    ).first()
    if active_borrowings:
        raise HTTPException(status_code=400, detail="读者有未归还的图书")
    
    db.delete(reader)
    db.commit()
    return {"message": "删除成功"}


@router.get("/{reader_id}/borrowings", response_model=List[BorrowingSchema])
def get_reader_borrowings(reader_id: int, db: Session = Depends(get_db)):
    reader = db.query(Reader).filter(Reader.id == reader_id).first()
    if not reader:
        raise HTTPException(status_code=404, detail="读者不存在")
    
    return db.query(Borrowing).filter(Borrowing.reader_id == reader_id).order_by(Borrowing.borrow_date.desc()).all()


@router.get("/{reader_id}/active-borrowings", response_model=List[BorrowingSchema])
def get_reader_active_borrowings(reader_id: int, db: Session = Depends(get_db)):
    reader = db.query(Reader).filter(Reader.id == reader_id).first()
    if not reader:
        raise HTTPException(status_code=404, detail="读者不存在")
    
    return db.query(Borrowing).filter(
        Borrowing.reader_id == reader_id,
        Borrowing.is_returned == False,
    ).order_by(Borrowing.due_date.asc()).all()


@router.get("/{reader_id}/seat-reservations", response_model=List[SeatReservationSchema])
def get_reader_seat_reservations(reader_id: int, db: Session = Depends(get_db)):
    reader = db.query(Reader).filter(Reader.id == reader_id).first()
    if not reader:
        raise HTTPException(status_code=404, detail="读者不存在")
    
    return db.query(SeatReservation).filter(
        SeatReservation.reader_id == reader_id,
    ).order_by(SeatReservation.reservation_date.desc(), SeatReservation.start_time.desc()).all()
