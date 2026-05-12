from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from datetime import datetime, timedelta
from app.database import get_db
from app.models import TechnicalManual, BorrowRecord
from app.schemas import (
    TechnicalManualCreate, TechnicalManualResponse,
    BorrowRecordCreate, BorrowRecordResponse, MessageResponse
)

router = APIRouter(prefix="/api/manuals", tags=["技术手册管理"])


@router.post("/", response_model=TechnicalManualResponse)
def create_manual(manual: TechnicalManualCreate, db: Session = Depends(get_db)):
    existing = db.query(TechnicalManual).filter(
        TechnicalManual.manual_code == manual.manual_code
    ).first()
    if existing:
        raise HTTPException(status_code=400, detail="手册编号已存在")
    
    db_manual = TechnicalManual(
        manual_code=manual.manual_code,
        title=manual.title,
        author=manual.author,
        version=manual.version
    )
    db.add(db_manual)
    db.commit()
    db.refresh(db_manual)
    return db_manual


@router.get("/", response_model=List[TechnicalManualResponse])
def get_manuals(db: Session = Depends(get_db)):
    return db.query(TechnicalManual).all()


@router.get("/{manual_id}", response_model=TechnicalManualResponse)
def get_manual(manual_id: int, db: Session = Depends(get_db)):
    manual = db.query(TechnicalManual).filter(TechnicalManual.id == manual_id).first()
    if not manual:
        raise HTTPException(status_code=404, detail="手册不存在")
    return manual


@router.post("/{manual_id}/borrow", response_model=BorrowRecordResponse)
def borrow_manual(
    manual_id: int,
    borrow: BorrowRecordCreate,
    db: Session = Depends(get_db)
):
    manual = db.query(TechnicalManual).filter(TechnicalManual.id == manual_id).first()
    if not manual:
        raise HTTPException(status_code=404, detail="手册不存在")
    
    active_borrow = db.query(BorrowRecord).filter(
        BorrowRecord.manual_id == manual_id,
        BorrowRecord.return_date == None
    ).first()
    if active_borrow:
        raise HTTPException(status_code=400, detail="该手册已被借阅")
    
    borrow_date = datetime.now()
    due_date = borrow_date + timedelta(days=30)
    
    db_record = BorrowRecord(
        manual_id=manual_id,
        borrower=borrow.borrower,
        due_date=due_date
    )
    db.add(db_record)
    db.commit()
    db.refresh(db_record)
    
    return _build_borrow_response(db_record)


@router.get("/{manual_id}/borrow-records", response_model=List[BorrowRecordResponse])
def get_manual_borrow_records(manual_id: int, db: Session = Depends(get_db)):
    manual = db.query(TechnicalManual).filter(TechnicalManual.id == manual_id).first()
    if not manual:
        raise HTTPException(status_code=404, detail="手册不存在")
    
    records = db.query(BorrowRecord).filter(
        BorrowRecord.manual_id == manual_id
    ).order_by(BorrowRecord.created_at.desc()).all()
    
    return [_build_borrow_response(r) for r in records]


@router.get("/borrow-records/", response_model=List[BorrowRecordResponse])
def get_all_borrow_records(overdue: bool = None, db: Session = Depends(get_db)):
    query = db.query(BorrowRecord)
    
    if overdue is True:
        query = query.filter(
            BorrowRecord.return_date == None,
            BorrowRecord.due_date < datetime.now()
        )
    elif overdue is False:
        query = query.filter(
            BorrowRecord.return_date == None,
            BorrowRecord.due_date >= datetime.now()
        )
    
    records = query.order_by(BorrowRecord.created_at.desc()).all()
    
    for record in records:
        if record.return_date is None and record.due_date < datetime.now():
            record.is_overdue = True
            db.add(record)
    db.commit()
    
    return [_build_borrow_response(r) for r in records]


@router.post("/borrow-records/{record_id}/return", response_model=MessageResponse)
def return_manual(record_id: int, db: Session = Depends(get_db)):
    record = db.query(BorrowRecord).filter(BorrowRecord.id == record_id).first()
    if not record:
        raise HTTPException(status_code=404, detail="借阅记录不存在")
    
    if record.return_date is not None:
        raise HTTPException(status_code=400, detail="该手册已归还")
    
    record.return_date = datetime.now()
    db.commit()
    
    return MessageResponse(message="手册已归还")


@router.get("/borrow-records/overdue", response_model=List[BorrowRecordResponse])
def get_overdue_records(db: Session = Depends(get_db)):
    now = datetime.now()
    records = db.query(BorrowRecord).filter(
        BorrowRecord.return_date == None,
        BorrowRecord.due_date < now
    ).order_by(BorrowRecord.due_date.asc()).all()
    
    for record in records:
        record.is_overdue = True
        db.add(record)
    db.commit()
    
    return [_build_borrow_response(r) for r in records]


def _build_borrow_response(record: BorrowRecord) -> BorrowRecordResponse:
    return BorrowRecordResponse(
        id=record.id,
        manual_id=record.manual_id,
        manual_code=record.manual.manual_code,
        manual_title=record.manual.title,
        borrower=record.borrower,
        borrow_date=record.borrow_date,
        due_date=record.due_date,
        return_date=record.return_date,
        is_overdue=record.is_overdue,
        created_at=record.created_at
    )
