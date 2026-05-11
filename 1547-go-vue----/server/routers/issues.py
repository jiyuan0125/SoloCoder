from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List, Optional
from datetime import date
from server.database import get_db
from server.models import Issue, River, Patrol, IssueStatus
from server.schemas import (
    IssueCreate, IssueUpdate, Issue as IssueSchema, IssueDetail, Notification
)
from server.services import (
    calculate_deadline, check_and_escalate_issue,
    start_rectification, submit_for_review, start_review,
    approve_review, reject_review, check_all_overdue_issues
)

router = APIRouter(prefix="/issues", tags=["issues"])

@router.post("/check-overdue")
def check_overdue(db: Session = Depends(get_db)):
    check_all_overdue_issues(db)
    return {"message": "已检查所有超期问题"}

@router.get("/", response_model=List[IssueSchema])
def list_issues(
    status: Optional[IssueStatus] = None,
    db: Session = Depends(get_db)
):
    query = db.query(Issue)
    if status:
        query = query.filter(Issue.status == status)
    return query.order_by(Issue.created_at.desc()).all()

@router.post("/", response_model=IssueSchema)
def create_issue(issue: IssueCreate, db: Session = Depends(get_db)):
    river = db.query(River).filter(River.id == issue.river_id).first()
    if not river:
        raise HTTPException(status_code=400, detail="河流不存在")
    if issue.patrol_id:
        patrol = db.query(Patrol).filter(Patrol.id == issue.patrol_id).first()
        if not patrol:
            raise HTTPException(status_code=400, detail="关联的巡河记录不存在")
    deadline = calculate_deadline(issue.severity)
    db_issue = Issue(**issue.model_dump(), deadline=deadline)
    db.add(db_issue)
    db.commit()
    db.refresh(db_issue)
    return db_issue

@router.get("/{issue_id}", response_model=IssueDetail)
def get_issue(issue_id: int, db: Session = Depends(get_db)):
    issue = db.query(Issue).filter(Issue.id == issue_id).first()
    if not issue:
        raise HTTPException(status_code=404, detail="问题不存在")
    issue = check_and_escalate_issue(db, issue)
    return issue

@router.put("/{issue_id}", response_model=IssueSchema)
def update_issue(issue_id: int, issue: IssueUpdate, db: Session = Depends(get_db)):
    db_issue = db.query(Issue).filter(Issue.id == issue_id).first()
    if not db_issue:
        raise HTTPException(status_code=404, detail="问题不存在")
    update_data = issue.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(db_issue, key, value)
    db.commit()
    db.refresh(db_issue)
    return db_issue

@router.post("/{issue_id}/start-rectification", response_model=IssueSchema)
def api_start_rectification(issue_id: int, db: Session = Depends(get_db)):
    issue = db.query(Issue).filter(Issue.id == issue_id).first()
    if not issue:
        raise HTTPException(status_code=404, detail="问题不存在")
    return start_rectification(db, issue)

@router.post("/{issue_id}/submit-review", response_model=IssueSchema)
def api_submit_review(issue_id: int, rectification_note: str, db: Session = Depends(get_db)):
    issue = db.query(Issue).filter(Issue.id == issue_id).first()
    if not issue:
        raise HTTPException(status_code=404, detail="问题不存在")
    return submit_for_review(db, issue, rectification_note)

@router.post("/{issue_id}/start-review", response_model=IssueSchema)
def api_start_review(issue_id: int, db: Session = Depends(get_db)):
    issue = db.query(Issue).filter(Issue.id == issue_id).first()
    if not issue:
        raise HTTPException(status_code=404, detail="问题不存在")
    return start_review(db, issue)

@router.post("/{issue_id}/approve", response_model=IssueSchema)
def api_approve_review(issue_id: int, review_note: str, db: Session = Depends(get_db)):
    issue = db.query(Issue).filter(Issue.id == issue_id).first()
    if not issue:
        raise HTTPException(status_code=404, detail="问题不存在")
    return approve_review(db, issue, review_note)

@router.post("/{issue_id}/reject", response_model=IssueSchema)
def api_reject_review(issue_id: int, review_note: str, db: Session = Depends(get_db)):
    issue = db.query(Issue).filter(Issue.id == issue_id).first()
    if not issue:
        raise HTTPException(status_code=404, detail="问题不存在")
    return reject_review(db, issue, review_note)

@router.delete("/{issue_id}")
def delete_issue(issue_id: int, db: Session = Depends(get_db)):
    db_issue = db.query(Issue).filter(Issue.id == issue_id).first()
    if not db_issue:
        raise HTTPException(status_code=404, detail="问题不存在")
    db.delete(db_issue)
    db.commit()
    return {"message": "删除成功"}
