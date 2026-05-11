import json
from datetime import datetime, date, timedelta
from typing import Optional, List, Tuple
from sqlalchemy.orm import Session, joinedload
from sqlalchemy import func, and_

from src.core import models, schemas
from src.core.models import (
    User, Area, PatrolTask, PatrolReport, Anomaly,
    PestAnomaly, FireRiskAnomaly, Todo,
    Severity, TodoStatus, UserRole
)


def create_user(db: Session, user_in: schemas.UserCreate) -> User:
    user = User(
        name=user_in.name,
        phone=user_in.phone,
        role=user_in.role
    )
    db.add(user)
    db.commit()
    db.refresh(user)
    return user


def get_users(db: Session, skip: int = 0, limit: int = 100) -> List[User]:
    return db.query(User).offset(skip).limit(limit).all()


def get_user_by_id(db: Session, user_id: int) -> Optional[User]:
    return db.query(User).filter(User.id == user_id).first()


def get_users_by_role(db: Session, role: UserRole) -> List[User]:
    return db.query(User).filter(User.role == role).all()


def create_area(db: Session, area_in: schemas.AreaCreate) -> Area:
    area = Area(
        name=area_in.name,
        area_km2=area_in.area_km2,
        main_tree_species=area_in.main_tree_species,
        ranger_id=area_in.ranger_id,
        manager_id=area_in.manager_id
    )
    db.add(area)
    db.commit()
    db.refresh(area)
    return area


def get_areas(db: Session, skip: int = 0, limit: int = 100) -> List[Area]:
    return db.query(Area).offset(skip).limit(limit).all()


def get_area_by_id(db: Session, area_id: int) -> Optional[Area]:
    return (
        db.query(Area)
        .options(joinedload(Area.ranger), joinedload(Area.manager))
        .filter(Area.id == area_id)
        .first()
    )


def create_patrol_task(db: Session, task_in: schemas.PatrolTaskCreate) -> PatrolTask:
    today = date.today()
    if task_in.patrol_date < today:
        raise ValueError("巡护日期不能是过去的日期")

    waypoints_json = json.dumps([wp.model_dump() for wp in task_in.route_waypoints])

    task = PatrolTask(
        area_id=task_in.area_id,
        ranger_id=task_in.ranger_id,
        patrol_date=task_in.patrol_date,
        route_waypoints=waypoints_json
    )
    db.add(task)
    db.commit()
    db.refresh(task)
    return task


def get_patrol_tasks(db: Session, skip: int = 0, limit: int = 100) -> List[PatrolTask]:
    return (
        db.query(PatrolTask)
        .options(joinedload(PatrolTask.area), joinedload(PatrolTask.ranger))
        .offset(skip)
        .limit(limit)
        .all()
    )


def get_patrol_task_by_id(db: Session, task_id: int) -> Optional[PatrolTask]:
    return (
        db.query(PatrolTask)
        .options(joinedload(PatrolTask.area), joinedload(PatrolTask.ranger))
        .filter(PatrolTask.id == task_id)
        .first()
    )


def parse_waypoints(waypoints_json: str) -> List[schemas.Waypoint]:
    if not waypoints_json:
        return []
    data = json.loads(waypoints_json)
    return [schemas.Waypoint(lat=wp["lat"], lng=wp["lng"]) for wp in data]


def is_location_same(lat1: float, lng1: float, lat2: float, lng2: float, threshold: float = 0.01) -> bool:
    import math
    distance = math.sqrt((lat1 - lat2) ** 2 + (lng1 - lng2) ** 2)
    return distance < threshold


def find_duplicate_anomaly(
    db: Session,
    anomaly_type: str,
    lat: float,
    lng: float,
    within_days: int = 7
) -> Optional[Anomaly]:
    cutoff_time = datetime.utcnow() - timedelta(days=within_days)

    existing = (
        db.query(Anomaly)
        .filter(
            Anomaly.anomaly_type == anomaly_type,
            Anomaly.created_at >= cutoff_time
        )
        .all()
    )

    for anom in existing:
        if is_location_same(lat, lng, anom.location_lat, anom.location_lng):
            return anom

    return None


