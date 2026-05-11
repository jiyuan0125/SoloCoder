import json
from datetime import datetime, date, timedelta
from typing import List, Optional, Dict
from sqlalchemy.orm import Session
from .models import (
    User, ForestArea, PatrolTask, PatrolReport, Anomaly, Todo,
    MonthlyPestStats, AreaHealthScore, Location, ReportStatus,
    AnomalyType, Role, Severity, TodoStatus, FireRiskStatus
)
from .entities import (
    UserEntity, ForestAreaEntity, PatrolTaskEntity, PatrolReportEntity,
    AnomalyEntity, PestAnomalyEntity, FireRiskAnomalyEntity, TodoEntity
)


def location_to_json(loc: Location) -> str:
    return json.dumps({"latitude": loc.latitude, "longitude": loc.longitude})


def locations_to_json(locs: List[Location]) -> str:
    return json.dumps([{"latitude": loc.latitude, "longitude": loc.longitude} for loc in locs])


def json_to_location(s: str) -> Location:
    data = json.loads(s)
    return Location(latitude=data["latitude"], longitude=data["longitude"])


def json_to_locations(s: str) -> List[Location]:
    data = json.loads(s)
    return [Location(latitude=item["latitude"], longitude=item["longitude"]) for item in data]


def entity_to_user(entity: UserEntity) -> User:
    return User(
        id=entity.id,
        name=entity.name,
        role=entity.role,
        phone=entity.phone
    )


def entity_to_area(entity: ForestAreaEntity) -> ForestArea:
    return ForestArea(
        id=entity.id,
        name=entity.name,
        area_km2=entity.area_km2,
        main_tree_species=entity.main_tree_species,
        ranger_id=entity.ranger_id
    )


def entity_to_task(entity: PatrolTaskEntity) -> PatrolTask:
    return PatrolTask(
        id=entity.id,
        area_id=entity.area_id,
        ranger_id=entity.ranger_id,
        patrol_date=entity.patrol_date,
        route_keypoints=json_to_locations(entity.route_keypoints),
        created_at=entity.created_at
    )


def entity_to_anomaly(entity: AnomalyEntity) -> Anomaly:
    pest = None
    fire_risk = None
    
    if entity.pest:
        pest = {
            "type": entity.pest.type,
            "severity": entity.pest.severity,
            "trees_affected": entity.pest.trees_affected,
            "location": json_to_location(entity.pest.location)
        }
    
    if entity.fire_risk:
        fire_risk = {
            "type": entity.fire_risk.type,
            "status": entity.fire_risk.status,
            "location": json_to_location(entity.fire_risk.location)
        }
    
    from .models import PestAnomaly, FireRiskAnomaly
    
    return Anomaly(
        id=entity.id,
        report_id=entity.report_id,
        anomaly_type=entity.anomaly_type,
        pest=PestAnomaly(**pest) if pest else None,
        fire_risk=FireRiskAnomaly(**fire_risk) if fire_risk else None,
        location=json_to_location(entity.location),
        is_duplicate=entity.is_duplicate,
        created_at=entity.created_at
    )


def entity_to_report(entity: PatrolReportEntity) -> PatrolReport:
    return PatrolReport(
        id=entity.id,
        task_id=entity.task_id,
        submitted_at=entity.submitted_at,
        actual_route=json_to_locations(entity.actual_route),
        anomalies=[entity_to_anomaly(a) for a in entity.anomalies],
        status=entity.status
    )


def entity_to_todo(entity: TodoEntity) -> Todo:
    return Todo(
        id=entity.id,
        anomaly_id=entity.anomaly_id,
        anomaly_type=entity.anomaly_type,
        area_id=entity.area_id,
        assigned_user_id=entity.assigned_user_id,
        status=entity.status,
        created_at=entity.created_at,
        resolved_at=entity.resolved_at
    )


def is_location_near(loc1: Location, loc2: Location, threshold_km: float = 0.5) -> bool:
    from math import radians, sin, cos, sqrt, atan2
    
    R = 6371.0
    
    lat1, lon1 = radians(loc1.latitude), radians(loc1.longitude)
    lat2, lon2 = radians(loc2.latitude), radians(loc2.longitude)
    
    dlon = lon2 - lon1
    dlat = lat2 - lat1
    
    a = sin(dlat / 2)**2 + cos(lat1) * cos(lat2) * sin(dlon / 2)**2
    c = 2 * atan2(sqrt(a), sqrt(1 - a))
    
    distance = R * c
    return distance <= threshold_km


