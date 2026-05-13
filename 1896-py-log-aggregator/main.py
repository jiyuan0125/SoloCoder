import os
from typing import Optional, List, Dict, Any
from fastapi import FastAPI, HTTPException, Request, Query
from fastapi.responses import JSONResponse
from contextlib import asynccontextmanager

from storage import LogStore
from query_engine import QueryEngine
from models import LogEntry
from timestamp_parser import parse_timestamp


store: LogStore = None
query_engine: QueryEngine = None
MAX_BATCH_SIZE = 500


@asynccontextmanager
async def lifespan(app: FastAPI):
    global store, query_engine
    store = LogStore(data_dir="./log_data", retention_days=7)
    store.start()
    query_engine = QueryEngine(store)
    yield
    store.stop()


app = FastAPI(title="Log Aggregator", lifespan=lifespan)


@app.post("/api/v1/logs/batch")
async def push_logs(request: Request):
    try:
        body = await request.json()
    except Exception:
        raise HTTPException(status_code=400, detail="Invalid JSON body")
    
    if not isinstance(body, list):
        raise HTTPException(status_code=400, detail="Expected an array of log entries")
    
    was_truncated = False
    if len(body) > MAX_BATCH_SIZE:
        body = body[:MAX_BATCH_SIZE]
        was_truncated = True
    
    valid_entries = []
    for idx, item in enumerate(body):
        if not isinstance(item, dict):
            raise HTTPException(
                status_code=400,
                detail={"error": "Invalid log entry", "line_number": idx}
            )
        
        for field in ["timestamp", "service", "level", "message"]:
            if field not in item:
                raise HTTPException(
                    status_code=400,
                    detail={"error": f"Missing required field: {field}", "line_number": idx}
                )
        
        try:
            timestamp = parse_timestamp(item["timestamp"])
        except Exception as e:
            raise HTTPException(
                status_code=400,
                detail={"error": f"Invalid timestamp: {e}", "line_number": idx}
            )
        
        if not isinstance(item["service"], str) or not item["service"].strip():
            raise HTTPException(
                status_code=400,
                detail={"error": "service must be a non-empty string", "line_number": idx}
            )
        
        if not isinstance(item["level"], str) or not item["level"].strip():
            raise HTTPException(
                status_code=400,
                detail={"error": "level must be a non-empty string", "line_number": idx}
            )
        
        if not isinstance(item["message"], str):
            raise HTTPException(
                status_code=400,
                detail={"error": "message must be a string", "line_number": idx}
            )
        
        entry = LogEntry.from_dict(item, timestamp)
        valid_entries.append(entry)
    
    store.enqueue_entries(valid_entries)
    
    response = {
        "received": len(valid_entries),
        "status": "accepted",
    }
    
    if was_truncated:
        response["warning"] = f"Batch truncated to {MAX_BATCH_SIZE} entries (received more)"
    
    return JSONResponse(status_code=202, content=response)


@app.get("/api/v1/logs")
async def query_logs(
    service: Optional[str] = None,
    level: Optional[str] = None,
    start_time: Optional[float] = Query(None, description="Unix timestamp (seconds)"),
    end_time: Optional[float] = Query(None, description="Unix timestamp (seconds)"),
    message: Optional[str] = None,
    page: int = Query(1, ge=1),
    page_size: int = Query(50, ge=1, le=200),
):
    result = query_engine.query(
        service=service,
        level=level,
        start_time=start_time,
        end_time=end_time,
        message_contains=message,
        page=page,
        page_size=page_size,
    )
    
    return {
        "total": result.total,
        "page": result.page,
        "page_size": result.page_size,
        "logs": [entry.to_dict() for entry in result.logs],
    }


@app.get("/api/v1/logs/trace/{trace_id}")
async def query_trace(trace_id: str):
    logs = query_engine.query_by_trace(trace_id)
    return {
        "trace_id": trace_id,
        "count": len(logs),
        "logs": [entry.to_dict() for entry in logs],
    }


@app.get("/api/v1/stats")
async def get_stats():
    return query_engine.get_statistics()


@app.get("/api/v1/health")
async def health():
    stats = store.get_failure_stats()
    return {
        "status": "healthy",
        "queue_size": store.write_queue.qsize(),
        "total_logs": len(store.all_logs),
        "write_failures": stats["total_failed"],
        "services": sorted(list(store.get_services())),
    }


@app.post("/api/v1/config/retention")
async def set_retention(request: Request):
    try:
        body = await request.json()
    except Exception:
        raise HTTPException(status_code=400, detail="Invalid JSON body")
    
    if not isinstance(body, dict):
        raise HTTPException(status_code=400, detail="Expected object")
    
    for service, days in body.items():
        if not isinstance(days, int) or days <= 0:
            raise HTTPException(
                status_code=400,
                detail=f"Retention days must be positive integer for service {service}"
            )
        store.set_service_retention(service, days)
    
    return {"status": "ok", "retention": store.service_retention.copy()}


if __name__ == "__main__":
    import uvicorn
    port = int(os.environ.get("PORT", 8000))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=False)