def create_anomaly(
    db: Session,
    report_id: int,
    anomaly_in: schemas.AnomalyCreate
) -> Anomaly:
    if anomaly_in.pest_data:
        anomaly_type = "pest"
    elif anomaly_in.fire_risk_data:
        anomaly_type = "fire_risk"
    else:
        raise ValueError("必须提供病虫害或火险隐患数据")

    duplicate = find_duplicate_anomaly(
        db,
        anomaly_type,
        anomaly_in.location_lat,
        anomaly_in.location_lng
    )

    anomaly = Anomaly(
        report_id=report_id,
        anomaly_type=anomaly_type,
        location_lat=anomaly_in.location_lat,
        location_lng=anomaly_in.location_lng,
        description=anomaly_in.description,
        is_duplicate=1 if duplicate else 0,
        duplicate_of_id=duplicate.id if duplicate else None
    )
    db.add(anomaly)
    db.flush()

    if anomaly_in.pest_data:
        pest = PestAnomaly(
            anomaly_id=anomaly.id,
            pest_type=anomaly_in.pest_data.pest_type,
            severity=anomaly_in.pest_data.severity,
            trees_affected=anomaly_in.pest_data.trees_affected,
            notes=anomaly_in.pest_data.notes
        )
        db.add(pest)
    elif anomaly_in.fire_risk_data:
        fire = FireRiskAnomaly(
            anomaly_id=anomaly.id,
            risk_type=anomaly_in.fire_risk_data.risk_type,
            status=anomaly_in.fire_risk_data.status,
            notes=anomaly_in.fire_risk_data.notes
        )
        db.add(fire)

    return anomaly


def create_todo_for_anomaly(
    db: Session,
    anomaly: Anomaly,
    area_manager_id: int
) -> Todo:
    priority = 2 if anomaly.anomaly_type == "fire_risk" else 1
    due_date = date.today() + timedelta(days=14)

    todo = Todo(
        anomaly_id=anomaly.id,
        assignee_id=area_manager_id,
        status=TodoStatus.PENDING,
        priority=priority,
        due_date=due_date
    )
    db.add(todo)
    return todo


def create_patrol_report(
    db: Session,
    report_in: schemas.PatrolReportCreate,
    anomalies: List[schemas.AnomalyCreate]
) -> PatrolReport:
    task = get_patrol_task_by_id(db, report_in.task_id)
    if not task:
        raise ValueError("任务不存在")

    if task.report:
        raise ValueError("该任务已有报告")

    required_waypoints = parse_waypoints(task.route_waypoints)
    actual_waypoints = report_in.actual_route

    is_qualified = 1
    if len(actual_waypoints) < len(required_waypoints) / 2:
        is_qualified = 0

    actual_route_json = json.dumps([wp.model_dump() for wp in actual_waypoints])

    report = PatrolReport(
        task_id=task.id,
        area_id=task.area_id,
        ranger_id=task.ranger_id,
        actual_route=actual_route_json,
        is_qualified=is_qualified
    )
    db.add(report)
    db.flush()

    area = get_area_by_id(db, task.area_id)
    manager_id = area.manager_id if area and area.manager_id else None

    created_anomalies = []
    for anom_in in anomalies:
        anom = create_anomaly(db, report.id, anom_in)
        created_anomalies.append(anom)

        if manager_id and not anom.is_duplicate:
            create_todo_for_anomaly(db, anom, manager_id)

    db.commit()
    db.refresh(report)
    return report


def get_patrol_reports(db: Session, skip: int = 0, limit: int = 100) -> List[PatrolReport]:
    return (
        db.query(PatrolReport)
        .options(
            joinedload(PatrolReport.anomalies)
            .options(
                joinedload(Anomaly.pest_data),
                joinedload(Anomaly.fire_risk_data)
            )
        )
        .offset(skip)
        .limit(limit)
        .all()
    )


def get_patrol_report_by_id(db: Session, report_id: int) -> Optional[PatrolReport]:
    return (
        db.query(PatrolReport)
        .options(
            joinedload(PatrolReport.anomalies)
            .options(
                joinedload(Anomaly.pest_data),
                joinedload(Anomaly.fire_risk_data)
            )
        )
        .filter(PatrolReport.id == report_id)
        .first()
    )


