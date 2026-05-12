from datetime import datetime, date, timedelta
from typing import Optional, List
from sqlalchemy.orm import Session
from sqlalchemy import and_, or_

from server.models import (
    Ship, License, Inspection, Rectification, Penalty, AuditLog,
    LicenseType, LicenseStatus, InspectionResult, RectificationStatus,
    PenaltyStatus, ActionType
)
from server.schemas import (
    ShipCreate, ShipUpdate, LicenseCreate, LicenseUpdate,
    InspectionCreate, InspectionUpdate, RectificationCreate,
    RectificationUpdate, PenaltyCreate, PenaltyUpdate
)


def add_working_days(start_date: date, days: int) -> date:
    result = start_date
    added = 0
    while added < days:
        result += timedelta(days=1)
        if result.weekday() < 5:
            added += 1
    return result


def add_days_without_weekend(start_date: date, days: int) -> date:
    result = start_date
    added = 0
    while added < days:
        result += timedelta(days=1)
        if result.weekday() < 5:
            added += 1
    return result


def create_audit_log(db: Session, entity_type: str, entity_id: int,
                     action: ActionType, details: Optional[str] = None):
    audit_log = AuditLog(
        entity_type=entity_type,
        entity_id=entity_id,
        action=action,
        details=details
    )
    db.add(audit_log)
    db.flush()


def get_ship(db: Session, ship_id: int) -> Optional[Ship]:
    return db.query(Ship).filter(Ship.id == ship_id).first()


def get_ships(db: Session, skip: int = 0, limit: int = 100) -> List[Ship]:
    return db.query(Ship).offset(skip).limit(limit).all()


def create_ship(db: Session, ship: ShipCreate) -> Ship:
    db_ship = Ship(**ship.model_dump())
    db.add(db_ship)
    db.flush()
    create_audit_log(db, "Ship", db_ship.id, ActionType.CREATE,
                     f"Created ship: {ship.model_dump_json()}")
    return db_ship


def update_ship(db: Session, ship_id: int, ship: ShipUpdate) -> Optional[Ship]:
    db_ship = get_ship(db, ship_id)
    if not db_ship:
        return None
    old_data = {c.name: getattr(db_ship, c.name) for c in Ship.__table__.columns}
    for key, value in ship.model_dump(exclude_unset=True).items():
        setattr(db_ship, key, value)
    db.flush()
    create_audit_log(db, "Ship", ship_id, ActionType.UPDATE,
                     f"Updated ship from {old_data} to {ship.model_dump_json()}")
    return db_ship


def has_active_license(db: Session, ship_id: Optional[int], license_type: LicenseType) -> bool:
    if not ship_id:
        return False
    today = date.today()
    active = db.query(License).filter(
        License.ship_id == ship_id,
        License.license_type == license_type,
        License.status == LicenseStatus.APPROVED,
        License.effective_date <= today,
        or_(License.expiration_date.is_(None), License.expiration_date >= today)
    ).first()
    return active is not None


def get_license(db: Session, license_id: int) -> Optional[License]:
    return db.query(License).filter(License.id == license_id).first()


def get_licenses(db: Session, skip: int = 0, limit: int = 100,
                 ship_id: Optional[int] = None) -> List[License]:
    query = db.query(License)
    if ship_id:
        query = query.filter(License.ship_id == ship_id)
    return query.offset(skip).limit(limit).all()


def create_license(db: Session, license_data: LicenseCreate) -> License:
    if license_data.ship_id and has_active_license(db, license_data.ship_id, license_data.license_type):
        db_license = License(
            **license_data.model_dump(),
            status=LicenseStatus.REJECTED,
            rejection_reason="该船舶已有有效同类型许可"
        )
    else:
        db_license = License(**license_data.model_dump())
    db.add(db_license)
    db.flush()
    create_audit_log(db, "License", db_license.id, ActionType.CREATE,
                     f"Created license: {license_data.model_dump_json()}")
    return db_license


def update_license(db: Session, license_id: int,
                   license_update: LicenseUpdate) -> Optional[License]:
    db_license = get_license(db, license_id)
    if not db_license:
        return None
    old_status = db_license.status
    for key, value in license_update.model_dump(exclude_unset=True).items():
        setattr(db_license, key, value)
    db.flush()
    create_audit_log(db, "License", license_id, ActionType.UPDATE,
                     f"Updated license status from {old_status} to {license_update.status}")
    return db_license


