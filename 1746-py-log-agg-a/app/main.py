from fastapi import FastAPI, Depends, HTTPException, Query
from fastapi.responses import JSONResponse
from sqlalchemy.orm import Session
from typing import Optional, List
from datetime import datetime
from app.database import engine, get_db, Base
from app.models import LogEntry
from app.schemas import (
    LogCreate,
    LogResponse,
    TraceAggregate,
    LogLevel,
    TimeWindow,
    StatsResponse,
    ServiceStats,
)
from app.services import (
    create_log,
    query_logs,
    get_trace_aggregate,
    get_stats,
    cleanup_expired_logs,
)
from app.config import settings

Base.metadata.create_all(bind=engine)

app = FastAPI(title="Log Aggregator Service")


@app.post("/api/logs", response_model=LogResponse, status_code=201)
def ingest_log(log_data: LogCreate, db: Session = Depends(get_db)):
    return create_log(db, log_data)


@app.post("/api/logs/batch", status_code=201)
def ingest_logs_batch(logs: List[LogCreate], db: Session = Depends(get_db)):
    created = []
    for log_data in logs:
        created.append(create_log(db, log_data))
    return {"ingested": len(created), "logs": [LogResponse.from_orm(l) for l in created]}


@app.get("/api/logs", response_model=List[LogResponse])
def search_logs(
    service_name: Optional[str] = None,
    level: Optional[LogLevel] = None,
    start_time: Optional[datetime] = None,
    end_time: Optional[datetime] = None,
    keyword: Optional[str] = None,
    trace_id: Optional[str] = None,
    is_exceptional: Optional[bool] = None,
    limit: int = Query(100, ge=1, le=1000),
    offset: int = Query(0, ge=0),
    db: Session = Depends(get_db),
):
    logs = query_logs(
        db=db,
        service_name=service_name,
        level=level.value if level else None,
        start_time=start_time,
        end_time=end_time,
        keyword=keyword,
        trace_id=trace_id,
        is_exceptional=is_exceptional,
        limit=limit,
        offset=offset,
    )
    return logs


@app.get("/api/traces/{trace_id}", response_model=TraceAggregate)
def get_trace(trace_id: str, db: Session = Depends(get_db)):
    result = get_trace_aggregate(db, trace_id)
    if not result:
        raise HTTPException(status_code=404, detail="Trace not found")
    return result


@app.get("/api/stats", response_model=StatsResponse)
def get_statistics(
    window: TimeWindow = Query(TimeWindow.TWENTY_FOUR_HOURS),
    db: Session = Depends(get_db),
):
    stats = get_stats(db, window)
    return StatsResponse(
        time_window=window.value,
        stats=[ServiceStats(**s) for s in stats],
    )


@app.post("/api/cleanup")
def trigger_cleanup(db: Session = Depends(get_db)):
    deleted = cleanup_expired_logs(db)
    return {
        "message": "Cleanup completed",
        "deleted": deleted,
        "retention_days": {
            "base": settings.RETENTION_DAYS,
            "error": settings.ERROR_RETENTION_DAYS,
        },
    }


@app.get("/health")
def health_check():
    return {"status": "healthy"}
