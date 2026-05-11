from datetime import datetime, date, timedelta
from typing import Optional, List
from sqlalchemy.orm import Session
from fastapi import HTTPException
from . import models, schemas
from .config import settings
from . import audit_service


def add_business_days(start_date: date, days: int) -> date:
    current = start_date
    added = 0
    while added < days:
        current += timedelta(days=1)
        if current.weekday() < 5:
            added += 1
    return current


def generate_number(prefix: str, db: Session, model) -> str:
    today = datetime.now()
    date_str = today.strftime("%Y%m%d")
    last_record = db.query(model).order_by(model.id.desc()).first()
    if last_record:
        count = last_record.id + 1
    else:
        count = 1
    return f"{prefix}{date_str}{count:04d}"


def create_vessel(db: Session, vessel: schemas.VesselCreate) -> models.Vessel:
    db_vessel = models.Vessel(**vessel.model_dump())
    db.add(db_vessel)
    db.commit()
    db.refresh(db_vessel)
    audit_service.log_action(db, "CREATE", "Vessel", db_vessel.id, f"创建船舶: {db_vessel.vessel_name}")
    return db_vessel


def get_vessels(db: Session, skip: int = 0, limit: int = 100) -> List[models.Vessel]:
    return db.query(models.Vessel).offset(skip).limit(limit).all()


def get_vessel(db: Session, vessel_id: int) -> Optional[models.Vessel]:
    return db.query(models.Vessel).filter(models.Vessel.id == vessel_id).first()


def get_vessel_detail(db: Session, vessel_id: int) -> Optional[models.Vessel]:
    vessel = get_vessel(db, vessel_id)
    return vessel


def update_vessel(db: Session, vessel_id: int, vessel_update: schemas.VesselUpdate) -> Optional[models.Vessel]:
    db_vessel = get_vessel(db, vessel_id)
    if not db_vessel:
        return None
    
    update_data = vessel_update.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(db_vessel, key, value)
    
    db.commit()
    db.refresh(db_vessel)
    audit_service.log_action(db, "UPDATE", "Vessel", db_vessel.id, f"更新船舶信息: {db_vessel.vessel_name}")
    return db_vessel


def delete_vessel(db: Session, vessel_id: int) -> bool:
    db_vessel = get_vessel(db, vessel_id)
    if not db_vessel:
        return False
    
    db.delete(db_vessel)
    db.commit()
    audit_service.log_action(db, "DELETE", "Vessel", vessel_id, f"删除船舶: {db_vessel.vessel_name}")
    return True


def has_active_permit(db: Session, vessel_id: int, permit_type: str) -> bool:
    today = date.today()
    active_permits = db.query(models.Permit).filter(
        models.Permit.vessel_id == vessel_id,
        models.Permit.permit_type == permit_type,
        models.Permit.status == models.PermitStatus.APPROVED.value,
        models.Permit.effective_date <= today,
        (models.Permit.expiry_date >= today) | (models.Permit.expiry_date.is_(None))
    ).first()
    return active_permits is not None


def create_permit(db: Session, permit: schemas.PermitCreate) -> models.Permit:
    min_effective_date = add_business_days(permit.application_date, settings.permit_effective_days)
    if permit.expected_effective_date < min_effective_date:
        raise HTTPException(
            status_code=400,
            detail=f"期望生效日期不能早于申请日期加{settings.permit_effective_days}个工作日，最早为 {min_effective_date}"
        )
    
    if permit.permit_type not in settings.permit_types:
        raise HTTPException(
            status_code=400,
            detail=f"许可类型无效，有效类型为: {', '.join(settings.permit_types)}"
        )
    
    if permit.vessel_id and has_active_permit(db, permit.vessel_id, permit.permit_type):
        db_permit = models.Permit(
            **permit.model_dump(),
            application_number=generate_number("XK", db, models.Permit),
            status=models.PermitStatus.REJECTED.value,
            reject_reason="同一时间已有有效的同类型许可"
        )
        db.add(db_permit)
        db.commit()
        db.refresh(db_permit)
        audit_service.log_action(db, "CREATE", "Permit", db_permit.id, f"自动驳回许可申请: {db_permit.application_number}，原因: 同一时间已有有效的同类型许可")
        return db_permit
    
    db_permit = models.Permit(
        **permit.model_dump(),
        application_number=generate_number("XK", db, models.Permit),
        status=models.PermitStatus.PENDING.value
    )
    db.add(db_permit)
    db.commit()
    db.refresh(db_permit)
    audit_service.log_action(db, "CREATE", "Permit", db_permit.id, f"创建许可申请: {db_permit.application_number}")
    return db_permit


def get_permits(db: Session, skip: int = 0, limit: int = 100) -> List[models.Permit]:
    return db.query(models.Permit).offset(skip).limit(limit).all()


def get_permit(db: Session, permit_id: int) -> Optional[models.Permit]:
    return db.query(models.Permit).filter(models.Permit.id == permit_id).first()