def count_recent_failed_inspections(db: Session, ship_id: Optional[int],
                                    days: int = 30) -> int:
    if not ship_id:
        return 0
    since = date.today() - timedelta(days=days)
    return db.query(Inspection).filter(
        Inspection.ship_id == ship_id,
        Inspection.result == InspectionResult.FAILED,
        Inspection.inspection_date >= since
    ).count()


def is_high_priority_ship(db: Session, ship_id: Optional[int]) -> bool:
    if not ship_id:
        return False
    return count_recent_failed_inspections(db, ship_id, 30) >= 2


def get_inspection(db: Session, inspection_id: int) -> Optional[Inspection]:
    return db.query(Inspection).filter(Inspection.id == inspection_id).first()


def get_inspections(db: Session, skip: int = 0, limit: int = 100,
                    ship_id: Optional[int] = None) -> List[Inspection]:
    query = db.query(Inspection)
    if ship_id:
        query = query.filter(Inspection.ship_id == ship_id)
    return query.offset(skip).limit(limit).all()


def create_inspection(db: Session, inspection: InspectionCreate) -> Inspection:
    db_inspection = Inspection(**inspection.model_dump())
    db.add(db_inspection)
    db.flush()
    create_audit_log(db, "Inspection", db_inspection.id, ActionType.CREATE,
                     f"Created inspection: {inspection.model_dump_json()}")

    if inspection.result == InspectionResult.FAILED:
        rect_deadline = add_working_days(inspection.inspection_date, 7)
        rect = Rectification(
            ship_id=inspection.ship_id,
            ship_temp_identifier=inspection.ship_temp_identifier,
            inspection_id=db_inspection.id,
            requirement=inspection.findings or "需整改检查中发现的问题",
            deadline=rect_deadline,
            status=RectificationStatus.CREATED
        )
        db.add(rect)
        db.flush()
        create_audit_log(db, "Rectification", rect.id, ActionType.CREATE,
                         f"Auto-created rectification for inspection {db_inspection.id}")

    return db_inspection


def update_inspection(db: Session, inspection_id: int,
                      inspection_update: InspectionUpdate) -> Optional[Inspection]:
    db_inspection = get_inspection(db, inspection_id)
    if not db_inspection:
        return None
    for key, value in inspection_update.model_dump(exclude_unset=True).items():
        setattr(db_inspection, key, value)
    db.flush()
    create_audit_log(db, "Inspection", inspection_id, ActionType.UPDATE,
                     f"Updated inspection: {inspection_update.model_dump_json()}")

    if inspection_update.result == InspectionResult.FAILED and not db_inspection.rectification:
        rect_deadline = add_working_days(db_inspection.inspection_date, 7)
        rect = Rectification(
            ship_id=db_inspection.ship_id,
            ship_temp_identifier=db_inspection.ship_temp_identifier,
            inspection_id=db_inspection.id,
            requirement=db_inspection.findings or "需整改检查中发现的问题",
            deadline=rect_deadline,
            status=RectificationStatus.CREATED
        )
        db.add(rect)
        db.flush()
        create_audit_log(db, "Rectification", rect.id, ActionType.CREATE,
                         f"Auto-created rectification for inspection {db_inspection.id}")

    return db_inspection


def get_rectification(db: Session, rect_id: int) -> Optional[Rectification]:
    return db.query(Rectification).filter(Rectification.id == rect_id).first()


def get_rectifications(db: Session, skip: int = 0, limit: int = 100,
                       ship_id: Optional[int] = None) -> List[Rectification]:
    query = db.query(Rectification)
    if ship_id:
        query = query.filter(Rectification.ship_id == ship_id)
    return query.offset(skip).limit(limit).all()


def create_rectification(db: Session, rect: RectificationCreate) -> Rectification:
    db_rect = Rectification(**rect.model_dump())
    db.add(db_rect)
    db.flush()
    create_audit_log(db, "Rectification", db_rect.id, ActionType.CREATE,
                     f"Created rectification: {rect.model_dump_json()}")
    return db_rect


