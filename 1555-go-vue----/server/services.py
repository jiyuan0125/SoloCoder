from sqlalchemy.orm import Session
from sqlalchemy import select, and_, or_
from datetime import datetime, date, time, timedelta
from typing import List, Optional, Tuple
from server.models import (
    Pilot, PilotApplication, PilotTask, WeatherStatus, TimeWindowConfig,
    MonthlyStats, PilotLevel, ShipType, TaskStatus, WeatherCondition
)
from server.schemas import (
    PilotCreate, PilotApplicationCreate, DispatchRecommendation,
    PilotWorkload, DailyStats, StatisticsResponse
)

MONTHLY_HOURS_LIMIT = 100.0
APPROACHING_LIMIT_THRESHOLD = 90.0


def get_pilot_level_priority(level: PilotLevel) -> int:
    priorities = {
        PilotLevel.LEVEL1: 3,
        PilotLevel.LEVEL2: 2,
        PilotLevel.LEVEL3: 1
    }
    return priorities.get(level, 0)


def can_pilot_handle_ship(pilot_level: PilotLevel, ship_type: ShipType) -> Tuple[bool, str]:
    if pilot_level == PilotLevel.LEVEL1:
        return True, "一级引航员可引领所有类型船舶"
    
    if ship_type == ShipType.LARGE_OIL_TANKER:
        return False, "超大型油轮仅限一级引航员引领"
    
    if pilot_level == PilotLevel.LEVEL2:
        if ship_type in [ShipType.OIL_TANKER, ShipType.GENERAL_CARGO, ShipType.PASSENGER]:
            return True, "二级引航员可引领普通油轮、货轮和客船"
        return False, "船舶类型不匹配"
    
    if pilot_level == PilotLevel.LEVEL3:
        if ship_type in [ShipType.GENERAL_CARGO, ShipType.PASSENGER]:
            return True, "三级引航员可引领普通货轮和客船"
        return False, "三级引航员仅限普通货轮和客船"
    
    return False, "未知的引航员级别"


def parse_time_str(time_str: str) -> time:
    hours, minutes = map(int, time_str.split(":"))
    return time(hours, minutes)


def is_in_time_window(check_time: datetime, start_str: str, end_str: str) -> bool:
    check_t = check_time.time()
    start_t = parse_time_str(start_str)
    end_t = parse_time_str(end_str)
    
    if start_t <= end_t:
        return start_t <= check_t <= end_t
    else:
        return check_t >= start_t or check_t <= end_t


def check_night_time_rule(db: Session, request_time: datetime, pilot_level: PilotLevel) -> Tuple[bool, str]:
    night_config = db.execute(
        select(TimeWindowConfig).where(
            TimeWindowConfig.rule_type == "night",
            TimeWindowConfig.is_active == True
        )
    ).scalar()
    
    if not night_config:
        return True, ""
    
    if is_in_time_window(request_time, night_config.start_time, night_config.end_time):
        if pilot_level != PilotLevel.LEVEL1:
            return False, f"夜间时段({night_config.start_time}-{night_config.end_time})仅限一级引航员执行"
    
    return True, ""


def check_passenger_peak_rule(db: Session, request_time: datetime, ship_type: ShipType) -> Tuple[bool, str]:
    peak_config = db.execute(
        select(TimeWindowConfig).where(
            TimeWindowConfig.rule_type == "passenger_peak",
            TimeWindowConfig.is_active == True
        )
    ).scalar()
    
    if not peak_config:
        return True, ""
    
    if is_in_time_window(request_time, peak_config.start_time, peak_config.end_time):
        if ship_type not in [ShipType.PASSENGER]:
            return False, f"客运高峰期({peak_config.start_time}-{peak_config.end_time})禁止货轮进出"
    
    return True, ""


