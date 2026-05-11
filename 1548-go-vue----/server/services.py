from sqlalchemy.orm import Session
from sqlalchemy import func, and_
from datetime import datetime, date, timedelta
from typing import Optional, List, Dict, Any
from dateutil.relativedelta import relativedelta

from .models import (
    License, LicenseStatus, MiningReport, WeighbridgeRecord,
    MiningStatistics, LawEnforcement, PenaltyType, PenaltyStatus,
    RiverSection, PatrolTask, MonthlyReport, AuditLog
)
from .schemas import (
    LicenseCreate, MiningReportCreate, WeighbridgeRecordCreate,
    LawEnforcementCreate, RiverSectionCreate, PatrolTaskCreate
)


def log_audit(
    db: Session,
    action: str,
    entity_type: str,
    entity_id: int,
    operator: str = "system",
    details: Optional[str] = None
) -> AuditLog:
    log = AuditLog(
        action=action,
        entity_type=entity_type,
        entity_id=entity_id,
        details=details,
        operator=operator
    )
    db.add(log)
    db.commit()
    db.refresh(log)
    return log


def generate_license_number() -> str:
    now = datetime.now()
    prefix = "SC"
    date_str = now.strftime("%Y%m%d")
    random_suffix = now.strftime("%H%M%S")
    return f"{prefix}-{date_str}-{random_suffix}"


def create_license(db: Session, license_data: LicenseCreate, operator: str = "system") -> License:
    db_license = License(**license_data.model_dump())
    db.add(db_license)
    db.commit()
    db.refresh(db_license)
    log_audit(db, "CREATE", "License", db_license.id, operator, "创建采砂许可证申请")
    return db_license


def submit_license(db: Session, license_id: int, operator: str = "system") -> Optional[License]:
    license = db.query(License).filter(License.id == license_id).first()
    if not license or license.status != LicenseStatus.DRAFT:
        return None
    license.status = LicenseStatus.SUBMITTED
    license.submit_time = datetime.utcnow()
    db.commit()
    db.refresh(license)
    log_audit(db, "SUBMIT", "License", license.id, operator, "提交采砂许可证申请")
    return license


def accept_license(db: Session, license_id: int, operator: str = "system") -> Optional[License]:
    license = db.query(License).filter(License.id == license_id).first()
    if not license or license.status != LicenseStatus.SUBMITTED:
        return None
    license.status = LicenseStatus.ACCEPTED
    license.accept_time = datetime.utcnow()
    db.commit()
    db.refresh(license)
    log_audit(db, "ACCEPT", "License", license.id, operator, "受理采砂许可证申请")
    return license


def inspect_license(
    db: Session,
    license_id: int,
    inspect_result: str,
    passed: bool = True,
    operator: str = "system"
) -> Optional[License]:
    license = db.query(License).filter(License.id == license_id).first()
    if not license or license.status != LicenseStatus.ACCEPTED:
        return None
    license.status = LicenseStatus.INSPECTED
    license.inspect_time = datetime.utcnow()
    license.inspect_result = inspect_result
    db.commit()
    db.refresh(license)
    log_audit(db, "INSPECT", "License", license.id, operator, f"核查完成：{'通过' if passed else '不通过'}")
    return license


def publish_license(
    db: Session,
    license_id: int,
    publish_days: int = 7,
    operator: str = "system"
) -> Optional[License]:
    license = db.query(License).filter(License.id == license_id).first()
    if not license or license.status != LicenseStatus.INSPECTED:
        return None
    license.status = LicenseStatus.PUBLISHED
    license.publish_time = datetime.utcnow()
    license.publish_end_time = datetime.utcnow() + timedelta(days=publish_days)
    db.commit()
    db.refresh(license)
    log_audit(db, "PUBLISH", "License", license.id, operator, f"公示期{publish_days}天")
    return license


def issue_license(db: Session, license_id: int, operator: str = "system") -> Optional[License]:
    license = db.query(License).filter(License.id == license_id).first()
    if not license or license.status != LicenseStatus.PUBLISHED:
        return None
    license.status = LicenseStatus.ISSUED
    license.issue_time = datetime.utcnow()
    license.license_number = generate_license_number()
    db.commit()
    db.refresh(license)
    log_audit(db, "ISSUE", "License", license.id, operator, f"发证，许可证号：{license.license_number}")
    return license


def create_mining_report(db: Session, report_data: MiningReportCreate, operator: str = "system") -> MiningReport:
    db_report = MiningReport(**report_data.model_dump())
    db.add(db_report)
    db.commit()
    db.refresh(db_report)
    log_audit(db, "CREATE", "MiningReport", db_report.id, operator, "提交采砂作业报备")
    return db_report


def create_weighbridge_record(db: Session, record_data: WeighbridgeRecordCreate, operator: str = "system") -> WeighbridgeRecord:
    db_record = WeighbridgeRecord(**record_data.model_dump())
    db.add(db_record)
    db.commit()
    db.refresh(db_record)
    log_audit(db, "CREATE", "WeighbridgeRecord", db_record.id, operator, "添加过磅记录")
    update_mining_statistics(db, record_data.license_id, record_data.record_time.year, record_data.record_time.month)
    return db_record


