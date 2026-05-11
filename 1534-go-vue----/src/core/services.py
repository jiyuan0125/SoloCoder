from datetime import datetime, date, timedelta
from typing import List, Optional, Tuple, Dict
from sqlalchemy.orm import Session
from sqlalchemy import func, and_, or_
from .models import (
    EnterpriseORM, FacilityORM, MaintenanceORM, MaintenancePartORM,
    EmissionORM, TodoORM, AgingAlertORM
)
from .config import (
    MAINTENANCE_CYCLES, DAILY_COMPLIANCE_THRESHOLD,
    MONTHLY_METRIC_THRESHOLD, CONSECUTIVE_PART_CHANGE_WARNING
)


def get_maintenance_cycle(facility_type: str) -> timedelta:
    return MAINTENANCE_CYCLES.get(facility_type, MAINTENANCE_CYCLES["other"])


def create_enterprise(db: Session, name: str) -> EnterpriseORM:
    enterprise = EnterpriseORM(name=name)
    db.add(enterprise)
    db.commit()
    db.refresh(enterprise)
    return enterprise


def list_enterprises(db: Session) -> List[EnterpriseORM]:
    return db.query(EnterpriseORM).all()


def get_enterprise(db: Session, enterprise_id: int) -> Optional[EnterpriseORM]:
    return db.query(EnterpriseORM).filter(EnterpriseORM.id == enterprise_id).first()


def create_facility(
    db: Session,
    enterprise_id: int,
    name: str,
    facility_type: str,
    installed_at: date,
    is_running: bool = True
) -> FacilityORM:
    facility = FacilityORM(
        enterprise_id=enterprise_id,
        name=name,
        facility_type=facility_type,
        installed_at=installed_at,
        is_running=is_running
    )
    db.add(facility)
    db.commit()
    db.refresh(facility)
    return facility


def list_facilities(db: Session, enterprise_id: Optional[int] = None) -> List[FacilityORM]:
    query = db.query(FacilityORM)
    if enterprise_id:
        query = query.filter(FacilityORM.enterprise_id == enterprise_id)
    return query.all()


def get_facility(db: Session, facility_id: int) -> Optional[FacilityORM]:
    return db.query(FacilityORM).filter(FacilityORM.id == facility_id).first()


def create_maintenance(
    db: Session,
    facility_id: int,
    maintenance_date: date,
    description: Optional[str],
    parts: List[Dict]
) -> MaintenanceORM:
    total_cost = sum(p["quantity"] * p["unit_cost"] for p in parts)
    maintenance = MaintenanceORM(
        facility_id=facility_id,
        maintenance_date=maintenance_date,
        description=description,
        total_cost=total_cost
    )
    db.add(maintenance)
    db.commit()
    db.refresh(maintenance)
    for p in parts:
        part = MaintenancePartORM(
            maintenance_id=maintenance.id,
            part_name=p["part_name"],
            quantity=p["quantity"],
            unit_cost=p["unit_cost"]
        )
        db.add(part)
    db.commit()
    db.refresh(maintenance)
    check_aging_alerts(db, facility_id)
    complete_maintenance_todo(db, facility_id, maintenance_date)
    return maintenance


def list_maintenances(
    db: Session,
    facility_id: Optional[int] = None,
    start_date: Optional[date] = None,
    end_date: Optional[date] = None
) -> List[MaintenanceORM]:
    query = db.query(MaintenanceORM)
    if facility_id:
        query = query.filter(MaintenanceORM.facility_id == facility_id)
    if start_date:
        query = query.filter(MaintenanceORM.maintenance_date >= start_date)
    if end_date:
        query = query.filter(MaintenanceORM.maintenance_date <= end_date)
    return query.order_by(MaintenanceORM.maintenance_date.desc()).all()


def get_maintenance(db: Session, maintenance_id: int) -> Optional[MaintenanceORM]:
    return db.query(MaintenanceORM).filter(MaintenanceORM.id == maintenance_id).first()


def aggregate_monthly_maintenance_cost(
    db: Session,
    year: int,
    month: int,
    enterprise_id: Optional[int] = None
) -> Dict[int, float]:
    start = date(year, month, 1)
    if month == 12:
        end = date(year + 1, 1, 1)
    else:
        end = date(year, month + 1, 1)
    query = db.query(
        FacilityORM.enterprise_id,
        func.sum(MaintenanceORM.total_cost)
    ).join(
        FacilityORM, MaintenanceORM.facility_id == FacilityORM.id
    ).filter(
        and_(MaintenanceORM.maintenance_date >= start, MaintenanceORM.maintenance_date < end)
    )
    if enterprise_id:
        query = query.filter(FacilityORM.enterprise_id == enterprise_id)
    query = query.group_by(FacilityORM.enterprise_id)
    results = query.all()
    return {row[0]: float(row[1] or 0.0) for row in results}


def create_emission(
    db: Session,
    facility_id: int,
    recorded_at: datetime,
    pollutant: str,
    value: float,
    limit_value: float
) -> EmissionORM:
    emission = EmissionORM(
        facility_id=facility_id,
        recorded_at=recorded_at,
        pollutant=pollutant,
        value=value,
        limit_value=limit_value
    )
    db.add(emission)
    db.commit()
    db.refresh(emission)
    return emission


