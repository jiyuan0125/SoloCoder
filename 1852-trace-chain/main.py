import os
import uuid
from datetime import datetime, timedelta, timezone
from typing import List, Optional, Dict, Any, Union
from collections import defaultdict

from fastapi import FastAPI, HTTPException, Query
from pydantic import BaseModel, Field
from sqlalchemy import (
    create_engine, Column, String, DateTime, Text, Integer, Index
)
from sqlalchemy.orm import sessionmaker, declarative_base
from apscheduler.schedulers.background import BackgroundScheduler

DATABASE_URL = os.getenv("DATABASE_URL", "sqlite:///./trace_chain.db")
PORT = int(os.getenv("PORT", "8000"))
DATA_RETENTION_DAYS = 7

engine = create_engine(
    DATABASE_URL,
    connect_args={"check_same_thread": False} if DATABASE_URL.startswith("sqlite") else {}
)
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)
Base = declarative_base()


class SpanModel(Base):
    __tablename__ = "spans"

    id = Column(Integer, primary_key=True, autoincrement=True)
    trace_id = Column(String(36), nullable=False, index=True)
    span_id = Column(String(36), nullable=False, unique=True)
    operation = Column(String(255), nullable=False)
    start_time = Column(DateTime, nullable=False, index=True)
    end_time = Column(DateTime, nullable=False, index=True)
    parent_id = Column(String(36), nullable=True, index=True)
    service_name = Column(String(255), nullable=True, index=True)
    tags = Column(Text, nullable=True)
    duration_ms = Column(Integer, nullable=False, index=True)

    __table_args__ = (
        Index("ix_spans_trace_start", "trace_id", "start_time"),
        Index("ix_spans_service_start_duration", "service_name", "start_time", "duration_ms"),
    )


Base.metadata.create_all(bind=engine)


class SpanIn(BaseModel):
    trace_id: str = Field(..., min_length=1)
    span_id: str = Field(..., min_length=1)
    operation: str = Field(..., min_length=1)
    start_time: datetime
    end_time: datetime
    parent_id: Optional[str] = None
    service_name: Optional[str] = None
    tags: Optional[Dict[str, Any]] = None


class SpanOut(BaseModel):
    trace_id: str
    span_id: str
    operation: str
    start_time: datetime
    end_time: datetime
    parent_id: Optional[str]
    service_name: Optional[str]
    tags: Optional[Dict[str, Any]]
    duration_ms: int
    is_orphan: Optional[bool] = None
    children: Optional[List["SpanOut"]] = None


SpanOut.model_rebuild()


class SpanTreeResponse(BaseModel):
    root_spans: List[SpanOut]
    orphan_spans: List[SpanOut]


class PaginationMeta(BaseModel):
    page: int
    page_size: int
    total: int
    total_pages: int


class SlowSpansResponse(BaseModel):
    data: List[SpanOut]
    pagination: PaginationMeta


app = FastAPI(title="Trace Chain Service")


def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()


def calculate_duration_ms(start: datetime, end: datetime) -> int:
    delta = end - start
    return int(delta.total_seconds() * 1000)


def span_model_to_out(span: SpanModel, include_children: bool = False) -> SpanOut:
    tags = None
    if span.tags:
        try:
            import json
            tags = json.loads(span.tags)
        except Exception:
            tags = None

    return SpanOut(
        trace_id=span.trace_id,
        span_id=span.span_id,
        operation=span.operation,
        start_time=span.start_time,
        end_time=span.end_time,
        parent_id=span.parent_id,
        service_name=span.service_name,
        tags=tags,
        duration_ms=span.duration_ms,
        children=[] if include_children else None,
    )


def build_span_tree(spans: List[SpanModel]) -> SpanTreeResponse:
    span_map: Dict[str, SpanOut] = {}
    parent_to_children: Dict[str, List[str]] = defaultdict(list)
    all_span_ids: set = set()

    for span in spans:
        all_span_ids.add(span.span_id)
        span_out = span_model_to_out(span, include_children=True)
        span_map[span.span_id] = span_out
        if span.parent_id:
            parent_to_children[span.parent_id].append(span.span_id)

    root_spans: List[SpanOut] = []
    orphan_spans: List[SpanOut] = []

    for span in spans:
        span_out = span_map[span.span_id]
        if span.parent_id is None:
            root_spans.append(span_out)
        else:
            if span.parent_id not in all_span_ids:
                span_out.is_orphan = True
                orphan_spans.append(span_out)
            else:
                parent = span_map.get(span.parent_id)
                if parent and parent.children is not None:
                    parent.children.append(span_out)

    def sort_children(span_list: List[SpanOut]):
        for span in span_list:
            if span.children:
                span.children.sort(key=lambda x: x.start_time)
                sort_children(span.children)

    root_spans.sort(key=lambda x: x.start_time)
    orphan_spans.sort(key=lambda x: x.start_time)
    sort_children(root_spans)

    return SpanTreeResponse(
        root_spans=root_spans,
        orphan_spans=orphan_spans
    )


