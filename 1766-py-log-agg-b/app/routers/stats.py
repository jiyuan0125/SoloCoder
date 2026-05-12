from datetime import datetime, timedelta
from typing import Optional, List, Dict

from fastapi import APIRouter, Depends, Query
from sqlalchemy import select, func, desc, case, extract
from sqlalchemy.ext.asyncio import AsyncSession

from app.database import get_db
from app.models import LogEntry
from app.schemas import (
    DashboardStats,
    ServiceStats,
    LevelStats,
    HourlyStats,
)
from app.log_cleanup import cleanup_service

router = APIRouter(prefix="/stats", tags=["stats"])


@router.get("/dashboard", response_model=DashboardStats)
async def get_dashboard_stats(
    hours: int = Query(24, ge=1, le=24 * 7),
    db: AsyncSession = Depends(get_db),
):
    since = datetime.utcnow() - timedelta(hours=hours)

    total_stmt = select(func.count(LogEntry.id))
    total_result = await db.execute(total_stmt)
    total_logs = total_result.scalar() or 0

    unique_services_stmt = select(
        func.count(func.distinct(LogEntry.service_name))
    )
    unique_result = await db.execute(unique_services_stmt)
    unique_services = unique_result.scalar() or 0

    level_stmt = (
        select(
            LogEntry.level,
            func.count(LogEntry.id).label("count"),
        )
        .group_by(LogEntry.level)
        .where(LogEntry.timestamp >= since)
    )
    level_result = await db.execute(level_stmt)
    level_counts = {row.level: row.count for row in level_result.all()}

    by_level = []
    time_window_total = sum(level_counts.values())
    for level in ["DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"]:
        count = level_counts.get(level, 0)
        pct = (count / time_window_total * 100) if time_window_total > 0 else 0.0
        by_level.append(LevelStats(level=level, count=count, percentage=pct))

    service_stmt = (
        select(
            LogEntry.service_name,
            func.count(LogEntry.id).label("total"),
            func.max(LogEntry.timestamp).label("last_time"),
        )
        .group_by(LogEntry.service_name)
        .order_by(desc("total"))
        .limit(10)
    )
    service_result = await db.execute(service_stmt)
    service_rows = service_result.all()

    level_by_service_stmt = (
        select(
            LogEntry.service_name,
            LogEntry.level,
            func.count(LogEntry.id).label("count"),
        )
        .where(LogEntry.service_name.in_([r.service_name for r in service_rows]))
        .group_by(LogEntry.service_name, LogEntry.level)
    )
    level_by_service_result = await db.execute(level_by_service_stmt)
    level_by_service = {}
    for row in level_by_service_result.all():
        if row.service_name not in level_by_service:
            level_by_service[row.service_name] = {}
        level_by_service[row.service_name][row.level] = row.count

    top_services = []
    for row in service_rows:
        by_level_svc = level_by_service.get(row.service_name, {})
        all_levels = {}
        for lvl in ["DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"]:
            all_levels[lvl] = by_level_svc.get(lvl, 0)
        top_services.append(
            ServiceStats(
                service_name=row.service_name,
                total_logs=row.total,
                by_level=all_levels,
                last_log_time=row.last_time.isoformat() if row.last_time else None,
            )
        )

    hourly_stmt = (
        select(
            extract("year", LogEntry.timestamp).label("year"),
            extract("month", LogEntry.timestamp).label("month"),
            extract("day", LogEntry.timestamp).label("day"),
            extract("hour", LogEntry.timestamp).label("hour"),
            LogEntry.level,
            func.count(LogEntry.id).label("count"),
        )
        .where(LogEntry.timestamp >= since)
        .group_by("year", "month", "day", "hour", LogEntry.level)
        .order_by("year", "month", "day", "hour")
    )
    hourly_result = await db.execute(hourly_stmt)

    hourly_map: Dict[str, Dict[str, int]] = {}
    for row in hourly_result.all():
        hour_key = (
            f"{int(row.year):04d}-{int(row.month):02d}-{int(row.day):02d}"
            f"T{int(row.hour):02d}:00:00"
        )
        if hour_key not in hourly_map:
            hourly_map[hour_key] = {}
        hourly_map[hour_key][row.level] = row.count

    hourly_trend = []
    for hour_key in sorted(hourly_map.keys()):
        levels = hourly_map[hour_key]
        by_level_hour = {}
        for lvl in ["DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"]:
            by_level_hour[lvl] = levels.get(lvl, 0)
        hourly_trend.append(
            HourlyStats(
                hour=hour_key,
                total=sum(by_level_hour.values()),
                by_level=by_level_hour,
            )
        )

    time_range_stmt = select(
        func.min(LogEntry.timestamp).label("oldest"),
        func.max(LogEntry.timestamp).label("newest"),
    )
    time_result = await db.execute(time_range_stmt)
    time_row = time_result.one_or_none()

    return DashboardStats(
        total_logs=total_logs,
        unique_services=unique_services,
        by_level=by_level,
        top_services=top_services,
        hourly_trend=hourly_trend,
        oldest_log=time_row.oldest.isoformat() if time_row and time_row.oldest else None,
        newest_log=time_row.newest.isoformat() if time_row and time_row.newest else None,
    )


@router.get("/services")
async def list_services(db: AsyncSession = Depends(get_db)):
    stmt = (
        select(
            LogEntry.service_name,
            func.count(LogEntry.id).label("total"),
            func.max(LogEntry.timestamp).label("last_time"),
        )
        .group_by(LogEntry.service_name)
        .order_by(desc("total"))
    )
    result = await db.execute(stmt)
    rows = result.all()
    return {
        "services": [
            {
                "name": r.service_name,
                "total_logs": r.total,
                "last_log_time": r.last_time.isoformat() if r.last_time else None,
            }
            for r in rows
        ],
        "count": len(rows),
    }


@router.get("/levels")
async def get_level_distribution(
    hours: int = Query(24, ge=1, le=24 * 30),
    db: AsyncSession = Depends(get_db),
):
    since = datetime.utcnow() - timedelta(hours=hours)
    stmt = (
        select(
            LogEntry.level,
            func.count(LogEntry.id).label("count"),
        )
        .where(LogEntry.timestamp >= since)
        .group_by(LogEntry.level)
        .order_by(desc("count"))
    )
    result = await db.execute(stmt)
    rows = result.all()
    total = sum(r.count for r in rows)

    return {
        "time_window_hours": hours,
        "total": total,
        "by_level": [
            {
                "level": r.level,
                "count": r.count,
                "percentage": (r.count / total * 100) if total > 0 else 0.0,
            }
            for r in rows
        ],
    }


@router.post("/cleanup")
async def run_manual_cleanup(
    older_than_days: int = Query(30, ge=1),
):
    result = await cleanup_service.manual_cleanup(older_than_days)
    return result


@router.get("/health")
async def health_check(db: AsyncSession = Depends(get_db)):
    try:
        stmt = select(func.count(LogEntry.id))
        result = await db.execute(stmt)
        count = result.scalar() or 0
        return {
            "status": "healthy",
            "database": "connected",
            "log_count": count,
            "timestamp": datetime.utcnow().isoformat(),
        }
    except Exception as e:
        return {
            "status": "unhealthy",
            "database": "error",
            "error": str(e),
            "timestamp": datetime.utcnow().isoformat(),
        }
