from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from sqlalchemy import func
from typing import List, Optional
from datetime import date

from database import get_db
from models import InquiryRecord
from schemas import (
    InquiryRecordCreate,
    InquiryRecordUpdate,
    InquiryRecordResponse
)
from config import settings

router = APIRouter(prefix="/inquiries", tags=["问询记录"])


def generate_record_number(db: Session, record_date: str) -> str:
    max_seq = db.query(func.max(InquiryRecord.sequence_number)).filter(
        InquiryRecord.record_date == record_date
    ).scalar()
    
    next_seq = (max_seq or 0) + 1
    seq_str = f"{next_seq:03d}"
    
    return f"{settings.WX_RECORD_PREFIX}{record_date}{seq_str}", next_seq


@router.post("/", response_model=InquiryRecordResponse, summary="创建问询记录")
def create_inquiry(
    inquiry: InquiryRecordCreate,
    db: Session = Depends(get_db)
):
    today = date.today().strftime("%Y%m%d")
    record_number, seq = generate_record_number(db, today)
    
    db_inquiry = InquiryRecord(
        record_number=record_number,
        record_date=today,
        sequence_number=seq,
        inquirer_name=inquiry.inquirer_name,
        inquirer_phone=inquiry.inquirer_phone,
        content=inquiry.content,
        response=inquiry.response
    )
    
    db.add(db_inquiry)
    db.commit()
    db.refresh(db_inquiry)
    
    return db_inquiry


@router.get("/", response_model=List[InquiryRecordResponse], summary="查询问询记录列表")
def list_inquiries(
    date_filter: Optional[str] = Query(None, description="按日期筛选，格式 YYYYMMDD"),
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db)
):
    query = db.query(InquiryRecord)
    if date_filter:
        query = query.filter(InquiryRecord.record_date == date_filter)
    
    return query.order_by(InquiryRecord.created_at.desc()).offset(skip).limit(limit).all()


@router.get("/{record_id}", response_model=InquiryRecordResponse, summary="查询单条问询记录")
def get_inquiry(record_id: int, db: Session = Depends(get_db)):
    inquiry = db.query(InquiryRecord).filter(InquiryRecord.id == record_id).first()
    if not inquiry:
        raise HTTPException(status_code=404, detail="问询记录不存在")
    return inquiry


@router.get("/by-number/{record_number}", response_model=InquiryRecordResponse, summary="通过编号查询问询记录")
def get_inquiry_by_number(record_number: str, db: Session = Depends(get_db)):
    inquiry = db.query(InquiryRecord).filter(InquiryRecord.record_number == record_number).first()
    if not inquiry:
        raise HTTPException(status_code=404, detail="问询记录不存在")
    return inquiry


@router.put("/{record_id}", response_model=InquiryRecordResponse, summary="更新问询记录")
def update_inquiry(
    record_id: int,
    inquiry: InquiryRecordUpdate,
    db: Session = Depends(get_db)
):
    db_inquiry = db.query(InquiryRecord).filter(InquiryRecord.id == record_id).first()
    if not db_inquiry:
        raise HTTPException(status_code=404, detail="问询记录不存在")
    
    update_data = inquiry.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(db_inquiry, key, value)
    
    db.commit()
    db.refresh(db_inquiry)
    return db_inquiry


@router.delete("/{record_id}", summary="删除问询记录")
def delete_inquiry(record_id: int, db: Session = Depends(get_db)):
    db_inquiry = db.query(InquiryRecord).filter(InquiryRecord.id == record_id).first()
    if not db_inquiry:
        raise HTTPException(status_code=404, detail="问询记录不存在")
    
    db.delete(db_inquiry)
    db.commit()
    return {"message": "删除成功"}
