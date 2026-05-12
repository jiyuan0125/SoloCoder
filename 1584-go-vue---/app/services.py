from datetime import datetime, timedelta, date, time
from typing import List, Optional, Tuple
from sqlalchemy.orm import Session
from app.models import (
    Task, TaskStaff, MaintenanceWindow, Staff, Qualification,
    Line, TaskStatus, Notification
)
from app.schemas import ConflictDetail


def get_maintenance_window(
    db: Session,
    line_id: int,
    target_date: date
) -> Optional[MaintenanceWindow]:
    window = db.query(MaintenanceWindow).filter(
        MaintenanceWindow.line_id == line_id,
        MaintenanceWindow.date == target_date
    ).first()
    
    if window:
        return window
    
    return db.query(MaintenanceWindow).filter(
        MaintenanceWindow.line_id == line_id,
        MaintenanceWindow.is_general == True
    ).first()


def is_time_in_window(
    target_dt: datetime,
    start_time: time,
    end_time: time,
    target_date: date
) -> bool:
    target_time = target_dt.time()
    
    if start_time <= end_time:
        return start_time <= target_time <= end_time
    else:
        return target_time >= start_time or target_time <= end_time


def check_task_in_maintenance_window(
    db: Session,
    task: Task
) -> Tuple[bool, str]:
    line_id = task.line_id
    start_dt = task.scheduled_start
    end_dt = task.scheduled_end
    
    current_date = start_dt.date()
    end_date = end_dt.date()
    
    while current_date <= end_date:
        window = get_maintenance_window(db, line_id, current_date)
        
        if not window:
            return False, f"未找到线路 {line_id} 在日期 {current_date} 的天窗时间配置"
        
        if current_date == start_dt.date():
            if not is_time_in_window(start_dt, window.start_time, window.end_time, current_date):
                return False, f"作业开始时间不在 {current_date} 的天窗时间范围内"
        
        if current_date == end_dt.date():
            if not is_time_in_window(end_dt, window.start_time, window.end_time, current_date):
                return False, f"作业结束时间不在 {current_date} 的天窗时间范围内"
        
        current_date += timedelta(days=1)
    
    return True, ""


def is_interval_overlap(
    start1: int, end1: int,
    start2: int, end2: int
) -> bool:
    return not (end1 < start2 or end2 < start1)


def is_time_overlap_with_buffer(
    start1: datetime, end1: datetime,
    start2: datetime, end2: datetime,
    buffer_minutes: int = 60
) -> bool:
    buffered_end1 = end1 + timedelta(minutes=buffer_minutes)
    buffered_end2 = end2 + timedelta(minutes=buffer_minutes)
    
    return start1 < buffered_end2 and start2 < buffered_end1


def check_safety_conflicts(
    db: Session,
    task: Task,
    exclude_task_id: Optional[int] = None
) -> List[ConflictDetail]:
    conflicts = []
    
    active_statuses = [TaskStatus.APPROVED, TaskStatus.IN_PROGRESS]
    
    query = db.query(Task).filter(
        Task.line_id == task.line_id,
        Task.status.in_(active_statuses)
    )
    
    if exclude_task_id:
        query = query.filter(Task.id != exclude_task_id)
    
    other_tasks = query.all()
    
    for other_task in other_tasks:
        if is_interval_overlap(
            task.start_kp, task.end_kp,
            other_task.start_kp, other_task.end_kp
        ):
            if is_time_overlap_with_buffer(
                task.scheduled_start, task.scheduled_end,
                other_task.scheduled_start, other_task.scheduled_end,
                buffer_minutes=0
            ):
                conflicts.append(ConflictDetail(
                    conflict_type="safety",
                    task_id=other_task.id,
                    task_title=other_task.title,
                    reason=f"同线路同区间（KP {other_task.start_kp}-{other_task.end_kp}）同时段作业冲突"
                ))
    
    return conflicts