def validate_application(db: Session, app_data: PilotApplicationCreate) -> Tuple[bool, str]:
    if app_data.estimated_duration_hours <= 0:
        return False, "预估时长必须大于0"
    
    if app_data.requested_start_time < datetime.utcnow():
        return False, "申请时间不能早于当前时间"
    
    time_ok, time_msg = check_passenger_peak_rule(db, app_data.requested_start_time, app_data.ship_type)
    if not time_ok:
        return False, time_msg
    
    return True, "申请校验通过"


def get_pilot_continuous_work_hours(db: Session, pilot_id: int) -> float:
    now = datetime.utcnow()
    eight_hours_ago = now - timedelta(hours=8)
    
    tasks = db.execute(
        select(PilotTask).where(
            PilotTask.pilot_id == pilot_id,
            or_(
                PilotTask.status.in_([TaskStatus.IN_PROGRESS, TaskStatus.SUSPENDED]),
                and_(
                    PilotTask.status == TaskStatus.COMPLETED,
                    PilotTask.completed_at >= eight_hours_ago
                )
            )
        )
    ).scalars().all()
    
    total_hours = 0.0
    for task in tasks:
        if task.status in [TaskStatus.IN_PROGRESS, TaskStatus.SUSPENDED]:
            start_time = task.started_at or task.assigned_at
            total_hours += (now - start_time).total_seconds() / 3600
        elif task.status == TaskStatus.COMPLETED and task.actual_duration_hours:
            total_hours += task.actual_duration_hours
    
    return total_hours


def get_pilot_available_tasks(db: Session, pilot_id: int) -> List[PilotTask]:
    return db.execute(
        select(PilotTask).where(
            PilotTask.pilot_id == pilot_id,
            PilotTask.status.in_([
                TaskStatus.DISPATCHED,
                TaskStatus.IN_PROGRESS,
                TaskStatus.SUSPENDED
            ])
        )
    ).scalars().all()


def get_recommended_pilots(db: Session, application_id: int) -> List[DispatchRecommendation]:
    application = db.execute(
        select(PilotApplication).where(PilotApplication.id == application_id)
    ).scalar()
    
    if not application:
        return []
    
    now = datetime.utcnow()
    current_month = now.month
    current_year = now.year
    
    all_pilots = db.execute(select(Pilot)).scalars().all()
    
    recommendations = []
    
    for pilot in all_pilots:
        score = 0
        reasons = []
        
        if not pilot.is_on_duty:
            continue
        reasons.append("值班中")
        score += 10
        
        can_handle, handle_msg = can_pilot_handle_ship(pilot.level, application.ship_type)
        if not can_handle:
            continue
        reasons.append(f"级别符合: {handle_msg}")
        score += get_pilot_level_priority(pilot.level) * 10
        
        active_tasks = get_pilot_available_tasks(db, pilot.id)
        if active_tasks:
            continue
        reasons.append("无其他任务")
        score += 20
        
        continuous_hours = get_pilot_continuous_work_hours(db, pilot.id)
        if continuous_hours + application.estimated_duration_hours > 8:
            continue
        reasons.append(f"连续工作时长符合要求(当前{continuous_hours:.1f}h)")
        score += 15
        
        night_ok, night_msg = check_night_time_rule(db, application.requested_start_time, pilot.level)
        if not night_ok:
            continue
        if night_msg:
            reasons.append(night_msg)
        
        if pilot.monthly_work_hours >= APPROACHING_LIMIT_THRESHOLD:
            continue
        reasons.append(f"月度工时未达上限({pilot.monthly_work_hours:.1f}h/{MONTHLY_HOURS_LIMIT}h)")
        score += 10
        
        recommendations.append(DispatchRecommendation(
            pilot_id=pilot.id,
            pilot_name=pilot.name,
            level=pilot.level,
            is_on_duty=pilot.is_on_duty,
            match_score=score,
            reason="; ".join(reasons)
        ))
    
    recommendations.sort(key=lambda x: x.match_score, reverse=True)
    
    return recommendations