def list_emissions(
    db: Session,
    facility_id: Optional[int] = None,
    start_time: Optional[datetime] = None,
    end_time: Optional[datetime] = None
) -> List[EmissionORM]:
    query = db.query(EmissionORM)
    if facility_id:
        query = query.filter(EmissionORM.facility_id == facility_id)
    if start_time:
        query = query.filter(EmissionORM.recorded_at >= start_time)
    if end_time:
        query = query.filter(EmissionORM.recorded_at <= end_time)
    return query.order_by(EmissionORM.recorded_at.desc()).all()


def is_emission_compliant(emission: EmissionORM) -> bool:
    return emission.value <= emission.limit_value


def aggregate_hourly_emissions(
    db: Session,
    facility_id: int,
    pollutant: str,
    hour_start: datetime
) -> Dict:
    hour_end = hour_start + timedelta(hours=1)
    emissions = db.query(EmissionORM).filter(
        and_(
            EmissionORM.facility_id == facility_id,
            EmissionORM.pollutant == pollutant,
            EmissionORM.recorded_at >= hour_start,
            EmissionORM.recorded_at < hour_end
        )
    ).all()
    if not emissions:
        return {"count": 0, "avg_value": 0.0, "limit_value": 0.0, "compliant_count": 0}
    avg_value = sum(e.value for e in emissions) / len(emissions)
    limit_value = emissions[0].limit_value
    compliant_count = sum(1 for e in emissions if is_emission_compliant(e))
    return {
        "count": len(emissions),
        "avg_value": avg_value,
        "limit_value": limit_value,
        "compliant_count": compliant_count,
        "is_compliant_avg": avg_value <= limit_value
    }


def calculate_daily_compliance_rate(
    db: Session,
    facility_id: int,
    day: date
) -> float:
    start = datetime.combine(day, datetime.min.time())
    end = start + timedelta(days=1)
    emissions = db.query(EmissionORM).filter(
        and_(
            EmissionORM.facility_id == facility_id,
            EmissionORM.recorded_at >= start,
            EmissionORM.recorded_at < end
        )
    ).all()
    if not emissions:
        return 1.0
    compliant = sum(1 for e in emissions if is_emission_compliant(e))
    return compliant / len(emissions)


def generate_daily_compliance_todos(db: Session, target_day: Optional[date] = None):
    if target_day is None:
        target_day = date.today() - timedelta(days=1)
    facilities = db.query(FacilityORM).all()
    for facility in facilities:
        rate = calculate_daily_compliance_rate(db, facility.id, target_day)
        if rate < DAILY_COMPLIANCE_THRESHOLD:
            existing = db.query(TodoORM).filter(
                and_(
                    TodoORM.facility_id == facility.id,
                    TodoORM.todo_type == "compliance",
                    func.date(TodoORM.created_at) == date.today()
                )
            ).first()
            if not existing:
                todo = TodoORM(
                    facility_id=facility.id,
                    todo_type="compliance",
                    title=f"排放达标率异常 ({target_day})",
                    description=f"设施 {facility.name} 在 {target_day} 的排放达标率为 {rate:.2%}，低于阈值 {DAILY_COMPLIANCE_THRESHOLD:.0%}",
                    due_date=date.today() + timedelta(days=3),
                    status="pending"
                )
                db.add(todo)
    db.commit()


def generate_maintenance_todos(db: Session):
    today = date.today()
    facilities = db.query(FacilityORM).all()
    for facility in facilities:
        cycle = get_maintenance_cycle(facility.facility_type)
        last_maintenance = db.query(MaintenanceORM).filter(
            MaintenanceORM.facility_id == facility.id
        ).order_by(MaintenanceORM.maintenance_date.desc()).first()
        if last_maintenance:
            next_due = last_maintenance.maintenance_date + cycle
        else:
            next_due = facility.installed_at + cycle
        if next_due <= today + timedelta(days=7):
            existing = db.query(TodoORM).filter(
                and_(
                    TodoORM.facility_id == facility.id,
                    TodoORM.todo_type == "maintenance",
                    TodoORM.due_date == next_due,
                    TodoORM.status == "pending"
                )
            ).first()
            if not existing:
                todo = TodoORM(
                    facility_id=facility.id,
                    todo_type="maintenance",
                    title=f"定期维保到期",
                    description=f"设施 {facility.name} ({facility.facility_type}) 维保周期到期",
                    due_date=next_due,
                    status="pending"
                )
                db.add(todo)
    db.commit()


def complete_maintenance_todo(db: Session, facility_id: int, maintenance_date: date):
    todos = db.query(TodoORM).filter(
        and_(
            TodoORM.facility_id == facility_id,
            TodoORM.todo_type == "maintenance",
            TodoORM.status == "pending",
            TodoORM.due_date <= maintenance_date
        )
    ).all()
    for todo in todos:
        todo.status = "completed"
    db.commit()