def get_permits_by_vessel(db: Session, vessel_id: int) -> List[models.Permit]:
    return db.query(models.Permit).filter(models.Permit.vessel_id == vessel_id).all()


def update_permit_status(db: Session, permit_id: int, permit_update: schemas.PermitUpdate) -> Optional[models.Permit]:
    db_permit = get_permit(db, permit_id)
    if not db_permit:
        return None
    
    update_data = permit_update.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(db_permit, key, value)
    
    db.commit()
    db.refresh(db_permit)
    audit_service.log_action(db, "UPDATE", "Permit", db_permit.id, f"更新许可状态: {db_permit.application_number} -> {db_permit.status}")
    return db_permit


def check_focus_attention(db: Session, vessel_id: int) -> bool:
    today = date.today()
    thirty_days_ago = today - timedelta(days=settings.focus_attention_days)
    
    unqualified_count = db.query(models.Inspection).filter(
        models.Inspection.vessel_id == vessel_id,
        models.Inspection.result == models.InspectionResult.UNQUALIFIED.value,
        models.Inspection.inspection_date >= thirty_days_ago
    ).count()
    
    return unqualified_count >= settings.focus_attention_checks


def create_rectification_from_inspection(db: Session, inspection: models.Inspection) -> models.Rectification:
    rectification = models.Rectification(
        vessel_id=inspection.vessel_id,
        inspection_id=inspection.id,
        rectification_number=generate_number("ZG", db, models.Rectification),
        rectification_content=inspection.unqualified_items or "检查不合格项",
        created_date=date.today(),
        status=models.RectificationStatus.PENDING.value
    )
    db.add(rectification)
    db.commit()
    db.refresh(rectification)
    audit_service.log_action(db, "CREATE", "Rectification", rectification.id, f"从检查自动生成整改记录: {rectification.rectification_number}")
    return rectification


def create_inspection(db: Session, inspection: schemas.InspectionCreate) -> models.Inspection:
    db_inspection = models.Inspection(
        **inspection.model_dump(),
        inspection_number=generate_number("JC", db, models.Inspection)
    )
    db.add(db_inspection)
    db.commit()
    db.refresh(db_inspection)
    
    audit_service.log_action(db, "CREATE", "Inspection", db_inspection.id, f"创建执法检查: {db_inspection.inspection_number}")
    
    if db_inspection.result == models.InspectionResult.UNQUALIFIED.value:
        create_rectification_from_inspection(db, db_inspection)
        
        if db_inspection.vessel_id and check_focus_attention(db, db_inspection.vessel_id):
            vessel = get_vessel(db, db_inspection.vessel_id)
            if vessel:
                vessel.is_focus_attention = True
                db.commit()
                audit_service.log_action(db, "UPDATE", "Vessel", vessel.id, f"标记船舶为重点关注: {vessel.vessel_name}")
    
    return db_inspection


def get_inspections(db: Session, skip: int = 0, limit: int = 100) -> List[models.Inspection]:
    return db.query(models.Inspection).offset(skip).limit(limit).all()


def get_inspection(db: Session, inspection_id: int) -> Optional[models.Inspection]:
    return db.query(models.Inspection).filter(models.Inspection.id == inspection_id).first()


def get_inspections_by_vessel(db: Session, vessel_id: int) -> List[models.Inspection]:
    return db.query(models.Inspection).filter(models.Inspection.vessel_id == vessel_id).all()


def validate_fine_amount(amount: float) -> bool:
    if amount < settings.min_fine or amount > settings.max_fine:
        return False
    if amount % settings.fine_increment != 0:
        return False
    return True


def create_penalty(db: Session, penalty: schemas.PenaltyCreate) -> models.Penalty:
    if not validate_fine_amount(penalty.fine_amount):
        raise HTTPException(
            status_code=400,
            detail=f"罚款金额必须在{settings.min_fine}到{settings.max_fine}元之间，且必须是{settings.fine_increment}的整数倍"
        )
    
    db_penalty = models.Penalty(
        **penalty.model_dump(),
        penalty_number=generate_number("CF", db, models.Penalty),
        status=models.PenaltyStatus.PENDING.value
    )
    db.add(db_penalty)
    db.commit()
    db.refresh(db_penalty)
    audit_service.log_action(db, "CREATE", "Penalty", db_penalty.id, f"创建行政处罚: {db_penalty.penalty_number}，罚款金额: {db_penalty.fine_amount}元")
    return db_penalty


def get_penalties(db: Session, skip: int = 0, limit: int = 100) -> List[models.Penalty]:
    return db.query(models.Penalty).offset(skip).limit(limit).all()


def get_penalty(db: Session, penalty_id: int) -> Optional[models.Penalty]:
    return db.query(models.Penalty).filter(models.Penalty.id == penalty_id).first()


def get_penalties_by_vessel(db: Session, vessel_id: int) -> List[models.Penalty]:
    return db.query(models.Penalty).filter(models.Penalty.vessel_id == vessel_id).all()


