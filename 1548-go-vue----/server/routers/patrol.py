from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from typing import List, Optional
from datetime import date, datetime

from ..database import get_db
from ..models import RiverSection, PatrolTask, MonthlyReport
from ..schemas import (
    RiverSectionCreate, RiverSectionResponse,
    PatrolTaskCreate, PatrolTaskUpdate, PatrolTaskResponse,
    MonthlyReportResponse
)
from ..services import (
    create_river_section, generate_patrol_tasks,
    create_monthly_report, execute_monthly_report_if_needed
)

router = APIRouter(tags=["巡查与月报"])


@router.post("/river-sections", response_model=RiverSectionResponse)
def create_section(
    section_data: RiverSectionCreate,
    db: Session = Depends(get_db)
):
    existing = db.query(RiverSection).filter(RiverSection.name == section_data.name).first()
    if existing:
        raise HTTPException(status_code=400, detail="河段名称已存在")
    return create_river_section(db, section_data)


@router.get("/river-sections", response_model=List[RiverSectionResponse])
def list_river_sections(
    skip: int = Query(0, ge=0),
    limit: int = Query(20, ge=1, le=100),
    db: Session = Depends(get_db)
):
    return db.query(RiverSection).offset(skip).limit(limit).all()


@router.get("/river-sections/{section_id}", response_model=RiverSectionResponse)
def get_river_section(section_id: int, db: Session = Depends(get_db)):
    section = db.query(RiverSection).filter(RiverSection.id == section_id).first()
    if not section:
        raise HTTPException(status_code=404, detail="河段不存在")
    return section


@router.post("/patrol-tasks/generate")
def generate_tasks(db: Session = Depends(get_db)):
    tasks = generate_patrol_tasks(db)
    return {"generated_count": len(tasks), "tasks": tasks}


@router.post("/patrol-tasks", response_model=PatrolTaskResponse)
def create_patrol_task(
    task_data: PatrolTaskCreate,
    db: Session = Depends(get_db)
):
    section = db.query(RiverSection).filter(RiverSection.id == task_data.river_section_id).first()
    if not section:
        raise HTTPException(status_code=404, detail="河段不存在")
    
    db_task = PatrolTask(**task_data.model_dump())
    db.add(db_task)
    db.commit()
    db.refresh(db_task)
    return db_task


@router.get("/patrol-tasks", response_model=List[PatrolTaskResponse])
def list_patrol_tasks(
    river_section_id: Optional[int] = None,
    status: Optional[str] = None,
    start_date: Optional[date] = None,
    end_date: Optional[date] = None,
    skip: int = Query(0, ge=0),
    limit: int = Query(20, ge=1, le=100),
    db: Session = Depends(get_db)
):
    query = db.query(PatrolTask)
    if river_section_id:
        query = query.filter(PatrolTask.river_section_id == river_section_id)
    if status:
        query = query.filter(PatrolTask.status == status)
    if start_date:
        query = query.filter(PatrolTask.task_date >= start_date)
    if end_date:
        query = query.filter(PatrolTask.task_date <= end_date)
    
    return query.order_by(PatrolTask.task_date.desc()).offset(skip).limit(limit).all()


@router.get("/patrol-tasks/{task_id}", response_model=PatrolTaskResponse)
def get_patrol_task(task_id: int, db: Session = Depends(get_db)):
    task = db.query(PatrolTask).filter(PatrolTask.id == task_id).first()
    if not task:
        raise HTTPException(status_code=404, detail="巡查任务不存在")
    return task


@router.put("/patrol-tasks/{task_id}", response_model=PatrolTaskResponse)
def update_patrol_task(
    task_id: int,
    update_data: PatrolTaskUpdate,
    db: Session = Depends(get_db)
):
    task = db.query(PatrolTask).filter(PatrolTask.id == task_id).first()
    if not task:
        raise HTTPException(status_code=404, detail="巡查任务不存在")
    
    update_dict = update_data.model_dump(exclude_unset=True)
    for key, value in update_dict.items():
        setattr(task, key, value)
    
    if update_dict.get("status") == "completed":
        task.completed_at = datetime.utcnow()
    
    db.commit()
    db.refresh(task)
    return task


@router.post("/monthly-reports/generate")
def generate_report(
    year: int = Query(...),
    month: int = Query(..., ge=1, le=12),
    db: Session = Depends(get_db)
):
    report = create_monthly_report(db, year, month)
    return {"report_id": report.id, "content": report.content}


@router.post("/monthly-reports/generate-if-needed")
def generate_report_if_needed(db: Session = Depends(get_db)):
    report = execute_monthly_report_if_needed(db)
    if report:
        return {
            "generated": True,
            "report_id": report.id,
            "report_period": f"{report.report_year}年{report.report_month}月"
        }
    return {"generated": False, "message": "非1-3日，无需生成月报"}


@router.get("/monthly-reports", response_model=List[MonthlyReportResponse])
def list_monthly_reports(
    year: Optional[int] = None,
    month: Optional[int] = None,
    status: Optional[str] = None,
    skip: int = Query(0, ge=0),
    limit: int = Query(20, ge=1, le=100),
    db: Session = Depends(get_db)
):
    query = db.query(MonthlyReport)
    if year:
        query = query.filter(MonthlyReport.report_year == year)
    if month:
        query = query.filter(MonthlyReport.report_month == month)
    if status:
        query = query.filter(MonthlyReport.status == status)
    
    return query.order_by(
        MonthlyReport.report_year.desc(),
        MonthlyReport.report_month.desc()
    ).offset(skip).limit(limit).all()


@router.get("/monthly-reports/{report_id}", response_model=MonthlyReportResponse)
def get_monthly_report(report_id: int, db: Session = Depends(get_db)):
    report = db.query(MonthlyReport).filter(MonthlyReport.id == report_id).first()
    if not report:
        raise HTTPException(status_code=404, detail="月报不存在")
    return report


@router.post("/monthly-reports/{report_id}/review")
def review_monthly_report(
    report_id: int,
    review_comment: Optional[str] = None,
    reviewer: str = "reviewer",
    db: Session = Depends(get_db)
):
    report = db.query(MonthlyReport).filter(MonthlyReport.id == report_id).first()
    if not report:
        raise HTTPException(status_code=404, detail="月报不存在")
    if report.status != "pending":
        raise HTTPException(status_code=400, detail="月报已审阅")
    
    report.status = "reviewed"
    report.reviewer = reviewer
    report.review_time = datetime.utcnow()
    report.review_comment = review_comment
    db.commit()
    db.refresh(report)
    
    return {"message": "月报已审阅", "report_id": report.id}