class UserService:
    def __init__(self, db: Session):
        self.db = db
    
    def create_user(self, user: User) -> User:
        entity = UserEntity(
            name=user.name,
            role=user.role,
            phone=user.phone
        )
        self.db.add(entity)
        self.db.commit()
        self.db.refresh(entity)
        return entity_to_user(entity)
    
    def get_user(self, user_id: int) -> Optional[User]:
        entity = self.db.query(UserEntity).filter(UserEntity.id == user_id).first()
        return entity_to_user(entity) if entity else None
    
    def get_users(self) -> List[User]:
        entities = self.db.query(UserEntity).all()
        return [entity_to_user(e) for e in entities]
    
    def get_area_managers(self) -> List[User]:
        entities = self.db.query(UserEntity).filter(UserEntity.role == Role.AREA_MANAGER).all()
        return [entity_to_user(e) for e in entities]


class ForestAreaService:
    def __init__(self, db: Session):
        self.db = db
    
    def create_area(self, area: ForestArea) -> ForestArea:
        entity = ForestAreaEntity(
            name=area.name,
            area_km2=area.area_km2,
            main_tree_species=area.main_tree_species,
            ranger_id=area.ranger_id
        )
        self.db.add(entity)
        self.db.commit()
        self.db.refresh(entity)
        return entity_to_area(entity)
    
    def get_area(self, area_id: int) -> Optional[ForestArea]:
        entity = self.db.query(ForestAreaEntity).filter(ForestAreaEntity.id == area_id).first()
        return entity_to_area(entity) if entity else None
    
    def get_areas(self) -> List[ForestArea]:
        entities = self.db.query(ForestAreaEntity).all()
        return [entity_to_area(e) for e in entities]
    
    def get_areas_by_ranger(self, ranger_id: int) -> List[ForestArea]:
        entities = self.db.query(ForestAreaEntity).filter(ForestAreaEntity.ranger_id == ranger_id).all()
        return [entity_to_area(e) for e in entities]


class PatrolTaskService:
    def __init__(self, db: Session):
        self.db = db
    
    def create_task(self, task: PatrolTask) -> PatrolTask:
        if task.patrol_date < date.today():
            raise ValueError("巡护任务日期不能是过去的日期")
        
        entity = PatrolTaskEntity(
            area_id=task.area_id,
            ranger_id=task.ranger_id,
            patrol_date=task.patrol_date,
            route_keypoints=locations_to_json(task.route_keypoints),
            created_at=datetime.utcnow()
        )
        self.db.add(entity)
        self.db.commit()
        self.db.refresh(entity)
        return entity_to_task(entity)
    
    def get_task(self, task_id: int) -> Optional[PatrolTask]:
        entity = self.db.query(PatrolTaskEntity).filter(PatrolTaskEntity.id == task_id).first()
        return entity_to_task(entity) if entity else None
    
    def get_tasks(self) -> List[PatrolTask]:
        entities = self.db.query(PatrolTaskEntity).all()
        return [entity_to_task(e) for e in entities]
    
    def get_tasks_by_ranger(self, ranger_id: int) -> List[PatrolTask]:
        entities = self.db.query(PatrolTaskEntity).filter(PatrolTaskEntity.ranger_id == ranger_id).all()
        return [entity_to_task(e) for e in entities]
    
    def get_pending_tasks(self) -> List[PatrolTask]:
        reported_task_ids = self.db.query(PatrolReportEntity.task_id).all()
        reported_ids = {t[0] for t in reported_task_ids}
        entities = self.db.query(PatrolTaskEntity).filter(~PatrolTaskEntity.id.in_(reported_ids)).all()
        return [entity_to_task(e) for e in entities]


