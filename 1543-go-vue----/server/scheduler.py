from apscheduler.schedulers.background import BackgroundScheduler
from server.database import SessionLocal
from server.services import evaluate_flood_situation
from server.models import MonitoringStation
import random
from datetime import datetime

scheduler = BackgroundScheduler()


def scheduled_collect():
    db = SessionLocal()
    try:
        stations = db.query(MonitoringStation).filter(MonitoringStation.is_active == True).all()
        print(f"[{datetime.utcnow().isoformat()}] Scheduler: Collecting from {len(stations)} stations")
    finally:
        db.close()


def scheduled_evaluate():
    db = SessionLocal()
    try:
        results = evaluate_flood_situation(db)
        print(f"[{datetime.utcnow().isoformat()}] Scheduler: Evaluated {len(results)} active events")
    finally:
        db.close()


def start_scheduler():
    scheduler.add_job(scheduled_collect, 'interval', minutes=10, id='collect_job')
    scheduler.add_job(scheduled_evaluate, 'interval', hours=6, id='evaluate_job')
    scheduler.start()
    print("Scheduler started: 10min collection, 6h evaluation")


def stop_scheduler():
    scheduler.shutdown()
