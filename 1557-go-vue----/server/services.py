from datetime import date, datetime, timedelta
from dateutil.relativedelta import relativedelta
from typing import Optional, List, Tuple
from sqlalchemy.orm import Session
from sqlalchemy import and_, or_

from . import models, schemas
from .config import (
    INSPECTION_TYPES,
    CERTIFICATE_REMINDER_DAYS,
    INSPECTION_ARRANGE_BEFORE_DAYS,
    SPECIAL_INSPECTION_INTERVAL_YEARS,
    INTERMEDIATE_INSPECTION_INTERVAL_MONTHS,
)


def get_next_inspection_info(
    db: Session, vessel_id: int
) -> Tuple[Optional[str], Optional[date]]:
    inspections = (
        db.query(models.Inspection)
        .filter(models.Inspection.vessel_id == vessel_id)
        .order_by(models.Inspection.inspection_date.desc())
        .all()
    )

    if not inspections:
        return "initial", None

    last_inspection = inspections[0]

    if last_inspection.inspection_type == "temporary":
        regular_inspections = [
            i for i in inspections if i.inspection_type != "temporary"
        ]
        if not regular_inspections:
            return "initial", None
        last_inspection = regular_inspections[0]

    if last_inspection.inspection_type == "initial":
        next_type = "annual"
        next_due = last_inspection.inspection_date + relativedelta(years=1)
    elif last_inspection.inspection_type == "annual":
        special_inspections = [
            i for i in inspections if i.inspection_type == "special"
        ]
        last_special_date = (
            special_inspections[0].inspection_date if special_inspections else None
        )

        if last_special_date:
            intermediate_due = last_special_date + relativedelta(
                months=INTERMEDIATE_INSPECTION_INTERVAL_MONTHS
            )
            if (
                last_inspection.inspection_date
                < intermediate_due
                <= last_inspection.inspection_date + relativedelta(years=1)
            ):
                next_type = "intermediate"
                next_due = intermediate_due
            else:
                annual_count = 0
                for i in inspections:
                    if i.inspection_type == "annual":
                        if i.inspection_date > last_special_date:
                            annual_count += 1
                        else:
                            break
                if annual_count >= 3:
                    next_type = "special"
                    next_due = last_special_date + relativedelta(
                        years=SPECIAL_INSPECTION_INTERVAL_YEARS
                    )
                else:
                    next_type = "annual"
                    next_due = last_inspection.inspection_date + relativedelta(years=1)
        else:
            next_type = "annual"
            next_due = last_inspection.inspection_date + relativedelta(years=1)
    elif last_inspection.inspection_type == "intermediate":
        special_inspections = [
            i for i in inspections if i.inspection_type == "special"
        ]
        last_special_date = (
            special_inspections[0].inspection_date if special_inspections else None
        )

        if last_special_date:
            special_due = last_special_date + relativedelta(
                years=SPECIAL_INSPECTION_INTERVAL_YEARS
            )
            time_to_special = (special_due - last_inspection.inspection_date).days
            if time_to_special > 365:
                next_type = "annual"
                next_due = last_inspection.inspection_date + relativedelta(years=1)
            else:
                next_type = "special"
                next_due = special_due
        else:
            next_type = "annual"
            next_due = last_inspection.inspection_date + relativedelta(years=1)
    elif last_inspection.inspection_type == "special":
        next_type = "annual"
        next_due = last_inspection.inspection_date + relativedelta(years=1)
    else:
        next_type = "annual"
        next_due = last_inspection.inspection_date + relativedelta(years=1)

    return next_type, next_due


def create_inspection_todo(
    db: Session, vessel_id: int, inspection_type: str, due_date: date
) -> models.TodoItem:
    from .config import INSPECTION_TYPES

    arrange_deadline = due_date - timedelta(days=INSPECTION_ARRANGE_BEFORE_DAYS)

    existing = (
        db.query(models.TodoItem)
        .filter(
            models.TodoItem.vessel_id == vessel_id,
            models.TodoItem.todo_type == "inspection",
            models.TodoItem.inspection_type == inspection_type,
            models.TodoItem.due_date == due_date,
            models.TodoItem.status.in_(["pending", "in_progress"]),
        )
        .first()
    )

    if existing:
        return existing

    todo = models.TodoItem(
        vessel_id=vessel_id,
        todo_type="inspection",
        title=f"{INSPECTION_TYPES.get(inspection_type, inspection_type)}到期提醒",
        inspection_type=inspection_type,
        due_date=due_date,
        arrange_deadline=arrange_deadline,
        status="pending",
    )
    db.add(todo)
    db.commit()
    db.refresh(todo)
    return todo


def generate_inspection_todos(db: Session, vessel_id: int) -> List[models.TodoItem]:
    todos = []
    next_type, next_due = get_next_inspection_info(db, vessel_id)

    if next_due and next_type:
        todo = create_inspection_todo(db, vessel_id, next_type, next_due)
        todos.append(todo)

    return todos


def check_dock_availability(
    db: Session, dock_id: int, start_date: date, end_date: date
) -> Tuple[bool, str]:
    dock = db.query(models.Dock).filter(models.Dock.id == dock_id).first()
    if not dock:
        return False, "船坞不存在"

    existing_dockings = (
        db.query(models.Docking)
        .filter(
            models.Docking.dock_id == dock_id,
            models.Docking.status.in_(["scheduled", "in_progress"]),
            or_(
                and_(
                    models.Docking.start_date <= end_date,
                    models.Docking.end_date >= start_date,
                ),
            ),
        )
        .all()
    )

    if existing_dockings:
        return False, f"船坞在 {start_date} 至 {end_date} 期间已被占用"

    maintenances = (
        db.query(models.DockMaintenance)
        .filter(
            models.DockMaintenance.dock_id == dock_id,
            and_(
                models.DockMaintenance.start_date <= end_date,
                models.DockMaintenance.end_date >= start_date,
            ),
        )
        .all()
    )

    if maintenances:
        return False, f"船坞在 {start_date} 至 {end_date} 期间处于维护期"

    return True, "船坞可用"