def create_pilot_task(db: Session, application_id: int, pilot_id: int) -> Optional[PilotTask]:
    application = db.execute(
        select(PilotApplication).where(PilotApplication.id == application_id)
    ).scalar()
    
    pilot = db.execute(
        select(Pilot).where(Pilot.id == pilot_id)
    ).scalar()
    
    if not application or not pilot:
        return None
    
    existing_task = db.execute(
        select(PilotTask).where(PilotTask.application_id == application_id)
    ).scalar()
    
    if existing_task:
        return existing_task
    
    task = PilotTask(
        application_id=application_id,
        pilot_id=pilot_id,
        status=TaskStatus.DISPATCHED
    )
    
    application.status = "dispatched"
    
    db.add(task)
    db.commit()
    db.refresh(task)
    
    return task


def start_task(db: Session, task_id: int) -> Optional[PilotTask]:
    task = db.execute(
        select(PilotTask).where(PilotTask.id == task_id)
    ).scalar()
    
    if not task:
        return None
    
    weather = db.execute(select(WeatherStatus)).scalar()
    if weather and weather.condition == WeatherCondition.BAD:
        return None
    
    if task.status not in [TaskStatus.DISPATCHED, TaskStatus.SUSPENDED]:
        return None
    
    task.status = TaskStatus.IN_PROGRESS
    if not task.started_at:
        task.started_at = datetime.utcnow()
    
    db.commit()
    db.refresh(task)
    
    return task


def suspend_task(db: Session, task_id: int, reason: str) -> Optional[PilotTask]:
    task = db.execute(
        select(PilotTask).where(PilotTask.id == task_id)
    ).scalar()
    
    if not task or task.status != TaskStatus.IN_PROGRESS:
        return None
    
    task.status = TaskStatus.SUSPENDED
    task.suspended_at = datetime.utcnow()
    task.suspension_reason = reason
    
    if task.started_at:
        elapsed = (datetime.utcnow() - task.started_at).total_seconds() / 3600
        task.progress_percent = min(100, int((elapsed / task.application.estimated_duration_hours) * 100))
    
    db.commit()
    db.refresh(task)
    
    return task


def complete_task(db: Session, task_id: int) -> Optional[PilotTask]:
    task = db.execute(
        select(PilotTask).where(PilotTask.id == task_id)
    ).scalar()
    
    if not task or task.status not in [TaskStatus.IN_PROGRESS, TaskStatus.SUSPENDED]:
        return None
    
    if not task.started_at:
        return None
    
    now = datetime.utcnow()
    actual_duration = (now - task.started_at).total_seconds() / 3600
    
    task.status = TaskStatus.COMPLETED
    task.completed_at = now
    task.actual_duration_hours = actual_duration
    task.progress_percent = 100
    
    if actual_duration >= task.application.estimated_duration_hours * 2:
        task.is_abnormal = True
    
    pilot = db.execute(
        select(Pilot).where(Pilot.id == task.pilot_id)
    ).scalar()
    
    if pilot:
        pilot.total_work_hours += actual_duration
        pilot.monthly_work_hours += actual_duration
        pilot.monthly_task_count += 1
    
    application = db.execute(
        select(PilotApplication).where(PilotApplication.id == task.application_id)
    ).scalar()
    
    if application:
        application.status = "completed"
    
    update_monthly_stats(db, task.pilot_id, actual_duration)
    
    db.commit()
    db.refresh(task)
    
    return task


def update_monthly_stats(db: Session, pilot_id: int, hours: float):
    now = datetime.utcnow()
    
    stats = db.execute(
        select(MonthlyStats).where(
            MonthlyStats.pilot_id == pilot_id,
            MonthlyStats.year == now.year,
            MonthlyStats.month == now.month
        )
    ).scalar()
    
    if stats:
        stats.total_hours += hours
        stats.task_count += 1
    else:
        stats = MonthlyStats(
            pilot_id=pilot_id,
            year=now.year,
            month=now.month,
            total_hours=hours,
            task_count=1
        )
        db.add(stats)


def get_current_weather(db: Session) -> Optional[WeatherStatus]:
    return db.execute(select(WeatherStatus)).scalar()


