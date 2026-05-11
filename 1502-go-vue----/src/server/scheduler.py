from apscheduler.schedulers.background import BackgroundScheduler

from core.database import SessionLocal
from core import services

scheduler: BackgroundScheduler = None


def check_and_execute_feeding_plans():
    db = SessionLocal()
    try:
        plans = services.get_pending_feeding_plans(db)
        for plan in plans:
            services.execute_feeding_plan(db, plan.id)
    finally:
        db.close()


def start_scheduler():
    global scheduler
    scheduler = BackgroundScheduler()
    scheduler.add_job(
        check_and_execute_feeding_plans,
        trigger="interval",
        seconds=60,
        id="feeding_check",
        replace_existing=True,
    )
    scheduler.start()


def stop_scheduler():
    global scheduler
    if scheduler and scheduler.running:
        scheduler.shutdown()