def check_vessel_operation_records(
    db: Session, vessel_id: int, check_date: date
) -> bool:
    records = (
        db.query(models.OperationRecord)
        .filter(
            models.OperationRecord.vessel_id == vessel_id,
            models.OperationRecord.start_date <= check_date,
            or_(
                models.OperationRecord.end_date == None,
                models.OperationRecord.end_date >= check_date,
            ),
        )
        .all()
    )
    return len(records) > 0


def mark_expired_todos_for_operation_period(db: Session, vessel_id: int):
    active_operations = (
        db.query(models.OperationRecord)
        .filter(
            models.OperationRecord.vessel_id == vessel_id,
            models.OperationRecord.status == "active",
        )
        .all()
    )

    today = date.today()

    for operation in active_operations:
        op_start = operation.start_date
        op_end = operation.end_date or today

        expired_todos = (
            db.query(models.TodoItem)
            .filter(
                models.TodoItem.vessel_id == vessel_id,
                models.TodoItem.status == "pending",
                models.TodoItem.todo_type == "inspection",
                models.TodoItem.due_date < op_start,
            )
            .all()
        )

        for todo in expired_todos:
            todo.status = "expired"
            todo.remarks = (
                f"证书过期期间船舶有营运记录: {op_start} - {op_end}"
            )

    db.commit()


def check_illegal_operation(db: Session, vessel_id: int) -> bool:
    today = date.today()

    active_operations = (
        db.query(models.OperationRecord)
        .filter(
            models.OperationRecord.vessel_id == vessel_id,
            models.OperationRecord.status == "active",
        )
        .all()
    )

    if not active_operations:
        return False

    valid_certificates = (
        db.query(models.Certificate)
        .filter(
            models.Certificate.vessel_id == vessel_id,
            models.Certificate.status == "valid",
            models.Certificate.expiry_date >= today,
        )
        .all()
    )

    if not valid_certificates:
        for op in active_operations:
            op.is_illegal = True
        db.commit()
        return True

    return False


def create_certificate_reminders(db: Session):
    today = date.today()

    certificates = (
        db.query(models.Certificate)
        .filter(
            models.Certificate.status == "valid",
            models.Certificate.expiry_date >= today,
        )
        .all()
    )

    for cert in certificates:
        days_until_expiry = (cert.expiry_date - today).days

        for days in CERTIFICATE_REMINDER_DAYS:
            if days_until_expiry <= days:
                existing = (
                    db.query(models.Reminder)
                    .filter(
                        models.Reminder.certificate_id == cert.id,
                        models.Reminder.days_before_expiry == days,
                    )
                    .first()
                )

                if not existing:
                    reminder = models.Reminder(
                        vessel_id=cert.vessel_id,
                        certificate_id=cert.id,
                        reminder_type="certificate_expiry",
                        days_before_expiry=days,
                        message=f"证书 '{cert.certificate_type}' 将在 {days} 天后过期",
                        sent_date=today,
                    )
                    db.add(reminder)

    db.commit()


def complete_inspection(
    db: Session, inspection_id: int, result: str = "passed"
) -> models.Inspection:
    inspection = (
        db.query(models.Inspection).filter(models.Inspection.id == inspection_id).first()
    )

    if not inspection:
        raise ValueError("检验记录不存在")

    inspection.result = result
    db.commit()

    pending_todos = (
        db.query(models.TodoItem)
        .filter(
            models.TodoItem.vessel_id == inspection.vessel_id,
            models.TodoItem.todo_type == "inspection",
            models.TodoItem.inspection_type == inspection.inspection_type,
            models.TodoItem.status.in_(["pending", "in_progress"]),
        )
        .all()
    )

    for todo in pending_todos:
        todo.status = "completed"

    mark_expired_todos_for_operation_period(db, inspection.vessel_id)

    generate_inspection_todos(db, inspection.vessel_id)

    db.commit()
    db.refresh(inspection)
    return inspection


def update_certificate_status(db: Session):
    today = date.today()

    expired_certs = (
        db.query(models.Certificate)
        .filter(
            models.Certificate.status == "valid",
            models.Certificate.expiry_date < today,
        )
        .all()
    )

    for cert in expired_certs:
        cert.status = "expired"
        check_illegal_operation(db, cert.vessel_id)

    db.commit()


def get_vessel_inspections(db: Session, vessel_id: int):
    return (
        db.query(models.Inspection)
        .filter(models.Inspection.vessel_id == vessel_id)
        .order_by(models.Inspection.inspection_date.desc())
        .all()
    )


def get_vessel_certificates(db: Session, vessel_id: int):
    return (
        db.query(models.Certificate)
        .filter(models.Certificate.vessel_id == vessel_id)
        .order_by(models.Certificate.issue_date.desc())
        .all()
    )


def get_vessel_dockings(db: Session, vessel_id: int):
    return (
        db.query(models.Docking)
        .filter(models.Docking.vessel_id == vessel_id)
        .order_by(models.Docking.start_date.desc())
        .all()
    )


def get_vessel_todos(db: Session, vessel_id: int, status: Optional[str] = None):
    query = db.query(models.TodoItem).filter(models.TodoItem.vessel_id == vessel_id)

    if status:
        query = query.filter(models.TodoItem.status == status)

    return query.order_by(models.TodoItem.due_date.asc()).all()
