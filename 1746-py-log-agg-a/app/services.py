from sqlalchemy.orm import Session
from sqlalchemy import func, and_, or_
from datetime import datetime, timedelta
from typing import List, Optional, Dict
from app.models import LogEntry
from app.schemas import LogCreate, LogLevel, TimeWindow
from app.config import settings


def check_trace_exceptional(db: Session, trace_id: str) -> bool:
    return db.query(LogEntry).filter(
        and_(LogEntry.trace_id == trace_id, LogEntry.level == "ERROR")
    ).count() > 0


def mark_trace_exceptional(db: Session, trace_id: str) -> None:
    db.query(LogEntry).filter(LogEntry.trace_id == trace_id).update(
        {LogEntry.is_exceptional: True}
    )
    db.commit()


def create_log(db: Session, log_data: LogCreate) -> LogEntry:
    log = LogEntry(
        service_name=log_data.service_name,
        level=log_data.level.value,
        message=log_data.message,
        trace_id=log_data.trace_id,
    )
    db.add(log)
    db.commit()
    db.refresh(log)

    if log.trace_id and log.level == "ERROR":
        mark_trace_exceptional(db, log.trace_id)
    elif log.trace_id and check_trace_exceptional(db, log.trace_id):
        db.query(LogEntry).filter(LogEntry.id == log.id).update(
            {LogEntry.is_exceptional: True}
        )
        db.commit()
        db.refresh(log)

    return log


def query_logs(
    db: Session,
    service_name: Optional[str] = None,
    level: Optional[str] = None,
    start_time: Optional[datetime] = None,
    end_time: Optional[datetime] = None,
    keyword: Optional[str] = None,
    trace_id: Optional[str] = None,
    is_exceptional: Optional[bool] = None,
    limit: int = 100,
    offset: int = 0,
) -> List[LogEntry]:
    query = db.query(LogEntry)

    if service_name:
        query = query.filter(LogEntry.service_name == service_name)
    if level:
        query = query.filter(LogEntry.level == level)
    if start_time:
        query = query.filter(LogEntry.created_at >= start_time)
    if end_time:
        query = query.filter(LogEntry.created_at <= end_time)
    if keyword:
        query = query.filter(LogEntry.message.contains(keyword))
    if trace_id:
        query = query.filter(LogEntry.trace_id == trace_id)
    if is_exceptional is not None:
        query = query.filter(LogEntry.is_exceptional == is_exceptional)

    return query.order_by(LogEntry.created_at.desc()).offset(offset).limit(limit).all()


def get_trace_aggregate(db: Session, trace_id: str) -> Optional[Dict]:
    logs = db.query(LogEntry).filter(LogEntry.trace_id == trace_id).order_by(
        LogEntry.created_at.asc()
    ).all()

    if not logs:
        return None

    is_exceptional = any(log.level == "ERROR" or log.is_exceptional for log in logs)

    return {
        "trace_id": trace_id,
        "logs": logs,
        "is_exceptional": is_exceptional,
    }


def get_time_window_start(window: TimeWindow) -> datetime:
    now = datetime.utcnow()
    if window == TimeWindow.ONE_HOUR:
        return now - timedelta(hours=1)
    elif window == TimeWindow.TWENTY_FOUR_HOURS:
        return now - timedelta(hours=24)
    else:
        return now - timedelta(days=7)


def get_stats(db: Session, window: TimeWindow) -> List[Dict]:
    start_time = get_time_window_start(window)

    result = db.query(
        LogEntry.service_name,
        LogEntry.level,
        func.count(LogEntry.id).label("count")
    ).filter(
        LogEntry.created_at >= start_time
    ).group_by(
        LogEntry.service_name,
        LogEntry.level
    ).all()

    stats_dict: Dict[str, Dict[str, int]] = {}

    for row in result:
        service = row.service_name
        level = row.level
        count = row.count

        if service not in stats_dict:
            stats_dict[service] = {
                "service_name": service,
                "debug_count": 0,
                "info_count": 0,
                "warn_count": 0,
                "error_count": 0,
                "total_count": 0,
            }

        key = f"{level.lower()}_count"
        if key in stats_dict[service]:
            stats_dict[service][key] += count
        stats_dict[service]["total_count"] += count

    return list(stats_dict.values())


def cleanup_expired_logs(db: Session) -> Dict[str, int]:
    now = datetime.utcnow()
    base_cutoff = now - timedelta(days=settings.RETENTION_DAYS)
    error_cutoff = now - timedelta(days=settings.ERROR_RETENTION_DAYS)

    deleted = {
        "DEBUG": 0,
        "INFO": 0,
        "WARN": 0,
        "ERROR": 0,
    }

    debug_count = db.query(LogEntry).filter(
        and_(
            LogEntry.level == "DEBUG",
            LogEntry.created_at < base_cutoff
        )
    ).delete(synchronize_session=False)
    deleted["DEBUG"] = debug_count

    info_count = db.query(LogEntry).filter(
        and_(
            LogEntry.level == "INFO",
            LogEntry.created_at < base_cutoff
        )
    ).delete(synchronize_session=False)
    deleted["INFO"] = info_count

    warn_count = db.query(LogEntry).filter(
        and_(
            LogEntry.level == "WARN",
            LogEntry.created_at < base_cutoff
        )
    ).delete(synchronize_session=False)
    deleted["WARN"] = warn_count

    error_count = db.query(LogEntry).filter(
        and_(
            LogEntry.level == "ERROR",
            LogEntry.created_at < error_cutoff
        )
    ).delete(synchronize_session=False)
    deleted["ERROR"] = error_count

    db.commit()

    return deleted
