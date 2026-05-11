from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List, Optional
from datetime import date as date_type
from src.core.database import get_db
from src.core.services import ReportService
from src.server.schemas import DailyReportResponse


router = APIRouter(prefix="/reports", tags=["reports"])


@router.post("/generate")
def generate_daily_report(
    report_date: Optional[date_type] = None,
    db: Session = Depends(get_db)
):
    service = ReportService(db)
    report = service.generate_daily_report(report_date)
    return {"message": "Report generated", "date": str(report.report_date)}


@router.get("/", response_model=List[DailyReportResponse])
def list_reports(db: Session = Depends(get_db)):
    service = ReportService(db)
    return service.list_reports()


@router.get("/{report_date}", response_model=DailyReportResponse)
def get_report(report_date: date_type, db: Session = Depends(get_db)):
    service = ReportService(db)
    report = service.get_report(report_date)
    if not report:
        raise HTTPException(status_code=404, detail="Report not found")
    return report