class PatrolReportService:
    def __init__(self, db: Session):
        self.db = db
        self.anomaly_service = AnomalyService(db)
        self.todo_service = TodoService(db)
    
    def create_report(self, report: PatrolReport) -> PatrolReport:
        task = self.db.query(PatrolTaskEntity).filter(PatrolTaskEntity.id == report.task_id).first()
        if not task:
            raise ValueError("任务不存在")
        
        existing_report = self.db.query(PatrolReportEntity).filter(PatrolReportEntity.task_id == report.task_id).first()
        if existing_report:
            raise ValueError("该任务已提交报告")
        
        required_points = len(json_to_locations(task.route_keypoints))
        actual_points = len(report.actual_route)
        status = ReportStatus.QUALIFIED if actual_points >= required_points / 2 else ReportStatus.UNQUALIFIED
        
        report_entity = PatrolReportEntity(
            task_id=report.task_id,
            submitted_at=datetime.utcnow(),
            actual_route=locations_to_json(report.actual_route),
            status=status
        )
        self.db.add(report_entity)
        self.db.flush()
        
        for anomaly in report.anomalies:
            self._process_anomaly(report_entity.id, anomaly, task.area_id)
        
        self.db.commit()
        self.db.refresh(report_entity)
        return entity_to_report(report_entity)
    
    def _process_anomaly(self, report_id: int, anomaly: Anomaly, area_id: int):
        is_dup = self.anomaly_service.check_duplicate(anomaly.location, anomaly.anomaly_type)
        
        anomaly_entity = AnomalyEntity(
            report_id=report_id,
            anomaly_type=anomaly.anomaly_type,
            location=location_to_json(anomaly.location),
            is_duplicate=is_dup,
            created_at=datetime.utcnow()
        )
        self.db.add(anomaly_entity)
        self.db.flush()
        
        if anomaly.anomaly_type == AnomalyType.PEST and anomaly.pest:
            pest_entity = PestAnomalyEntity(
                anomaly_id=anomaly_entity.id,
                type=anomaly.pest.type,
                severity=anomaly.pest.severity,
                trees_affected=anomaly.pest.trees_affected,
                location=location_to_json(anomaly.pest.location)
            )
            self.db.add(pest_entity)
        
        if anomaly.anomaly_type == AnomalyType.FIRE_RISK and anomaly.fire_risk:
            fire_entity = FireRiskAnomalyEntity(
                anomaly_id=anomaly_entity.id,
                type=anomaly.fire_risk.type,
                status=FireRiskStatus.PENDING,
                location=location_to_json(anomaly.fire_risk.location)
            )
            self.db.add(fire_entity)
        
        self.todo_service.create_todo_for_anomaly(
            anomaly_id=anomaly_entity.id,
            anomaly_type=anomaly.anomaly_type,
            area_id=area_id
        )
    
    def get_report(self, report_id: int) -> Optional[PatrolReport]:
        entity = self.db.query(PatrolReportEntity).filter(PatrolReportEntity.id == report_id).first()
        return entity_to_report(entity) if entity else None
    
    def get_reports(self) -> List[PatrolReport]:
        entities = self.db.query(PatrolReportEntity).all()
        return [entity_to_report(e) for e in entities]
    
    def get_reports_by_task(self, task_id: int) -> Optional[PatrolReport]:
        entity = self.db.query(PatrolReportEntity).filter(PatrolReportEntity.task_id == task_id).first()
        return entity_to_report(entity) if entity else None


class AnomalyService:
    def __init__(self, db: Session):
        self.db = db
    
    def check_duplicate(self, location: Location, anomaly_type: AnomalyType, days: int = 7) -> bool:
        cutoff = datetime.utcnow() - timedelta(days=days)
        entities = self.db.query(AnomalyEntity).filter(
            AnomalyEntity.anomaly_type == anomaly_type,
            AnomalyEntity.created_at >= cutoff
        ).all()
        
        for entity in entities:
            existing_loc = json_to_location(entity.location)
            if is_location_near(location, existing_loc):
                return True
        return False
    
    def get_anomalies(self) -> List[Anomaly]:
        entities = self.db.query(AnomalyEntity).all()
        return [entity_to_anomaly(e) for e in entities]
    
    def get_anomaly(self, anomaly_id: int) -> Optional[Anomaly]:
        entity = self.db.query(AnomalyEntity).filter(AnomalyEntity.id == anomaly_id).first()
        return entity_to_anomaly(entity) if entity else None


