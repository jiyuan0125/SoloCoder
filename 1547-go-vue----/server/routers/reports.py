from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from datetime import date
from server.database import get_db
from server.models import MonthlyReport, River
from server.schemas import (
    MonthlyReportCreate, MonthlyReportUpdate,
    MonthlyReport as MonthlyReportSchema
)
from server.services import (
    get_monthly_report_deadline, calculate_report_stats
)

router = APIRouter(prefix="/reports", tags=["reports"])

@router.get("/", response_model=List[MonthlyReportSchema])
def list_reports(db: Session = Depends(get_db)):
    return db.query(MonthlyReport).order_by(MonthlyReport.created_at.desc()).all()

@router.post("/", response_model=MonthlyReportSchema)
def create_report(report: MonthlyReportCreate, db: Session = Depends(get_db)):
    river = db.query(River).filter(River.id == report.river_id).first()
    if not river:
        raise HTTPException(status_code=400, detail="河流不存在")
    
    existing = db.query(MonthlyReport).filter(
        MonthlyReport.river_id == report.river_id,
        MonthlyReport.report_month == report.report_month
    ).first()
    if existing:
        raise HTTPException(status_code=400, detail="该月份的报告已存在")
    
    deadline = get_monthly_report_deadline(report.report_month)
    stats = calculate_report_stats(db, report.river_id, report.report_month)
    
    db_report = MonthlyReport(
        **report.model_dump(),
        deadline=deadline,
        **stats
    )
    db.add(db_report)
    db.commit()
    db.refresh(db_report)
    return db_report

@router.get("/{report_id}", response_model=MonthlyReportSchema)
def get_report(report_id: int, db: Session = Depends(get_db)):
    report = db.query(MonthlyReport).filter(MonthlyReport.id == report_id).first()
    if not report:
        raise HTTPException(status_code=404, detail="月报不存在")
    return report

@router.put("/{report_id}", response_model=MonthlyReportSchema)
def update_report(report_id: int, report: MonthlyReportUpdate, db: Session = Depends(get_db)):
    db_report = db.query(MonthlyReport).filter(MonthlyReport.id == report_id).first()
    if not db_report:
        raise HTTPException(status_code=404, detail="月报不存在")
    update_data = report.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(db_report, key, value)
    db.commit()
    db.refresh(db_report)
    return db_report

@router.post("/{report_id}/submit", response_model=MonthlyReportSchema)
def submit_report(report_id: int, db: Session = Depends(get_db)):
    db_report = db.query(MonthlyReport).filter(MonthlyReport.id == report_id).first()
    if not db_report:
        raise HTTPException(status_code=404, detail="月报不存在")
    if db_report.is_submitted:
        raise HTTPException(status_code=400, detail="报告已提交")
    db_report.is_submitted = True
    db_report.submitted_date = date.today()
    db.commit()
    db.refresh(db_report)
    return db_report

@router.delete("/{report_id}")
def delete_report(report_id: int, db: Session = Depends(get_db)):
    db_report = db.query(MonthlyReport).filter(MonthlyReport.id == report_id).first()
    if not db_report:
        raise HTTPException(status_code=404, detail="月报不存在")
    db.delete(db_report)
    db.commit()
    return {"message": "删除成功"}
