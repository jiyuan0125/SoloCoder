from typing import List, Optional
from datetime import datetime
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from core import (
    get_db, Report, Company,
    ReportResponse, ReportStatus,
    calculate_monthly_summary, MonthlySummary
)

router = APIRouter()

@router.get("", response_model=List[ReportResponse])
def list_reports(
    company_id: Optional[int] = None,
    year: Optional[int] = None,
    month: Optional[int] = None,
    status: Optional[ReportStatus] = None,
    db: Session = Depends(get_db)
):
    query = db.query(Report)
    if company_id:
        query = query.filter(Report.company_id == company_id)
    if year:
        query = query.filter(Report.year == year)
    if month:
        query = query.filter(Report.month == month)
    if status:
        query = query.filter(Report.status == status)
    return query.order_by(Report.year.desc(), Report.month.desc()).all()

@router.post("/generate", response_model=ReportResponse, status_code=status.HTTP_201_CREATED)
def generate_report(company_id: int, year: int, month: int, db: Session = Depends(get_db)):
    company = db.query(Company).filter(Company.id == company_id).first()
    if not company:
        raise HTTPException(status_code=404, detail="企业不存在")
    
    existing = db.query(Report).filter(
        Report.company_id == company_id,
        Report.year == year,
        Report.month == month
    ).first()
    if existing:
        return existing
    
    summary = calculate_monthly_summary(db, company_id, year, month)
    
    report = Report(
        company_id=company_id,
        year=year,
        month=month,
        total_emission=summary["total_emission"],
        status=ReportStatus.DRAFT
    )
    
    db.add(report)
    db.commit()
    db.refresh(report)
    return report

@router.get("/{report_id}", response_model=ReportResponse)
def get_report(report_id: int, db: Session = Depends(get_db)):
    report = db.query(Report).filter(Report.id == report_id).first()
    if not report:
        raise HTTPException(status_code=404, detail="报告不存在")
    return report

@router.get("/summary/{company_id}/{year}/{month}", response_model=MonthlySummary)
def get_monthly_summary(company_id: int, year: int, month: int, db: Session = Depends(get_db)):
    company = db.query(Company).filter(Company.id == company_id).first()
    if not company:
        raise HTTPException(status_code=404, detail="企业不存在")
    
    summary = calculate_monthly_summary(db, company_id, year, month)
    if not summary:
        raise HTTPException(status_code=500, detail="生成月度汇总失败")
    return MonthlySummary(**summary)

@router.post("/{report_id}/confirm", response_model=ReportResponse)
def confirm_report(report_id: int, db: Session = Depends(get_db)):
    report = db.query(Report).filter(Report.id == report_id).first()
    if not report:
        raise HTTPException(status_code=404, detail="报告不存在")
    
    if report.status == ReportStatus.CONFIRMED:
        raise HTTPException(status_code=400, detail="报告已确认")
    
    report.status = ReportStatus.CONFIRMED
    report.confirmed_at = datetime.now()
    
    db.commit()
    db.refresh(report)
    return report

@router.delete("/{report_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_report(report_id: int, db: Session = Depends(get_db)):
    report = db.query(Report).filter(Report.id == report_id).first()
    if not report:
        raise HTTPException(status_code=404, detail="报告不存在")
    
    if report.status == ReportStatus.CONFIRMED:
        raise HTTPException(status_code=400, detail="已确认的报告无法删除")
    
    db.delete(report)
    db.commit()
