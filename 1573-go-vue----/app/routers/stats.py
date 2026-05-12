from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session
from app.database import get_db
from app.services import stats_service
from app.models import Alert
from typing import List
from datetime import datetime

router = APIRouter()


@router.get("/hourly")
def get_hourly_stats(db: Session = Depends(get_db)):
    return stats_service.get_hourly_stats(db)


@router.get("/daily")
def get_daily_stats(db: Session = Depends(get_db)):
    return stats_service.get_daily_stats(db)


@router.get("/alerts")
def get_alerts(resolved: bool = False, db: Session = Depends(get_db)):
    alerts = db.query(Alert).filter(Alert.resolved == resolved).order_by(Alert.timestamp.desc()).all()
    return alerts