def update_mining_statistics(db: Session, license_id: int, year: int, month: int) -> MiningStatistics:
    stats = db.query(MiningStatistics).filter(
        and_(
            MiningStatistics.license_id == license_id,
            MiningStatistics.year == year,
            MiningStatistics.month == month
        )
    ).first()
    
    monthly_total = db.query(func.sum(WeighbridgeRecord.net_weight)).filter(
        and_(
            WeighbridgeRecord.license_id == license_id,
            func.strftime('%Y', WeighbridgeRecord.record_time) == str(year),
            func.strftime('%m', WeighbridgeRecord.record_time) == str(month).zfill(2)
        )
    ).scalar() or 0.0
    
    if stats:
        stats.total_volume = monthly_total
    else:
        stats = MiningStatistics(
            license_id=license_id,
            year=year,
            month=month,
            total_volume=monthly_total
        )
        db.add(stats)
    
    db.commit()
    db.refresh(stats)
    return stats


def get_yearly_mining_volume(db: Session, license_id: int, year: int) -> float:
    total = db.query(func.sum(MiningStatistics.total_volume)).filter(
        and_(
            MiningStatistics.license_id == license_id,
            MiningStatistics.year == year
        )
    ).scalar() or 0.0
    return total


def check_warning_and_stop(db: Session, license_id: int) -> Dict[str, Any]:
    license = db.query(License).filter(License.id == license_id).first()
    if not license:
        return {"warning": False, "should_stop": False, "message": "许可证不存在"}
    
    current_year = date.today().year
    current_volume = get_yearly_mining_volume(db, license_id, current_year)
    quota = license.annual_quota
    percentage = (current_volume / quota * 100) if quota > 0 else 0
    
    result = {
        "license_id": license.id,
        "license_number": license.license_number,
        "company_name": license.company_name,
        "annual_quota": quota,
        "current_volume": current_volume,
        "percentage": round(percentage, 2),
        "warning": False,
        "should_stop": False,
        "message": ""
    }
    
    if percentage >= 100:
        result["should_stop"] = True
        result["message"] = f"已达到年度控制量100%，本年度作业已停止"
        if license.status == LicenseStatus.ISSUED:
            license.status = LicenseStatus.EXPIRED
            log_audit(db, "STOP", "License", license.id, "system", "达到年度控制量100%，自动停止作业")
            db.commit()
    elif percentage >= 90:
        result["warning"] = True
        result["message"] = f"已达到年度控制量{round(percentage, 2)}%，请注意控制开采"
        
        stats = db.query(MiningStatistics).filter(
            and_(
                MiningStatistics.license_id == license_id,
                MiningStatistics.year == current_year
            )
        ).all()
        for stat in stats:
            if not stat.warning_sent:
                stat.warning_sent = True
        
        db.commit()
    
    return result


def create_law_enforcement(db: Session, enforcement_data: LawEnforcementCreate, operator: str = "system") -> LawEnforcement:
    db_enforcement = LawEnforcement(**enforcement_data.model_dump())
    db.add(db_enforcement)
    db.commit()
    db.refresh(db_enforcement)
    
    if db_enforcement.penalty_type in [PenaltyType.REVOKE, PenaltyType.SUSPEND]:
        update_license_status_by_penalty(db, db_enforcement, operator)
    
    log_audit(db, "CREATE", "LawEnforcement", db_enforcement.id, operator, f"创建执法记录，处罚类型：{db_enforcement.penalty_type.value}")
    return db_enforcement


def update_license_status_by_penalty(db: Session, enforcement: LawEnforcement, operator: str = "system"):
    if not enforcement.license_id:
        return
    
    license = db.query(License).filter(License.id == enforcement.license_id).first()
    if not license:
        return
    
    if enforcement.penalty_type == PenaltyType.REVOKE:
        license.status = LicenseStatus.REVOKED
        log_audit(db, "REVOKE", "License", license.id, operator, "因违法被吊销许可证")
    elif enforcement.penalty_type == PenaltyType.SUSPEND:
        license.status = LicenseStatus.SUSPENDED
        log_audit(db, "SUSPEND", "License", license.id, operator, "因违法被暂扣许可证")
    
    sync_penalty_status(db, license.id, operator)
    db.commit()


def sync_penalty_status(db: Session, license_id: int, operator: str = "system"):
    license = db.query(License).filter(License.id == license_id).first()
    if not license:
        return
    
    if license.status in [LicenseStatus.REVOKED, LicenseStatus.SUSPENDED]:
        pending_enforcements = db.query(LawEnforcement).filter(
            and_(
                LawEnforcement.license_id == license_id,
                LawEnforcement.penalty_status == PenaltyStatus.PENDING
            )
        ).all()
        
        for enf in pending_enforcements:
            if license.status == LicenseStatus.REVOKED:
                enf.penalty_status = PenaltyStatus.EXECUTED
                enf.penalty_time = datetime.utcnow()
            elif license.status == LicenseStatus.SUSPENDED and enf.penalty_type == PenaltyType.SUSPEND:
                enf.penalty_status = PenaltyStatus.EXECUTED
                enf.penalty_time = datetime.utcnow()
        
        for enf in pending_enforcements:
            log_audit(db, "SYNC", "LawEnforcement", enf.id, operator, f"处罚执行状态同步为已执行")
        
        db.commit()


