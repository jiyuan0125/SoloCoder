from datetime import date, timedelta
from typing import Dict
from sqlalchemy.orm import Session
from server.models import (
    Issue, IssueSeverity, IssueStatus, RiverLevel, 
    Notification, MonthlyReport, Patrol, River, RiverKeeper
)
from server.schemas import PatrolFrequency

PATROL_FREQUENCIES: Dict[RiverLevel, PatrolFrequency] = {
    RiverLevel.PROVINCIAL: PatrolFrequency(
        level=RiverLevel.PROVINCIAL,
        frequency="每季度",
        days=90
    ),
    RiverLevel.MUNICIPAL: PatrolFrequency(
        level=RiverLevel.MUNICIPAL,
        frequency="每两月",
        days=60
    ),
    RiverLevel.COUNTY: PatrolFrequency(
        level=RiverLevel.COUNTY,
        frequency="每月",
        days=30
    ),
    RiverLevel.TOWNSHIP: PatrolFrequency(
        level=RiverLevel.TOWNSHIP,
        frequency="每半月",
        days=15
    )
}

ISSUE_DEADLINES: Dict[IssueSeverity, int] = {
    IssueSeverity.NORMAL: 7,
    IssueSeverity.SERIOUS: 15,
    IssueSeverity.SEVERE: 30
}

NEXT_SEVERITY: Dict[IssueSeverity, IssueSeverity] = {
    IssueSeverity.NORMAL: IssueSeverity.SERIOUS,
    IssueSeverity.SERIOUS: IssueSeverity.SEVERE,
    IssueSeverity.SEVERE: IssueSeverity.SEVERE
}

NEXT_LEVEL: Dict[RiverLevel, RiverLevel] = {
    RiverLevel.TOWNSHIP: RiverLevel.COUNTY,
    RiverLevel.COUNTY: RiverLevel.MUNICIPAL,
    RiverLevel.MUNICIPAL: RiverLevel.PROVINCIAL,
    RiverLevel.PROVINCIAL: RiverLevel.PROVINCIAL
}

def calculate_deadline(severity: IssueSeverity, start_date: date = None) -> date:
    if start_date is None:
        start_date = date.today()
    days = ISSUE_DEADLINES.get(severity, 7)
    return start_date + timedelta(days=days)

def check_and_escalate_issue(db: Session, issue: Issue) -> Issue:
    today = date.today()
    if issue.status in [IssueStatus.CLOSED]:
        return issue
    if issue.deadline < today and not issue.escalated:
        new_severity = NEXT_SEVERITY.get(issue.severity, issue.severity)
        if new_severity != issue.severity:
            issue.severity = new_severity
            issue.deadline = calculate_deadline(new_severity, today)
            issue.escalated = True
            keeper_level = get_issue_keeper_level(db, issue)
            notify_level = NEXT_LEVEL.get(keeper_level, keeper_level)
            notification = Notification(
                issue_id=issue.id,
                message=f"问题ID:{issue.id} 已超期，问题等级已升级为 {new_severity.value}。原截止日期: {issue.deadline}",
                recipient_level=notify_level
            )
            db.add(notification)
            db.commit()
            db.refresh(issue)
    return issue

def get_issue_keeper_level(db: Session, issue: Issue) -> RiverLevel:
    if issue.patrol_id:
        patrol = db.query(Patrol).filter(Patrol.id == issue.patrol_id).first()
        if patrol and patrol.keeper_id:
            keeper = db.query(RiverKeeper).filter(RiverKeeper.id == patrol.keeper_id).first()
            if keeper:
                return keeper.level
    return RiverLevel.TOWNSHIP

def start_rectification(db: Session, issue: Issue) -> Issue:
    if issue.status == IssueStatus.DISCOVERED:
        issue.status = IssueStatus.RECTIFYING
        db.commit()
        db.refresh(issue)
    return issue

def submit_for_review(db: Session, issue: Issue, rectification_note: str) -> Issue:
    if issue.status == IssueStatus.RECTIFYING:
        issue.status = IssueStatus.PENDING_REVIEW
        issue.rectification_note = rectification_note
        issue.rectification_date = date.today()
        db.commit()
        db.refresh(issue)
    return issue

def start_review(db: Session, issue: Issue) -> Issue:
    if issue.status == IssueStatus.PENDING_REVIEW:
        issue.status = IssueStatus.REVIEWING
        db.commit()
        db.refresh(issue)
    return issue

def approve_review(db: Session, issue: Issue, review_note: str) -> Issue:
    if issue.status == IssueStatus.REVIEWING:
        issue.status = IssueStatus.CLOSED
        issue.review_note = review_note
        issue.review_date = date.today()
        db.commit()
        db.refresh(issue)
    return issue

def reject_review(db: Session, issue: Issue, review_note: str) -> Issue:
    if issue.status == IssueStatus.REVIEWING:
        issue.status = IssueStatus.RECTIFYING
        issue.review_note = review_note
        issue.review_date = date.today()
        issue.escalated = False
        db.commit()
        db.refresh(issue)
    return issue

def get_monthly_report_deadline(report_month: str) -> date:
    year, month = map(int, report_month.split("-"))
    if month == 12:
        year += 1
        month = 1
    else:
        month += 1
    return date(year, month, 5)

def calculate_report_stats(db: Session, river_id: int, report_month: str) -> Dict[str, int]:
    year, month = map(int, report_month.split("-"))
    first_day = date(year, month, 1)
    if month == 12:
        last_day = date(year + 1, 1, 1) - timedelta(days=1)
    else:
        last_day = date(year, month + 1, 1) - timedelta(days=1)
    
    patrol_count = db.query(Patrol).filter(
        Patrol.river_id == river_id,
        Patrol.patrol_date >= first_day,
        Patrol.patrol_date <= last_day
    ).count()
    
    issues = db.query(Issue).filter(
        Issue.river_id == river_id,
        Issue.discovered_date >= first_day,
        Issue.discovered_date <= last_day
    ).all()
    
    issue_count = len(issues)
    resolved_issue_count = sum(1 for i in issues if i.status == IssueStatus.CLOSED)
    
    return {
        "patrol_count": patrol_count,
        "issue_count": issue_count,
        "resolved_issue_count": resolved_issue_count
    }

def check_all_overdue_issues(db: Session):
    issues = db.query(Issue).filter(
        Issue.status != IssueStatus.CLOSED,
        Issue.escalated == False
    ).all()
    for issue in issues:
        check_and_escalate_issue(db, issue)
