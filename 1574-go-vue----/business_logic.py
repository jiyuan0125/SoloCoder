from datetime import datetime, timedelta
from typing import Optional
from sqlalchemy.orm import Session
from models import (
    Defect, Task, MaintenanceTeam, Rail, RailMaintenance, RailReplacement,
    Warning, Inspection, TeamStatus, TaskStatus, DefectLevel
)


LEVEL_PRIORITY = {
    DefectLevel.LEVEL_3.value: 3,
    DefectLevel.LEVEL_2.value: 2,
    DefectLevel.LEVEL_1.value: 1,
}


def calculate_due_date(defect_level: str, now: Optional[datetime] = None) -> datetime:
    if now is None:
        now = datetime.utcnow()
    
    if defect_level == DefectLevel.LEVEL_3.value:
        return now + timedelta(hours=24)
    elif defect_level == DefectLevel.LEVEL_2.value:
        return now + timedelta(days=7)
    elif defect_level == DefectLevel.LEVEL_1.value:
        return now + timedelta(days=30)
    return now


def get_priority(defect_level: str) -> int:
    return LEVEL_PRIORITY.get(defect_level, 0)


def find_idle_team(db: Session) -> Optional[MaintenanceTeam]:
    return db.query(MaintenanceTeam).filter(
        MaintenanceTeam.status == TeamStatus.IDLE.value
    ).first()


def assign_task_to_team(db: Session, task: Task, team: MaintenanceTeam):
    task.team_id = team.id
    team.status = TeamStatus.BUSY.value
    team.current_task_id = task.id
    db.add(task)
    db.add(team)
    db.flush()


def create_task_for_defect(db: Session, defect: Defect, task_type: str = "缺陷处理") -> Task:
    priority = get_priority(defect.level)
    due_date = calculate_due_date(defect.level)
    
    task = Task(
        defect_id=defect.id,
        task_type=task_type,
        status=TaskStatus.PENDING.value,
        priority=priority,
        due_date=due_date,
    )
    db.add(task)
    db.flush()
    
    idle_team = find_idle_team(db)
    if idle_team:
        assign_task_to_team(db, task, idle_team)
    
    return task


def update_tasks_on_defect_level_change(db: Session, defect: Defect, old_level: str, new_level: str):
    if old_level == new_level:
        return
    
    new_priority = get_priority(new_level)
    new_due_date = calculate_due_date(new_level)
    
    tasks = db.query(Task).filter(
        Task.defect_id == defect.id,
        Task.status.in_([TaskStatus.PENDING.value, TaskStatus.IN_PROGRESS.value])
    ).all()
    
    for task in tasks:
        task.priority = new_priority
        task.due_date = new_due_date
        db.add(task)


def check_rail_wear(db: Session, rail: Rail) -> list:
    warnings = []
    
    if rail.wear_mm > 2:
        existing_warning = db.query(Warning).filter(
            Warning.target_type == "Rail",
            Warning.target_id == rail.id,
            Warning.warning_type == "需要修理性打磨",
            Warning.is_acknowledged == False
        ).first()
        
        if not existing_warning:
            warning = Warning(
                warning_type="需要修理性打磨",
                target_type="Rail",
                target_id=rail.id,
                message=f"钢轨 {rail.rail_number} 磨耗量 {rail.wear_mm}mm 超过2mm，需要进行修理性打磨"
            )
            db.add(warning)
            warnings.append(warning)
    
    if rail.wear_mm > 1:
        existing = db.query(RailMaintenance).filter(
            RailMaintenance.rail_id == rail.id,
            RailMaintenance.maintenance_type == "修理性打磨"
        ).order_by(RailMaintenance.maintenance_date.desc()).first()
        
        if existing and existing.after_wear_mm is not None and existing.after_wear_mm > 1:
            existing_warning = db.query(Warning).filter(
                Warning.target_type == "Rail",
                Warning.target_id == rail.id,
                Warning.warning_type == "需要换轨",
                Warning.is_acknowledged == False
            ).first()
            
            if not existing_warning:
                warning = Warning(
                    warning_type="需要换轨",
                    target_type="Rail",
                    target_id=rail.id,
                    message=f"钢轨 {rail.rail_number} 修理性打磨后磨耗量仍超过1mm，需要换轨"
                )
                db.add(warning)
                warnings.append(warning)
    
    return warnings


def check_rail_life(db: Session, rail: Rail) -> Optional[Warning]:
    if rail.max_total_weight <= 0:
        return None
    
    remaining_percent = ((rail.max_total_weight - rail.current_total_weight) / rail.max_total_weight) * 100
    
    if remaining_percent <= 10:
        existing_warning = db.query(Warning).filter(
            Warning.target_type == "Rail",
            Warning.target_id == rail.id,
            Warning.warning_type == "剩余寿命不足",
            Warning.is_acknowledged == False
        ).first()
        
        if not existing_warning:
            warning = Warning(
                warning_type="剩余寿命不足",
                target_type="Rail",
                target_id=rail.id,
                message=f"钢轨 {rail.rail_number} 剩余寿命 {remaining_percent:.1f}%，低于10%，建议更换"
            )
            db.add(warning)
            return warning
    
    return None


def get_ordered_tasks(db: Session) -> list:
    return db.query(Task).filter(
        Task.status.in_([TaskStatus.PENDING.value, TaskStatus.IN_PROGRESS.value])
    ).order_by(Task.priority.desc(), Task.due_date.asc()).all()


def complete_task(db: Session, task: Task):
    task.status = TaskStatus.COMPLETED.value
    task.completed_at = datetime.utcnow()
    
    if task.team_id:
        team = db.query(MaintenanceTeam).filter(MaintenanceTeam.id == task.team_id).first()
        if team:
            team.status = TeamStatus.IDLE.value
            team.current_task_id = None
            db.add(team)
    
    db.add(task)
    db.flush()
    
    defect = db.query(Defect).filter(Defect.id == task.defect_id).first()
    if defect:
        all_tasks = db.query(Task).filter(
            Task.defect_id == defect.id,
            Task.status.in_([TaskStatus.PENDING.value, TaskStatus.IN_PROGRESS.value])
        ).count()
        
        if all_tasks == 0:
            defect.is_resolved = True
            defect.resolved_at = datetime.utcnow()
            db.add(defect)


def assign_inspection_to_team(db: Session, inspection: Inspection) -> bool:
    idle_team = find_idle_team(db)
    if idle_team:
        inspection.team_id = idle_team.id
        db.add(inspection)
        db.flush()
        return True
    return False