def list_todos(
    db: Session,
    facility_id: Optional[int] = None,
    status: Optional[str] = None
) -> List[TodoORM]:
    query = db.query(TodoORM)
    if facility_id:
        query = query.filter(TodoORM.facility_id == facility_id)
    if status:
        query = query.filter(TodoORM.status == status)
    return query.order_by(TodoORM.due_date).all()


def update_todo_status(db: Session, todo_id: int, status: str) -> Optional[TodoORM]:
    todo = db.query(TodoORM).filter(TodoORM.id == todo_id).first()
    if todo:
        todo.status = status
        db.commit()
        db.refresh(todo)
    return todo


def check_aging_alerts(db: Session, facility_id: int):
    parts = db.query(MaintenancePartORM.part_name).join(
        MaintenanceORM, MaintenancePartORM.maintenance_id == MaintenanceORM.id
    ).filter(
        MaintenanceORM.facility_id == facility_id
    ).group_by(MaintenancePartORM.part_name).all()
    part_names = [p[0] for p in parts]
    for part_name in part_names:
        recent_maintenances = db.query(MaintenanceORM).join(
            MaintenancePartORM, MaintenanceORM.id == MaintenancePartORM.maintenance_id
        ).filter(
            and_(
                MaintenanceORM.facility_id == facility_id,
                MaintenancePartORM.part_name == part_name
            )
        ).order_by(MaintenanceORM.maintenance_date.desc()).limit(CONSECUTIVE_PART_CHANGE_WARNING).all()
        if len(recent_maintenances) >= CONSECUTIVE_PART_CHANGE_WARNING:
            existing = db.query(AgingAlertORM).filter(
                and_(
                    AgingAlertORM.facility_id == facility_id,
                    AgingAlertORM.part_name == part_name,
                    AgingAlertORM.resolved == False
                )
            ).first()
            if not existing:
                alert = AgingAlertORM(
                    facility_id=facility_id,
                    part_name=part_name,
                    alert_date=date.today(),
                    resolved=False
                )
                db.add(alert)
    db.commit()


def list_aging_alerts(
    db: Session,
    facility_id: Optional[int] = None,
    resolved: Optional[bool] = None
) -> List[AgingAlertORM]:
    query = db.query(AgingAlertORM)
    if facility_id:
        query = query.filter(AgingAlertORM.facility_id == facility_id)
    if resolved is not None:
        query = query.filter(AgingAlertORM.resolved == resolved)
    return query.order_by(AgingAlertORM.alert_date.desc()).all()


def resolve_aging_alert(db: Session, alert_id: int) -> Optional[AgingAlertORM]:
    alert = db.query(AgingAlertORM).filter(AgingAlertORM.id == alert_id).first()
    if alert:
        alert.resolved = True
        db.commit()
        db.refresh(alert)
    return alert


def calculate_monthly_metrics(
    db: Session,
    facility_id: int,
    year: int,
    month: int
) -> Dict:
    start = date(year, month, 1)
    if month == 12:
        end = date(year + 1, 1, 1)
    else:
        end = date(year, month + 1, 1)
    days_in_month = (end - start).days
    facility = get_facility(db, facility_id)
    if not facility:
        return None
    maintenance_count = db.query(func.count(MaintenanceORM.id)).filter(
        and_(
            MaintenanceORM.facility_id == facility_id,
            MaintenanceORM.maintenance_date >= start,
            MaintenanceORM.maintenance_date < end
        )
    ).scalar() or 0
    cycle = get_maintenance_cycle(facility.facility_type)
    expected_maintenances = max(1, int(days_in_month / cycle.days))
    maintenance_completion_rate = min(1.0, maintenance_count / expected_maintenances)
    emissions = db.query(EmissionORM).filter(
        and_(
            EmissionORM.facility_id == facility_id,
            EmissionORM.recorded_at >= datetime.combine(start, datetime.min.time()),
            EmissionORM.recorded_at < datetime.combine(end, datetime.min.time())
        )
    ).all()
    if emissions:
        compliant_emissions = sum(1 for e in emissions if is_emission_compliant(e))
        compliance_rate = compliant_emissions / len(emissions)
    else:
        compliance_rate = 1.0
    operation_rate = 1.0 if facility.is_running else 0.0
    is_focus = any([
        operation_rate < MONTHLY_METRIC_THRESHOLD,
        compliance_rate < MONTHLY_METRIC_THRESHOLD,
        maintenance_completion_rate < MONTHLY_METRIC_THRESHOLD
    ])
    return {
        "year": year,
        "month": month,
        "facility_id": facility_id,
        "facility_name": facility.name,
        "operation_rate": round(operation_rate, 4),
        "compliance_rate": round(compliance_rate, 4),
        "maintenance_completion_rate": round(maintenance_completion_rate, 4),
        "is_focus": is_focus
    }


def get_all_monthly_metrics(db: Session, year: int, month: int) -> List[Dict]:
    facilities = list_facilities(db)
    metrics = []
    for facility in facilities:
        metric = calculate_monthly_metrics(db, facility.id, year, month)
        if metric:
            metrics.append(metric)
    return metrics