def check_staff_conflicts(
    db: Session,
    task: Task,
    assigned_staff_ids: List[int],
    exclude_task_id: Optional[int] = None
) -> List[ConflictDetail]:
    conflicts = []
    
    if not assigned_staff_ids:
        return conflicts
    
    active_statuses = [TaskStatus.APPROVED, TaskStatus.IN_PROGRESS]
    
    task_staff_ids = set(assigned_staff_ids)
    
    query = db.query(Task).filter(
        Task.status.in_(active_statuses)
    )
    
    if exclude_task_id:
        query = query.filter(Task.id != exclude_task_id)
    
    other_tasks = query.all()
    
    for other_task in other_tasks:
        other_staff_ids = set(
            ts.staff_id for ts in 
            db.query(TaskStaff).filter(TaskStaff.task_id == other_task.id).all()
        )
        
        common_staff = task_staff_ids & other_staff_ids
        
        if common_staff:
            if is_time_overlap_with_buffer(
                task.scheduled_start, task.scheduled_end,
                other_task.scheduled_start, other_task.scheduled_end,
                buffer_minutes=60
            ):
                staff_names = [
                    db.query(Staff).filter(Staff.id == sid).first().name
                    for sid in common_staff
                ]
                
                conflicts.append(ConflictDetail(
                    conflict_type="staff",
                    task_id=other_task.id,
                    task_title=other_task.title,
                    reason=f"人员冲突：{', '.join(staff_names)} 在 {other_task.scheduled_start} - {other_task.scheduled_end} 已有安排（含1小时缓冲）"
                ))
    
    return conflicts


def check_all_conflicts(
    db: Session,
    task: Task,
    assigned_staff_ids: List[int],
    exclude_task_id: Optional[int] = None
) -> List[ConflictDetail]:
    safety_conflicts = check_safety_conflicts(db, task, exclude_task_id)
    staff_conflicts = check_staff_conflicts(db, task, assigned_staff_ids, exclude_task_id)
    
    return safety_conflicts + staff_conflicts


def check_qualifications(
    db: Session,
    staff_ids: List[int],
    check_date: date
) -> Tuple[bool, List[str], bool]:
    all_valid = True
    messages = []
    has_new_staff = False
    
    for staff_id in staff_ids:
        staff = db.query(Staff).filter(Staff.id == staff_id).first()
        
        if not staff:
            all_valid = False
            messages.append(f"人员 {staff_id} 不存在")
            continue
        
        quals = db.query(Qualification).filter(
            Qualification.staff_id == staff_id
        ).all()
        
        if not quals:
            if staff.join_date and (check_date - staff.join_date).days < 30:
                has_new_staff = True
                messages.append(f"人员 {staff.name}（工号：{staff.employee_id}）为新入职员工，暂不做资质校验")
            else:
                all_valid = False
                messages.append(f"人员 {staff.name}（工号：{staff.employee_id}）无安全资质记录")
        else:
            has_valid = False
            for qual in quals:
                if qual.valid_from <= check_date:
                    if not qual.valid_to or qual.valid_to >= check_date:
                        has_valid = True
                        break
            
            if not has_valid:
                all_valid = False
                messages.append(f"人员 {staff.name}（工号：{staff.employee_id}）的安全资质已过期或无效")
    
    return all_valid, messages, has_new_staff


def create_notification(
    db: Session,
    task_id: int,
    staff_id: int,
    notification_type: str,
    message: str
) -> Notification:
    notification = Notification(
        task_id=task_id,
        staff_id=staff_id,
        type=notification_type,
        message=message
    )
    db.add(notification)
    db.commit()
    db.refresh(notification)
    return notification


def check_and_send_reminders(db: Session):
    now = datetime.utcnow()
    reminder_time = now + timedelta(minutes=30)
    
    tasks = db.query(Task).filter(
        Task.status == TaskStatus.APPROVED,
        Task.scheduled_start >= now,
        Task.scheduled_start <= reminder_time
    ).all()
    
    for task in tasks:
        existing = db.query(Notification).filter(
            Notification.task_id == task.id,
            Notification.type == "start_reminder"
        ).first()
        
        if not existing:
            create_notification(
                db,
                task.id,
                task.responsible_person_id,
                "start_reminder",
                f"作业「{task.title}」将在30分钟后开始，请做好准备。"
            )


def check_and_send_timeout_reminders(db: Session):
    now = datetime.utcnow()
    
    tasks = db.query(Task).filter(
        Task.status == TaskStatus.IN_PROGRESS,
        Task.scheduled_end < now
    ).all()
    
    for task in tasks:
        time_since_end = now - task.scheduled_end
        intervals_passed = int(time_since_end.total_seconds() // (30 * 60))
        
        latest_notification = db.query(Notification).filter(
            Notification.task_id == task.id,
            Notification.type == "timeout_reminder"
        ).order_by(Notification.sent_at.desc()).first()
        
        should_send = False
        if not latest_notification:
            should_send = True
        else:
            time_since_last = now - latest_notification.sent_at
            if time_since_last >= timedelta(minutes=30):
                should_send = True
        
        if should_send:
            create_notification(
                db,
                task.id,
                task.responsible_person_id,
                "timeout_reminder",
                f"作业「{task.title}」已超时 {intervals_passed * 30} 分钟，请尽快完成。"
            )