def get_todos(db: Session, skip: int = 0, limit: int = 100, overdue_only: bool = False) -> List[Todo]:
    today = date.today()

    query = db.query(Todo).options(
        joinedload(Todo.anomaly).options(
            joinedload(Anomaly.pest_data),
            joinedload(Anomaly.fire_risk_data)
        ),
        joinedload(Todo.assignee)
    )

    if overdue_only:
        query = query.filter(Todo.due_date < today, Todo.status != TodoStatus.DONE)
    else:
        for todo in query.all():
            if todo.status not in [TodoStatus.DONE, TodoStatus.OVERDUE] and todo.due_date < today:
                todo.status = TodoStatus.OVERDUE
        db.commit()

    todos = query.offset(skip).limit(limit).all()
    todos_sorted = sorted(todos, key=lambda t: (-t.priority, t.due_date))
    return todos_sorted


def get_todo_by_id(db: Session, todo_id: int) -> Optional[Todo]:
    return (
        db.query(Todo)
        .options(
            joinedload(Todo.anomaly).options(
                joinedload(Anomaly.pest_data),
                joinedload(Anomaly.fire_risk_data)
            ),
            joinedload(Todo.assignee)
        )
        .filter(Todo.id == todo_id)
        .first()
    )


def update_todo(db: Session, todo_id: int, update_in: schemas.TodoUpdate) -> Optional[Todo]:
    todo = get_todo_by_id(db, todo_id)
    if not todo:
        return None

    if update_in.status is not None:
        todo.status = update_in.status
        if update_in.status == TodoStatus.DONE:
            todo.completed_at = datetime.utcnow()

    if update_in.notes is not None:
        todo.notes = update_in.notes

    db.commit()
    db.refresh(todo)
    return todo


def get_monthly_pest_stats(db: Session) -> List[schemas.MonthlyPestStats]:
    result = (
        db.query(
            func.strftime('%Y', Anomaly.created_at).label('year'),
            func.strftime('%m', Anomaly.created_at).label('month'),
            func.count(Anomaly.id).label('total'),
            func.sum(PestAnomaly.trees_affected).label('trees'),
            func.sum(
                func.case((PestAnomaly.severity == Severity.SEVERE, 1), else_=0)
            ).label('severe')
        )
        .join(PestAnomaly, Anomaly.id == PestAnomaly.anomaly_id)
        .group_by('year', 'month')
        .order_by('year', 'month')
        .all()
    )

    stats = []
    for row in result:
        total = int(row.total)
        severe = int(row.severe or 0)
        ratio = severe / total if total > 0 else 0.0
        stats.append(schemas.MonthlyPestStats(
            year=int(row.year),
            month=int(row.month),
            total_anomalies=total,
            trees_affected=int(row.trees or 0),
            severe_ratio=ratio,
            severe_count=severe
        ))

    return stats


def get_area_health_scores(db: Session) -> List[schemas.AreaHealthScore]:
    areas = get_areas(db)
    scores = []

    for area in areas:
        anomalies = (
            db.query(Anomaly)
            .join(PatrolReport, Anomaly.report_id == PatrolReport.id)
            .filter(PatrolReport.area_id == area.id)
            .options(
                joinedload(Anomaly.pest_data),
                joinedload(Anomaly.fire_risk_data)
            )
            .all()
        )

        total = len(anomalies)
        severe = 0
        fire_count = 0
        pest_count = 0

        for anom in anomalies:
            if anom.anomaly_type == "fire_risk":
                fire_count += 1
            elif anom.anomaly_type == "pest":
                pest_count += 1
                if anom.pest_data and anom.pest_data.severity == Severity.SEVERE:
                    severe += 1

        base_score = 100.0
        base_score -= total * 2
        base_score -= severe * 5
        base_score -= fire_count * 3

        if base_score < 0:
            base_score = 0

        scores.append(schemas.AreaHealthScore(
            area_id=area.id,
            area_name=area.name,
            score=round(base_score, 2),
            total_anomalies=total,
            severe_anomalies=severe,
            fire_risk_count=fire_count,
            pest_count=pest_count
        ))

    return scores
