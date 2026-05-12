from datetime import datetime
from typing import Optional, List

from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy import select, func, update, or_, and_
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload

from app.database import get_db
from app.models import LogEntry, TraceMetadata
from app.schemas import (
    LogEntryCreate,
    LogEntryBatch,
    LogEntryResponse,
    PaginatedResponse,
    TraceInfo,
    TraceDetail,
    LogLevel,
)

router = APIRouter(prefix="/logs", tags=["logs"])


def _log_to_response(log: LogEntry) -> LogEntryResponse:
    return LogEntryResponse(**log.to_dict())


def _trace_to_response(trace: TraceMetadata) -> TraceInfo:
    return TraceInfo(**trace.to_dict())


@router.post("/ingest", status_code=201)
async def ingest_single(
    entry: LogEntryCreate,
    db: AsyncSession = Depends(get_db),
):
    timestamp = entry.timestamp or datetime.utcnow()
    db_entry = LogEntry(
        service_name=entry.service_name,
        level=entry.level.value,
        message=entry.message,
        timestamp=timestamp,
        trace_id=entry.trace_id,
        span_id=entry.span_id,
        source_host=entry.source_host,
        module=entry.module,
        function=entry.function,
        line_number=entry.line_number,
        extra=entry.extra,
    )
    db.add(db_entry)
    await db.commit()
    await db.refresh(db_entry)

    if entry.trace_id:
        await _update_trace_metadata(
            db,
            entry.trace_id,
            timestamp,
            entry.service_name,
            entry.level.value,
        )

    return {"id": db_entry.id, "status": "ok"}


@router.post("/ingest/batch", status_code=201)
async def ingest_batch(
    batch: LogEntryBatch,
    db: AsyncSession = Depends(get_db),
):
    entries = []
    trace_updates = {}

    for entry in batch.entries:
        timestamp = entry.timestamp or datetime.utcnow()
        db_entry = LogEntry(
            service_name=entry.service_name,
            level=entry.level.value,
            message=entry.message,
            timestamp=timestamp,
            trace_id=entry.trace_id,
            span_id=entry.span_id,
            source_host=entry.source_host,
            module=entry.module,
            function=entry.function,
            line_number=entry.line_number,
            extra=entry.extra,
        )
        entries.append(db_entry)

        if entry.trace_id:
            if entry.trace_id not in trace_updates:
                trace_updates[entry.trace_id] = {
                    "min_time": timestamp,
                    "max_time": timestamp,
                    "services": {entry.service_name},
                    "levels": {entry.level.value},
                }
            else:
                tu = trace_updates[entry.trace_id]
                tu["min_time"] = min(tu["min_time"], timestamp)
                tu["max_time"] = max(tu["max_time"], timestamp)
                tu["services"].add(entry.service_name)
                tu["levels"].add(entry.level.value)

    db.add_all(entries)
    await db.commit()

    for trace_id, data in trace_updates.items():
        await _update_trace_metadata(
            db,
            trace_id,
            data["min_time"],
            list(data["services"])[0],
            list(data["levels"])[0],
            data["max_time"],
            len(data["services"]),
        )

    return {"count": len(entries), "status": "ok"}