def check_and_update_overdue(db: Session, penalty_id: int) -> Optional[models.Penalty]:
    db_penalty = get_penalty(db, penalty_id)
    if not db_penalty:
        return None
    
    if db_penalty.status in [models.PenaltyStatus.PENDING.value, models.PenaltyStatus.PROCESSING.value]:
        today = date.today()
        deadline_with_overdue = add_business_days(db_penalty.deadline_date, settings.overdue_days)
        if today > deadline_with_overdue and db_penalty.payment_date is None:
            db_penalty.status = models.PenaltyStatus.OVERDUE.value
            db.commit()
            db.refresh(db_penalty)
            audit_service.log_action(db, "UPDATE", "Penalty", db_penalty.id, f"行政处罚标记为逾期: {db_penalty.penalty_number}")
    
    return db_penalty


def update_penalty(db: Session, penalty_id: int, penalty_update: schemas.PenaltyUpdate) -> Optional[models.Penalty]:
    db_penalty = get_penalty(db, penalty_id)
    if not db_penalty:
        return None
    
    update_data = penalty_update.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(db_penalty, key, value)
    
    if db_penalty.payment_date and db_penalty.status != models.PenaltyStatus.COMPLETED.value:
        db_penalty.status = models.PenaltyStatus.COMPLETED.value
    
    db.commit()
    db.refresh(db_penalty)
    audit_service.log_action(db, "UPDATE", "Penalty", db_penalty.id, f"更新行政处罚: {db_penalty.penalty_number}")
    return db_penalty


def get_rectifications(db: Session, skip: int = 0, limit: int = 100) -> List[models.Rectification]:
    return db.query(models.Rectification).offset(skip).limit(limit).all()


def get_rectification(db: Session, rectification_id: int) -> Optional[models.Rectification]:
    return db.query(models.Rectification).filter(models.Rectification.id == rectification_id).first()


def get_rectifications_by_vessel(db: Session, vessel_id: int) -> List[models.Rectification]:
    return db.query(models.Rectification).filter(models.Rectification.vessel_id == vessel_id).all()


def create_rectification(db: Session, rectification: schemas.RectificationCreate) -> models.Rectification:
    db_rectification = models.Rectification(
        **rectification.model_dump(),
        rectification_number=generate_number("ZG", db, models.Rectification),
        created_date=date.today(),
        status=models.RectificationStatus.PENDING.value
    )
    db.add(db_rectification)
    db.commit()
    db.refresh(db_rectification)
    audit_service.log_action(db, "CREATE", "Rectification", db_rectification.id, f"创建整改记录: {db_rectification.rectification_number}")
    return db_rectification


def submit_rectification(db: Session, rectification_id: int, data: schemas.RectificationSubmit) -> Optional[models.Rectification]:
    db_rectification = get_rectification(db, rectification_id)
    if not db_rectification:
        return None
    
    if db_rectification.status not in [models.RectificationStatus.PENDING.value, models.RectificationStatus.REJECTED.value]:
        raise HTTPException(
            status_code=400,
            detail="只能在待整改或复查不通过状态下提交整改"
        )
    
    db_rectification.rectification_measure = data.rectification_measure
    db_rectification.rectification_date = data.rectification_date or date.today()
    db_rectification.status = models.RectificationStatus.RECHECKING.value
    
    db.commit()
    db.refresh(db_rectification)
    audit_service.log_action(db, "UPDATE", "Rectification", db_rectification.id, f"提交整改: {db_rectification.rectification_number}")
    return db_rectification


def recheck_rectification(db: Session, rectification_id: int, data: schemas.RectificationRecheck, passed: bool) -> Optional[models.Rectification]:
    db_rectification = get_rectification(db, rectification_id)
    if not db_rectification:
        return None
    
    if db_rectification.status != models.RectificationStatus.RECHECKING.value:
        raise HTTPException(
            status_code=400,
            detail="只能在待复查状态下进行复查"
        )
    
    db_rectification.recheck_date = data.recheck_date
    db_rectification.recheck_result = data.recheck_result
    db_rectification.recheck_count += 1
    
    if passed:
        db_rectification.status = models.RectificationStatus.CLOSED.value
    else:
        db_rectification.status = models.RectificationStatus.REJECTED.value
    
    db.commit()
    db.refresh(db_rectification)
    audit_service.log_action(db, "UPDATE", "Rectification", db_rectification.id, f"复查整改: {db_rectification.rectification_number}，结果: {'通过' if passed else '不通过'}")
    return db_rectification


def update_rectification(db: Session, rectification_id: int, rectification_update: schemas.RectificationUpdate) -> Optional[models.Rectification]:
    db_rectification = get_rectification(db, rectification_id)
    if not db_rectification:
        return None
    
    update_data = rectification_update.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(db_rectification, key, value)
    
    db.commit()
    db.refresh(db_rectification)
    audit_service.log_action(db, "UPDATE", "Rectification", db_rectification.id, f"更新整改记录: {db_rectification.rectification_number}")
    return db_rectification
