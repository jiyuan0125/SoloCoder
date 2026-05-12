from typing import List, Optional
from datetime import date
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from app.database import get_db
from app.models import (
    Borrowing,
    BookReservation,
    Reader,
    Book,
    BookCopy,
    ReservationStatus,
)
from app.schemas import (
    BorrowingCreate,
    BorrowingReturn,
    BorrowingRenew,
    BookReservationCreate,
    Borrowing as BorrowingSchema,
    BookReservation as BookReservationSchema,
    LateFeeInfo,
)
from app.services import (
    create_borrowing,
    return_borrowing,
    renew_borrowing,
    create_book_reservation,
    calculate_late_fee,
)


router = APIRouter(prefix="/borrowings", tags=["borrowings"])


@router.post("/borrow", response_model=BorrowingSchema)
def borrow_book(data: BorrowingCreate, db: Session = Depends(get_db)):
    borrowing, msg = create_borrowing(db, data)
    if not borrowing:
        raise HTTPException(status_code=400, detail=msg)
    return borrowing


@router.post("/return", response_model=BorrowingSchema)
def return_book(data: BorrowingReturn, db: Session = Depends(get_db)):
    borrowing, msg = return_borrowing(db, data)
    if not borrowing:
        raise HTTPException(status_code=400, detail=msg)
    return borrowing


@router.post("/renew", response_model=BorrowingSchema)
def renew_book(data: BorrowingRenew, db: Session = Depends(get_db)):
    borrowing, msg = renew_borrowing(db, data)
    if not borrowing:
        raise HTTPException(status_code=400, detail=msg)
    return borrowing


@router.get("/{borrowing_id}", response_model=BorrowingSchema)
def get_borrowing(borrowing_id: int, db: Session = Depends(get_db)):
    borrowing = db.query(Borrowing).filter(Borrowing.id == borrowing_id).first()
    if not borrowing:
        raise HTTPException(status_code=404, detail="借阅记录不存在")
    return borrowing


@router.get("/late-fee/{borrowing_id}", response_model=LateFeeInfo)
def get_late_fee_info(borrowing_id: int, db: Session = Depends(get_db)):
    borrowing = db.query(Borrowing).filter(Borrowing.id == borrowing_id).first()
    if not borrowing:
        raise HTTPException(status_code=404, detail="借阅记录不存在")
    
    if borrowing.is_returned:
        return LateFeeInfo(
            overdue_days=0 if not borrowing.return_date else max(0, (borrowing.return_date - borrowing.due_date).days),
            daily_rate=borrowing.late_fee,
            max_fee=borrowing.late_fee,
            calculated_fee=borrowing.late_fee,
            final_fee=borrowing.late_fee,
        )
    
    copy = db.query(BookCopy).filter(BookCopy.id == borrowing.copy_id).first()
    book = db.query(Book).filter(Book.id == copy.book_id).first()
    return calculate_late_fee(borrowing.due_date, date.today(), book.price)


@router.post("/reserve", response_model=BookReservationSchema)
def reserve_book(data: BookReservationCreate, db: Session = Depends(get_db)):
    reservation, msg = create_book_reservation(db, data)
    if not reservation:
        raise HTTPException(status_code=400, detail=msg)
    return reservation


@router.get("/reservations/", response_model=List[BookReservationSchema])
def list_reservations(
    reader_id: Optional[int] = None,
    book_id: Optional[int] = None,
    status: Optional[ReservationStatus] = None,
    db: Session = Depends(get_db),
):
    query = db.query(BookReservation)
    if reader_id:
        query = query.filter(BookReservation.reader_id == reader_id)
    if book_id:
        query = query.filter(BookReservation.book_id == book_id)
    if status:
        query = query.filter(BookReservation.status == status.value)
    return query.order_by(BookReservation.reserved_at.desc()).all()


@router.put("/reservations/{reservation_id}/cancel")
def cancel_reservation(reservation_id: int, db: Session = Depends(get_db)):
    reservation = db.query(BookReservation).filter(BookReservation.id == reservation_id).first()
    if not reservation:
        raise HTTPException(status_code=404, detail="预约记录不存在")
    
    if reservation.status not in [ReservationStatus.PENDING.value, ReservationStatus.FULFILLED.value]:
        raise HTTPException(status_code=400, detail="该预约状态不可取消")
    
    reservation.status = ReservationStatus.CANCELLED.value
    
    pending_reservations = db.query(BookReservation).filter(
        BookReservation.book_id == reservation.book_id,
        BookReservation.status == ReservationStatus.PENDING.value,
    ).order_by(BookReservation.queue_position.asc()).all()
    
    for i, r in enumerate(pending_reservations, start=1):
        r.queue_position = i
    
    db.commit()
    db.refresh(reservation)
    
    return {"message": "取消成功", "reservation": reservation}