def update_rectification(db: Session, rect_id: int,
                         rect_update: RectificationUpdate) -> Optional[Rectification]:
    db_rect = get_rectification(db, rect_id)
    if not db_rect:
        return None

    new_status = rect_update.status
    current_status = db_rect.status

    if new_status:
        valid_transitions = {
            RectificationStatus.CREATED: [RectificationStatus.RECTIFYING],
            RectificationStatus.RECTIFYING: [RectificationStatus.RECHECKING],
            RectificationStatus.RECHECKING: [
                RectificationStatus.CLOSED,
                RectificationStatus.RECTIFYING
            ],
            RectificationStatus.CLOSED: []
        }
        if new_status not in valid_transitions.get(current_status, []) and new_status != current_status:
            raise ValueError(f"Invalid status transition from {current_status} to {new_status}")

    for key, value in rect_update.model_dump(exclude_unset=True).items():
        setattr(db_rect, key, value)
    db.flush()
    create_audit_log(db, "Rectification", rect_id, ActionType.UPDATE,
                     f"Updated rectification from {current_status} to {new_status}")
    return db_rect


def get_penalty(db: Session, penalty_id: int) -> Optional[Penalty]:
    return db.query(Penalty).filter(Penalty.id == penalty_id).first()


def get_penalties(db: Session, skip: int = 0, limit: int = 100,
                  ship_id: Optional[int] = None) -> List[Penalty]:
    query = db.query(Penalty)
    if ship_id:
        query = query.filter(Penalty.ship_id == ship_id)
    return query.offset(skip).limit(limit).all()


def create_penalty(db: Session, penalty: PenaltyCreate) -> Penalty:
    db_penalty = Penalty(**penalty.model_dump())
    db.add(db_penalty)
    db.flush()
    create_audit_log(db, "Penalty", db_penalty.id, ActionType.CREATE,
                     f"Created penalty: {penalty.model_dump_json()}")
    return db_penalty


def update_penalty(db: Session, penalty_id: int,
                   penalty_update: PenaltyUpdate) -> Optional[Penalty]:
    db_penalty = get_penalty(db, penalty_id)
    if not db_penalty:
        return None
    for key, value in penalty_update.model_dump(exclude_unset=True).items():
        setattr(db_penalty, key, value)
    db.flush()
    create_audit_log(db, "Penalty", penalty_id, ActionType.UPDATE,
                     f"Updated penalty: {penalty_update.model_dump_json()}")
    return db_penalty


def check_overdue_penalties(db: Session) -> int:
    today = date.today()
    overdue_count = 0
    pending_penalties = db.query(Penalty).filter(
        Penalty.status == PenaltyStatus.PENDING,
        Penalty.payment_date.is_(None)
    ).all()

    for penalty in pending_penalties:
        deadline = add_days_without_weekend(penalty.issue_date, 15)
        if today > deadline:
            penalty.status = PenaltyStatus.OVERDUE
            create_audit_log(db, "Penalty", penalty.id, ActionType.UPDATE,
                             f"Marked penalty as overdue")
            overdue_count += 1

    if overdue_count > 0:
        db.flush()
    return overdue_count


def get_audit_logs(db: Session, skip: int = 0, limit: int = 100,
                   entity_type: Optional[str] = None) -> List[AuditLog]:
    query = db.query(AuditLog)
    if entity_type:
        query = query.filter(AuditLog.entity_type == entity_type)
    return query.order_by(AuditLog.created_at.desc()).offset(skip).limit(limit).all()


def get_ship_detail(db: Session, ship_id: int):
    ship = get_ship(db, ship_id)
    if not ship:
        return None

    return {
        "ship": ship,
        "is_high_priority": is_high_priority_ship(db, ship_id),
        "licenses": ship.licenses,
        "inspections": ship.inspections,
        "penalties": ship.penalties,
        "rectifications": ship.rectifications
    }


def associate_records_with_ship(db: Session, ship_id: int,
                                temp_identifier: str) -> int:
    updated = 0
    inspections = db.query(Inspection).filter(
        Inspection.ship_temp_identifier == temp_identifier,
        Inspection.ship_id.is_(None)
    ).all()
    for insp in inspections:
        insp.ship_id = ship_id
        updated += 1

    rectifications = db.query(Rectification).filter(
        Rectification.ship_temp_identifier == temp_identifier,
        Rectification.ship_id.is_(None)
    ).all()
    for rect in rectifications:
        rect.ship_id = ship_id
        updated += 1

    penalties = db.query(Penalty).filter(
        Penalty.ship_temp_identifier == temp_identifier,
        Penalty.ship_id.is_(None)
    ).all()
    for pen in penalties:
        pen.ship_id = ship_id
        updated += 1

    if updated > 0:
        db.flush()
    return updated
