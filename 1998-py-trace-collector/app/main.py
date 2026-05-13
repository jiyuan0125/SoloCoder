import os
from typing import List, Optional
from fastapi import FastAPI, HTTPException, Query
from fastapi.responses import JSONResponse

from .models import Span, ServiceStats, SpanTreeNode
from .storage import TraceStorage
from .services import TraceService

MAX_BATCH_SIZE = 500

storage = TraceStorage()
trace_service = TraceService(storage)

app = FastAPI(
    title="Lightweight Trace Collector",
    description="A lightweight distributed tracing service",
    version="1.0.0"
)


@app.post("/spans", status_code=201)
def create_span(span: Span):
    storage.add_span(span)
    return {"message": "ok"}


@app.post("/spans/batch", status_code=201)
def create_spans_batch(spans: List[Span]):
    if len(spans) > MAX_BATCH_SIZE:
        raise HTTPException(
            status_code=400,
            detail=f"Batch size exceeds limit: max {MAX_BATCH_SIZE} spans per batch"
        )
    storage.add_spans(spans)
    return {"message": "ok", "count": len(spans)}


@app.get("/traces/{trace_id}", response_model=List[SpanTreeNode])
def get_trace(trace_id: str):
    tree = trace_service.build_trace_tree(trace_id)
    if tree is None:
        raise HTTPException(status_code=404, detail="Trace not found")
    return tree


@app.get("/spans/service/{service_name}", response_model=List[Span])
def get_spans_by_service(
    service_name: str,
    start_time: Optional[float] = Query(None, description="Unix timestamp start"),
    end_time: Optional[float] = Query(None, description="Unix timestamp end")
):
    spans = storage.get_spans_by_service(
        service_name=service_name,
        start_time=start_time,
        end_time=end_time
    )
    return spans


@app.get("/spans/orphans", response_model=List[Span])
def get_orphan_spans():
    spans = storage.get_orphan_spans()
    return spans


@app.get("/stats/services", response_model=List[ServiceStats])
def get_service_stats():
    stats = trace_service.get_service_stats()
    return stats


@app.get("/health")
def health_check():
    return {"status": "ok"}


def main():
    import uvicorn
    port = int(os.environ.get("PORT", 8000))
    uvicorn.run(app, host="0.0.0.0", port=port)


if __name__ == "__main__":
    main()
