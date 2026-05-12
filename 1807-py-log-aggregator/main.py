import os
import json
import logging
import asyncio
from datetime import datetime, timedelta, timezone
from typing import Optional
from contextlib import asynccontextmanager

from fastapi import FastAPI, HTTPException, Request
from fastapi.responses import JSONResponse

from models import (
    LogEntry, LogLevel, BatchLogResponse,
    LogQueryResponse, StorageStats
)
from storage import storage

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

MAX_SINGLE_LOG_SIZE = 10 * 1024
REQUIRED_FIELDS = {'service', 'level', 'message', 'timestamp'}


@asynccontextmanager
async def lifespan(app: FastAPI):
    cleanup_task = asyncio.create_task(periodic_cleanup())
    yield
    cleanup_task.cancel()
    try:
        await cleanup_task
    except asyncio.CancelledError:
        pass


app = FastAPI(lifespan=lifespan, title="日志聚合服务")


async def periodic_cleanup():
    while True:
        try:
            deleted = storage.cleanup_expired(retention_hours=24)
            if deleted > 0:
                logger.info(f"定时清理完成: 删除 {deleted} 条过期日志")
        except Exception as e:
            logger.error(f"定时清理任务出错: {e}")
        await asyncio.sleep(3600)


@app.post("/api/logs/batch", response_model=BatchLogResponse, status_code=201)
async def batch_upload(request: Request):
    try:
        body = await request.body()
        data = json.loads(body)
    except json.JSONDecodeError:
        raise HTTPException(status_code=400, detail="JSON 解析失败")

    if not isinstance(data, dict) or 'logs' not in data:
        raise HTTPException(status_code=400, detail="请求体必须包含 logs 字段")

    logs = data['logs']
    if not isinstance(logs, list):
        raise HTTPException(status_code=400, detail="logs 必须是数组")

    if len(logs) > 1000:
        raise HTTPException(
            status_code=400,
            detail=f"单次批量上报最多 1000 条，当前 {len(logs)} 条"
        )

    received = 0
    rejected = 0
    first_missing_field = None

    for log_item in logs:
        try:
            if not isinstance(log_item, dict):
                rejected += 1
                continue

            log_json = json.dumps(log_item, ensure_ascii=False)
            if len(log_json.encode('utf-8')) > MAX_SINGLE_LOG_SIZE:
                rejected += 1
                logger.warning(f"日志条目超过 10KB 限制，已丢弃")
                continue

            missing = REQUIRED_FIELDS - set(log_item.keys())
            if missing:
                if first_missing_field is None:
                    first_missing_field = list(missing)[0]
                rejected += 1
                continue

            entry = LogEntry(
                service=log_item['service'],
                level=LogLevel(log_item['level'].upper()),
                message=log_item['message'],
                timestamp=log_item['timestamp']
            )
            storage.add(entry)
            received += 1

        except Exception:
            rejected += 1
            continue

    if first_missing_field is not None:
        raise HTTPException(
            status_code=400,
            detail=f"缺少必要字段: {first_missing_field}"
        )

    return BatchLogResponse(received=received, rejected=rejected)


def parse_datetime(value: str, field_name: str) -> datetime:
    try:
        return datetime.fromisoformat(value.replace('Z', '+00:00'))
    except ValueError:
        raise HTTPException(
            status_code=400,
            detail=f"{field_name} 格式错误，期望 ISO 8601 格式，如: 2024-01-01T12:00:00Z"
        )


@app.get("/api/logs", response_model=LogQueryResponse)
async def query_logs(
    service: Optional[str] = None,
    level: Optional[LogLevel] = None,
    start_time: Optional[str] = None,
    end_time: Optional[str] = None,
    page: int = 1,
    page_size: int = 50
):
    if page < 1:
        raise HTTPException(status_code=400, detail="page 必须大于 0")
    if page_size < 1 or page_size > 1000:
        raise HTTPException(status_code=400, detail="page_size 必须在 1-1000 之间")

    parsed_start = None
    parsed_end = None

    if start_time is None and end_time is None:
        parsed_end = datetime.now(timezone.utc)
        parsed_start = parsed_end - timedelta(hours=1)
    else:
        if start_time:
            parsed_start = parse_datetime(start_time, "start_time")
        if end_time:
            parsed_end = parse_datetime(end_time, "end_time")

    logs, total = storage.query(
        service=service,
        level=level,
        start_time=parsed_start,
        end_time=parsed_end,
        page=page,
        page_size=page_size
    )

    total_pages = (total + page_size - 1) // page_size if total > 0 else 0

    return LogQueryResponse(
        logs=logs,
        total=total,
        page=page,
        page_size=page_size,
        total_pages=total_pages
    )


@app.get("/api/stats", response_model=StorageStats)
async def get_stats():
    return storage.get_stats()


@app.get("/health")
async def health():
    return {"status": "ok"}


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", "8000"))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=False)
