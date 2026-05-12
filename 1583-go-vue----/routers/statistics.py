from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session
from sqlalchemy import func
from datetime import datetime, date

from database import get_db
from models import InquiryRecord, LostItem, FoundItem, ItemMatch
from schemas import StatisticsResponse
from routers.complaints import get_monthly_statistics

router = APIRouter(prefix="/statistics", tags=["指标聚合"])


@router.get("/dashboard", response_model=StatisticsResponse, summary="获取dashboard聚合指标")
def get_dashboard_stats(db: Session = Depends(get_db)):
    today = date.today().strftime("%Y%m%d")
    
    today_inquiry_count = db.query(func.count(InquiryRecord.id)).filter(
        InquiryRecord.record_date == today
    ).scalar() or 0
    
    today_lost_count = db.query(func.count(LostItem.id)).filter(
        func.date(LostItem.created_at) == date.today()
    ).scalar() or 0
    
    today_found_count = db.query(func.count(FoundItem.id)).filter(
        func.date(FoundItem.created_at) == date.today()
    ).scalar() or 0
    
    today_match_count = db.query(func.count(ItemMatch.id)).filter(
        func.date(ItemMatch.created_at) == date.today()
    ).scalar() or 0
    
    monthly_stats = get_monthly_statistics(db)
    
    return StatisticsResponse(
        today_inquiry_count=today_inquiry_count,
        today_lost_items_count=today_lost_count,
        today_found_items_count=today_found_count,
        today_matches_count=today_match_count,
        monthly_complaint_count=monthly_stats["monthly_complaint_count"],
        avg_processing_hours=monthly_stats["avg_processing_hours"],
        overdue_count=monthly_stats["overdue_count"]
    )