class TodoService:
    def __init__(self, db: Session):
        self.db = db
    
    def create_todo_for_anomaly(self, anomaly_id: int, anomaly_type: AnomalyType, area_id: int):
        managers = self.db.query(UserEntity).filter(UserEntity.role == Role.AREA_MANAGER).all()
        if not managers:
            admins = self.db.query(UserEntity).filter(UserEntity.role == Role.ADMIN).all()
            if not admins:
                raise ValueError("没有可用的管理员来分配待办")
            assigned_user = admins[0]
        else:
            assigned_user = managers[0]
        
        priority_fire = anomaly_type == AnomalyType.FIRE_RISK
        
        todo = TodoEntity(
            anomaly_id=anomaly_id,
            anomaly_type=anomaly_type,
            area_id=area_id,
            assigned_user_id=assigned_user.id,
            status=TodoStatus.PENDING,
            created_at=datetime.utcnow()
        )
        self.db.add(todo)
    
    def get_todos(self) -> List[Todo]:
        self._update_overdue_status()
        entities = self.db.query(TodoEntity).order_by(
            TodoEntity.anomaly_type.desc(),
            TodoEntity.created_at.asc()
        ).all()
        return [entity_to_todo(e) for e in entities]
    
    def get_todo(self, todo_id: int) -> Optional[Todo]:
        self._update_overdue_status()
        entity = self.db.query(TodoEntity).filter(TodoEntity.id == todo_id).first()
        return entity_to_todo(entity) if entity else None
    
    def _update_overdue_status(self):
        cutoff = datetime.utcnow() - timedelta(days=14)
        self.db.query(TodoEntity).filter(
            TodoEntity.status.in_([TodoStatus.PENDING, TodoStatus.IN_PROGRESS]),
            TodoEntity.created_at < cutoff
        ).update({"status": TodoStatus.OVERDUE})
        self.db.commit()
    
    def resolve_todo(self, todo_id: int) -> Optional[Todo]:
        self._update_overdue_status()
        entity = self.db.query(TodoEntity).filter(TodoEntity.id == todo_id).first()
        if not entity:
            return None
        
        entity.status = TodoStatus.RESOLVED
        entity.resolved_at = datetime.utcnow()
        self.db.commit()
        self.db.refresh(entity)
        return entity_to_todo(entity)
    
    def get_todos_by_user(self, user_id: int) -> List[Todo]:
        self._update_overdue_status()
        entities = self.db.query(TodoEntity).filter(
            TodoEntity.assigned_user_id == user_id
        ).order_by(
            TodoEntity.anomaly_type.desc(),
            TodoEntity.created_at.asc()
        ).all()
        return [entity_to_todo(e) for e in entities]


class StatsService:
    def __init__(self, db: Session):
        self.db = db
    
    def get_monthly_pest_stats(self, year: int, month: int) -> MonthlyPestStats:
        start_date = datetime(year, month, 1)
        if month == 12:
            end_date = datetime(year + 1, 1, 1)
        else:
            end_date = datetime(year, month + 1, 1)
        
        anomalies = self.db.query(AnomalyEntity).filter(
            AnomalyEntity.anomaly_type == AnomalyType.PEST,
            AnomalyEntity.created_at >= start_date,
            AnomalyEntity.created_at < end_date
        ).all()
        
        total_count = len(anomalies)
        trees_affected = 0
        severe_count = 0
        
        for anomaly in anomalies:
            if anomaly.pest:
                trees_affected += anomaly.pest.trees_affected
                if anomaly.pest.severity == Severity.SEVERE:
                    severe_count += 1
        
        severe_percentage = (severe_count / total_count * 100) if total_count > 0 else 0.0
        
        return MonthlyPestStats(
            year=year,
            month=month,
            new_discovery_count=total_count,
            trees_affected_total=trees_affected,
            severe_percentage=round(severe_percentage, 2)
        )
    
    def get_area_health_scores(self) -> List[AreaHealthScore]:
        areas = self.db.query(ForestAreaEntity).all()
        scores = []
        
        for area in areas:
            total_anomalies = 0
            severe_anomalies = 0
            
            tasks = self.db.query(PatrolTaskEntity).filter(PatrolTaskEntity.area_id == area.id).all()
            task_ids = [t.id for t in tasks]
            
            reports = self.db.query(PatrolReportEntity).filter(PatrolReportEntity.task_id.in_(task_ids)).all()
            report_ids = [r.id for r in reports]
            
            anomalies = self.db.query(AnomalyEntity).filter(AnomalyEntity.report_id.in_(report_ids)).all()
            
            for anomaly in anomalies:
                total_anomalies += 1
                if anomaly.pest and anomaly.pest.severity == Severity.SEVERE:
                    severe_anomalies += 1
                if anomaly.fire_risk and anomaly.fire_risk.status != FireRiskStatus.RESOLVED:
                    severe_anomalies += 1
            
            base_score = 100.0
            penalty = total_anomalies * 5 + severe_anomalies * 15
            score = max(0.0, base_score - penalty)
            
            scores.append(AreaHealthScore(
                area_id=area.id,
                area_name=area.name,
                score=round(score, 2),
                anomaly_count=total_anomalies,
                severe_anomaly_count=severe_anomalies
            ))
        
        return scores
