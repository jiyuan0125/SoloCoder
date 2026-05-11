from apscheduler.schedulers.background import BackgroundScheduler
from apscheduler.triggers.cron import CronTrigger
from apscheduler.triggers.interval import IntervalTrigger
from datetime import datetime, timedelta
from sqlalchemy import and_
from src.core.database import SessionLocal
from src.core.services import PressureService, AlarmService, PatrolService, ReportService
from src.core import models


scheduler = BackgroundScheduler()


def hourly_pressure_aggregate():
    db = SessionLocal()
    try:
        pressure_service = PressureService(db)
        pressure_service.aggregate_hourly_data()

        check_pressure_for_leaks(db)

        alarm_service = AlarmService(db)
        alarm_service.escalate_alarms()
        alarm_service.send_reminders()

    finally:
        db.close()


def check_pressure_for_leaks(db):
    from src.core.services import PressureService, AlarmService

    pressure_service = PressureService(db)
    alarm_service = AlarmService(db)

    pipelines = db.query(models.Pipeline).filter(
        models.Pipeline.status == models.PipelineStatus.ACTIVE
    ).all()

    for pipeline in pipelines:
        adjacent_pairs = pressure_service.get_adjacent_points(pipeline.id)

        for start_point, end_point in adjacent_pairs:
            is_suspected, details = pressure_service.check_leak_suspected(
                start_point.id, end_point.id, pipeline.id
            )

            diff = details.get("pressure_diff")
            if diff is not None:
                pressure_service.record_pressure_diff(
                    pipeline.id, start_point.id, end_point.id, diff
                )

            if is_suspected:
                existing_alarm = db.query(models.Alarm).filter(
                    and_(
                        models.Alarm.pipeline_id == pipeline.id,
                        models.Alarm.status.notin_([
                            models.AlarmStatus.RESOLVED,
                            models.AlarmStatus.FALSE_ALARM
                        ]),
                        models.Alarm.created_at >= datetime.utcnow() - timedelta(hours=1)
                    )
                ).first()

                if not existing_alarm:
                    pd = pressure_service.record_pressure_diff(
                        pipeline.id, start_point.id, end_point.id, diff
                    )
                    alarm_service.create_leak_alarm(pipeline.id, pd.id, details)


def daily_tasks():
    db = SessionLocal()
    try:
        patrol_service = PatrolService(db)
        patrol_service.create_daily_plan()

        report_service = ReportService(db)
        report_service.generate_daily_report()

    finally:
        db.close()


def start_scheduler():
    scheduler.add_job(
        hourly_pressure_aggregate,
        IntervalTrigger(hours=1),
        id="hourly_aggregate",
        replace_existing=True
    )

    scheduler.add_job(
        daily_tasks,
        CronTrigger(hour=0, minute=1),
        id="daily_tasks",
        replace_existing=True
    )

    scheduler.start()


def stop_scheduler():
    scheduler.shutdown()