def update_weather(db: Session, condition: WeatherCondition, description: Optional[str] = None) -> WeatherStatus:
    weather = db.execute(select(WeatherStatus)).scalar()
    
    if not weather:
        weather = WeatherStatus()
        db.add(weather)
    
    weather.condition = condition
    if description:
        weather.description = description
    
    db.commit()
    db.refresh(weather)
    
    return weather


def get_statistics(db: Session) -> StatisticsResponse:
    today = date.today()
    today_start = datetime.combine(today, time.min)
    today_end = datetime.combine(today, time.max)
    
    now = datetime.utcnow()
    current_month = now.month
    current_year = now.year
    
    all_tasks = db.execute(
        select(PilotTask).where(
            PilotTask.assigned_at >= today_start,
            PilotTask.assigned_at <= today_end
        )
    ).scalars().all()
    
    total = len(all_tasks)
    completed = sum(1 for t in all_tasks if t.status == TaskStatus.COMPLETED)
    pending = sum(1 for t in all_tasks if t.status in [TaskStatus.DISPATCHED, TaskStatus.IN_PROGRESS])
    suspended = sum(1 for t in all_tasks if t.status == TaskStatus.SUSPENDED)
    
    daily_stats = DailyStats(
        date=today.isoformat(),
        total_tasks=total,
        completed_tasks=completed,
        pending_tasks=pending,
        suspended_tasks=suspended
    )
    
    all_pilots = db.execute(select(Pilot)).scalars().all()
    pilot_workloads = []
    
    for pilot in all_pilots:
        today_tasks = [t for t in all_tasks if t.pilot_id == pilot.id]
        today_hours = sum(
            t.actual_duration_hours or 0 
            for t in today_tasks 
            if t.status == TaskStatus.COMPLETED
        )
        
        pilot_workloads.append(PilotWorkload(
            pilot_id=pilot.id,
            pilot_name=pilot.name,
            level=pilot.level,
            today_tasks=len(today_tasks),
            today_hours=round(today_hours, 2),
            monthly_hours=round(pilot.monthly_work_hours, 2),
            monthly_tasks=pilot.monthly_task_count,
            approaching_monthly_limit=pilot.monthly_work_hours >= APPROACHING_LIMIT_THRESHOLD
        ))
    
    today_applications = db.execute(
        select(PilotApplication).where(
            PilotApplication.created_at >= today_start,
            PilotApplication.created_at <= today_end,
            PilotApplication.status != "pending"
        )
    ).scalars().all()
    
    waiting_times = []
    for app in today_applications:
        task = app.task
        if task and task.started_at:
            wait_hours = (task.started_at - app.requested_start_time).total_seconds() / 3600
            if wait_hours > 0:
                waiting_times.append(wait_hours)
    
    avg_waiting = sum(waiting_times) / len(waiting_times) if waiting_times else 0.0
    
    return StatisticsResponse(
        daily_stats=daily_stats,
        pilot_workloads=pilot_workloads,
        average_waiting_time_hours=round(avg_waiting, 2)
    )


def create_pilot(db: Session, pilot_data: PilotCreate) -> Pilot:
    pilot = Pilot(
        name=pilot_data.name,
        level=pilot_data.level,
        is_on_duty=pilot_data.is_on_duty
    )
    db.add(pilot)
    db.commit()
    db.refresh(pilot)
    return pilot


def create_application(db: Session, app_data: PilotApplicationCreate) -> PilotApplication:
    application = PilotApplication(
        ship_name=app_data.ship_name,
        ship_type=app_data.ship_type,
        from_location=app_data.from_location,
        to_location=app_data.to_location,
        requested_start_time=app_data.requested_start_time,
        estimated_duration_hours=app_data.estimated_duration_hours,
        remarks=app_data.remarks,
        status="pending"
    )
    db.add(application)
    db.commit()
    db.refresh(application)
    return application


def get_time_window_configs(db: Session) -> List[TimeWindowConfig]:
    return db.execute(select(TimeWindowConfig)).scalars().all()