@router.get("/search", response_model=PaginatedResponse)
async def search_logs(
    service_name: Optional[str] = None,
    level: Optional[str] = None,
    levels: Optional[List[str]] = Query(None),
    trace_id: Optional[str] = None,
    message_contains: Optional[str] = None,
    source_host: Optional[str] = None,
    module: Optional[str] = None,
    start_time: Optional[datetime] = None,
    end_time: Optional[datetime] = None,
    page: int = Query(1, ge=1),
    page_size: int = Query(100, ge=1, le=1000),
    sort_by: str = Query("timestamp", pattern="^(timestamp|created_at|id)$"),
    sort_order: str = Query("desc", pattern="^(asc|desc)$"),
    db: AsyncSession = Depends(get_db),
):
    conditions = []

    if service_name:
        conditions.append(LogEntry.service_name == service_name)

    if level:
        conditions.append(LogEntry.level == level.upper())

    if levels:
        level_list = [l.upper() for l in levels]
        conditions.append(LogEntry.level.in_(level_list))

    if trace_id:
        conditions.append(LogEntry.trace_id == trace_id)

    if message_contains:
        conditions.append(LogEntry.message.ilike(f"%{message_contains}%"))

    if source_host:
        conditions.append(LogEntry.source_host == source_host)

    if module:
        conditions.append(LogEntry.module.ilike(f"%{module}%"))

    if start_time:
        conditions.append(LogEntry.timestamp >= start_time)

    if end_time:
        conditions.append(LogEntry.timestamp <= end_time)

    count_stmt = select(func.count(LogEntry.id))
    if conditions:
        count_stmt = count_stmt.where(and_(*conditions))

    count_result = await db.execute(count_stmt)
    total = count_result.scalar() or 0

    stmt = select(LogEntry)
    if conditions:
        stmt = stmt.where(and_(*conditions))

    sort_column = getattr(LogEntry, sort_by)
    stmt = stmt.order_by(
        sort_column.asc() if sort_order == "asc" else sort_column.desc()
    )

    offset = (page - 1) * page_size
    stmt = stmt.offset(offset).limit(page_size)

    result = await db.execute(stmt)
    logs = result.scalars().all()

    items = [_log_to_response(log) for log in logs]

    return PaginatedResponse(
        items=items,
        total=total,
        page=page,
        page_size=page_size,
        has_more=(offset + len(items)) < total,
    )


@router.get("/trace/{trace_id}", response_model=TraceDetail)
async def get_trace_detail(
    trace_id: str,
    db: AsyncSession = Depends(get_db),
):
    trace_stmt = select(TraceMetadata).where(
        TraceMetadata.trace_id == trace_id
    )
    trace_result = await db.execute(trace_stmt)
    trace = trace_result.scalar_one_or_none()

    if not trace:
        logs_stmt = (
            select(LogEntry)
            .where(LogEntry.trace_id == trace_id)
            .order_by(LogEntry.timestamp.asc())
        )
        logs_result = await db.execute(logs_stmt)
        logs = logs_result.scalars().all()

        if not logs:
            raise HTTPException(status_code=404, detail="Trace not found")

        timestamps = [log.timestamp for log in logs]
        services = list({log.service_name for log in logs})

        trace_info = TraceInfo(
            trace_id=trace_id,
            first_seen=min(timestamps).isoformat() if timestamps else None,
            last_seen=max(timestamps).isoformat() if timestamps else None,
            service_count=len(services),
            entry_count=len(logs),
            root_service=services[0] if services else None,
            summary=None,
        )

        return TraceDetail(
            trace=trace_info,
            logs=[_log_to_response(log) for log in logs],
        )

    logs_stmt = (
        select(LogEntry)
        .where(LogEntry.trace_id == trace_id)
        .order_by(LogEntry.timestamp.asc())
    )
    logs_result = await db.execute(logs_stmt)
    logs = logs_result.scalars().all()

    return TraceDetail(
        trace=_trace_to_response(trace),
        logs=[_log_to_response(log) for log in logs],
    )


async def _update_trace_metadata(
    db: AsyncSession,
    trace_id: str,
    seen_time: datetime,
    service_name: str,
    level: str,
    max_time: Optional[datetime] = None,
    service_count: int = 1,
):
    existing_stmt = select(TraceMetadata).where(
        TraceMetadata.trace_id == trace_id
    )
    result = await db.execute(existing_stmt)
    existing = result.scalar_one_or_none()

    if not existing:
        metadata = TraceMetadata(
            trace_id=trace_id,
            first_seen=seen_time,
            last_seen=max_time or seen_time,
            service_count=service_count,
            entry_count=1,
            root_service=service_name,
            summary={
                "services": [service_name],
                "levels": [level],
            },
        )
        db.add(metadata)
    else:
        first_seen = min(existing.first_seen, seen_time)
        last_seen = max(
            existing.last_seen, max_time or seen_time
        )

        summary = existing.summary or {}
        services = set(summary.get("services", [existing.root_service or service_name]))
        services.add(service_name)

        levels = set(summary.get("levels", [level]))
        levels.add(level)

        existing.first_seen = first_seen
        existing.last_seen = last_seen
        existing.entry_count = (existing.entry_count or 0) + 1
        existing.service_count = len(services)
        existing.summary = {
            "services": list(services),
            "levels": list(levels),
        }

    await db.commit()