@app.post("/spans")
async def report_spans(spans: Union[SpanIn, List[SpanIn]]):
    import json

    if isinstance(spans, SpanIn):
        spans = [spans]

    if not spans:
        raise HTTPException(status_code=400, detail="No spans provided")

    db = next(get_db())
    try:
        for span_in in spans:
            if span_in.start_time > span_in.end_time:
                raise HTTPException(
                    status_code=400,
                    detail=f"Invalid time range for span {span_in.span_id}: start_time must be <= end_time"
                )

            duration_ms = calculate_duration_ms(span_in.start_time, span_in.end_time)

            tags_json = None
            if span_in.tags:
                tags_json = json.dumps(span_in.tags)

            existing = db.query(SpanModel).filter(
                SpanModel.span_id == span_in.span_id
            ).first()

            if existing:
                existing.trace_id = span_in.trace_id
                existing.operation = span_in.operation
                existing.start_time = span_in.start_time
                existing.end_time = span_in.end_time
                existing.parent_id = span_in.parent_id
                existing.service_name = span_in.service_name
                existing.tags = tags_json
                existing.duration_ms = duration_ms
            else:
                span_db = SpanModel(
                    trace_id=span_in.trace_id,
                    span_id=span_in.span_id,
                    operation=span_in.operation,
                    start_time=span_in.start_time,
                    end_time=span_in.end_time,
                    parent_id=span_in.parent_id,
                    service_name=span_in.service_name,
                    tags=tags_json,
                    duration_ms=duration_ms,
                )
                db.add(span_db)

        db.commit()
        return {"status": "success", "reported": len(spans)}
    except HTTPException:
        db.rollback()
        raise
    except Exception as e:
        db.rollback()
        raise HTTPException(status_code=500, detail=str(e))
    finally:
        db.close()


@app.get("/traces/{trace_id}", response_model=SpanTreeResponse)
async def get_trace(trace_id: str):
    db = next(get_db())
    try:
        spans = db.query(SpanModel).filter(
            SpanModel.trace_id == trace_id
        ).order_by(SpanModel.start_time).all()

        if not spans:
            raise HTTPException(status_code=404, detail=f"Trace not found: {trace_id}")

        return build_span_tree(spans)
    finally:
        db.close()


@app.get("/slow-spans", response_model=SlowSpansResponse)
async def get_slow_spans(
    service_name: str = Query(..., description="Service name to filter"),
    start_time: datetime = Query(..., description="Start time (ISO format)"),
    end_time: datetime = Query(..., description="End time (ISO format)"),
    threshold_ms: int = Query(1000, ge=0, description="Slow threshold in milliseconds"),
    page: int = Query(1, ge=1, description="Page number"),
    page_size: int = Query(50, ge=1, le=500, description="Page size"),
):
    if start_time > end_time:
        raise HTTPException(status_code=400, detail="start_time must be <= end_time")

    db = next(get_db())
    try:
        query = db.query(SpanModel).filter(
            SpanModel.service_name == service_name,
            SpanModel.start_time >= start_time,
            SpanModel.start_time <= end_time,
            SpanModel.duration_ms >= threshold_ms,
        )

        total = query.count()

        spans = (
            query.order_by(SpanModel.duration_ms.desc())
            .offset((page - 1) * page_size)
            .limit(page_size)
            .all()
        )

        total_pages = (total + page_size - 1) // page_size

        span_out_list = [span_model_to_out(span) for span in spans]

        return SlowSpansResponse(
            data=span_out_list,
            pagination=PaginationMeta(
                page=page,
                page_size=page_size,
                total=total,
                total_pages=total_pages,
            ),
        )
    finally:
        db.close()


@app.get("/health")
async def health_check():
    return {"status": "ok"}


def cleanup_old_data():
    db = next(get_db())
    try:
        cutoff_time = datetime.now(timezone.utc) - timedelta(days=DATA_RETENTION_DAYS)
        deleted = db.query(SpanModel).filter(
            SpanModel.start_time < cutoff_time
        ).delete()
        db.commit()
        if deleted > 0:
            print(f"[Cleanup] Deleted {deleted} old spans (older than {DATA_RETENTION_DAYS} days)")
    except Exception as e:
        db.rollback()
        print(f"[Cleanup] Error: {e}")
    finally:
        db.close()


scheduler = BackgroundScheduler()


@app.on_event("startup")
async def startup_event():
    scheduler.add_job(
        cleanup_old_data,
        "interval",
        hours=24,
        id="cleanup_old_spans",
        replace_existing=True,
        next_run_time=datetime.now(timezone.utc) + timedelta(minutes=1),
    )
    scheduler.start()


@app.on_event("shutdown")
async def shutdown_event():
    scheduler.shutdown()


if __name__ == "__main__":
    import uvicorn
    uvicorn.run("main:app", host="0.0.0.0", port=PORT, reload=False)