def get_company_penalty_count(db: Session, company_identifier: str, year: int) -> int:
    count = db.query(func.count(LawEnforcement.id)).filter(
        and_(
            LawEnforcement.company_identifier == company_identifier,
            func.strftime('%Y', LawEnforcement.violation_time) == str(year)
        )
    ).scalar() or 0
    return count


def check_annual_review_eligibility(db: Session, license_id: int) -> Dict[str, Any]:
    license = db.query(License).filter(License.id == license_id).first()
    if not license:
        return {"eligible": True, "reason": "许可证不存在"}
    
    current_year = date.today().year
    
    penalties = db.query(LawEnforcement).filter(
        and_(
            LawEnforcement.company_name == license.company_name,
            func.strftime('%Y', LawEnforcement.violation_time) == str(current_year)
        )
    ).all()
    
    penalty_count = len(penalties)
    
    if penalty_count >= 2:
        return {
            "eligible": False,
            "reason": f"本年度已被处罚{penalty_count}次，年审不予通过",
            "penalty_count": penalty_count
        }
    
    return {"eligible": True, "reason": "符合年审条件", "penalty_count": penalty_count}


def create_river_section(db: Session, section_data: RiverSectionCreate, operator: str = "system") -> RiverSection:
    db_section = RiverSection(**section_data.model_dump())
    db.add(db_section)
    db.commit()
    db.refresh(db_section)
    log_audit(db, "CREATE", "RiverSection", db_section.id, operator, "创建河段")
    return db_section


def generate_patrol_tasks(db: Session, operator: str = "system") -> List[PatrolTask]:
    sections = db.query(RiverSection).all()
    today = date.today()
    created_tasks = []
    
    for section in sections:
        existing_task = db.query(PatrolTask).filter(
            and_(
                PatrolTask.river_section_id == section.id,
                PatrolTask.task_date >= today
            )
        ).first()
        
        if existing_task:
            continue
        
        if section.importance.value == "important":
            task_date = today + timedelta(days=1)
        else:
            task_date = today + timedelta(days=7)
        
        existing_this_week = db.query(PatrolTask).filter(
            and_(
                PatrolTask.river_section_id == section.id,
                PatrolTask.task_date >= today - timedelta(days=7),
                PatrolTask.task_date < today
            )
        ).first()
        
        if existing_this_week and section.importance.value != "important":
            continue
        
        task = PatrolTask(
            river_section_id=section.id,
            task_date=task_date
        )
        db.add(task)
        created_tasks.append(task)
    
    db.commit()
    for task in created_tasks:
        log_audit(db, "GENERATE", "PatrolTask", task.id, operator, "自动生成巡查任务")
    
    return created_tasks


def create_monthly_report(db: Session, year: int, month: int, operator: str = "system") -> MonthlyReport:
    existing = db.query(MonthlyReport).filter(
        and_(
            MonthlyReport.report_year == year,
            MonthlyReport.report_month == month
        )
    ).first()
    
    if existing:
        return existing
    
    licenses_count = db.query(func.count(License.id)).filter(
        License.status == LicenseStatus.ISSUED
    ).scalar() or 0
    
    total_mining = db.query(func.sum(MiningStatistics.total_volume)).filter(
        and_(
            MiningStatistics.year == year,
            MiningStatistics.month == month
        )
    ).scalar() or 0.0
    
    penalties_count = db.query(func.count(LawEnforcement.id)).filter(
        and_(
            func.strftime('%Y', LawEnforcement.created_at) == str(year),
            func.strftime('%m', LawEnforcement.created_at) == str(month).zfill(2)
        )
    ).scalar() or 0
    
    content = f"""
    {year}年{month}月河道采砂管理月报
    
    一、许可证发放情况
    - 有效许可证数量：{licenses_count}个
    
    二、采砂量统计
    - 本月总开采量：{total_mining}吨
    
    三、执法情况
    - 本月执法案件：{penalties_count}起
    """
    
    report = MonthlyReport(
        report_year=year,
        report_month=month,
        content=content
    )
    db.add(report)
    db.commit()
    db.refresh(report)
    log_audit(db, "GENERATE", "MonthlyReport", report.id, operator, f"自动生成{year}年{month}月月报")
    
    return report


def execute_monthly_report_if_needed(db: Session, operator: str = "system") -> Optional[MonthlyReport]:
    today = date.today()
    if today.day <= 3:
        last_month = today.replace(day=1) - relativedelta(months=1)
        return create_monthly_report(db, last_month.year, last_month.month, operator)
    return None
