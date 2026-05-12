from sqlalchemy.orm import Session
from sqlalchemy import func
from app.models import SwitchLog, BlockLog
from datetime import datetime, timedelta
from typing import Dict


def get_hourly_stats(db: Session) -> Dict:
    now = datetime.utcnow()
    start_time = now - timedelta(hours=1)
    
    switch_ops = db.query(SwitchLog).filter(
        SwitchLog.timestamp >= start_time
    ).count()
    
    switch_faults = db.query(SwitchLog).filter(
        SwitchLog.timestamp >= start_time,
        SwitchLog.is_fault == True
    ).count()
    
    total_blocks = db.query(BlockLog).filter(
        BlockLog.timestamp >= start_time
    ).count()
    
    occupied_blocks = db.query(BlockLog).filter(
        BlockLog.timestamp >= start_time,
        BlockLog.occupied == True
    ).count()
    
    occupancy_rate = (occupied_blocks / total_blocks * 100) if total_blocks > 0 else 0.0
    
    return {
        "period": "hour",
        "switch_operations": switch_ops,
        "switch_faults": switch_faults,
        "block_occupancy_rate": round(occupancy_rate, 2)
    }


def get_daily_stats(db: Session) -> Dict:
    now = datetime.utcnow()
    start_time = now - timedelta(days=1)
    
    switch_ops = db.query(SwitchLog).filter(
        SwitchLog.timestamp >= start_time
    ).count()
    
    switch_faults = db.query(SwitchLog).filter(
        SwitchLog.timestamp >= start_time,
        SwitchLog.is_fault == True
    ).count()
    
    total_blocks = db.query(BlockLog).filter(
        BlockLog.timestamp >= start_time
    ).count()
    
    occupied_blocks = db.query(BlockLog).filter(
        BlockLog.timestamp >= start_time,
        BlockLog.occupied == True
    ).count()
    
    occupancy_rate = (occupied_blocks / total_blocks * 100) if total_blocks > 0 else 0.0
    
    return {
        "period": "day",
        "switch_operations": switch_ops,
        "switch_faults": switch_faults,
        "block_occupancy_rate": round(occupancy_rate, 2)
    }
