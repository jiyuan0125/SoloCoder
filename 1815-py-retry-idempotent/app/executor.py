from __future__ import annotations

import asyncio
from datetime import datetime, timedelta
from typing import Optional

import httpx

from app.models import BackoffStrategy, Task, TaskResult, TaskStatus
from app.store import task_store


def compute_backoff_delay(retry_count: int, strategy: BackoffStrategy) -> float:
    if strategy == BackoffStrategy.EXPONENTIAL:
        base = 1.0
        delay = base * (2 ** retry_count)
        return min(delay, 60.0)
    return 1.0


async def execute_task(task: Task) -> tuple[bool, TaskResult]:
    try:
        async with httpx.AsyncClient(timeout=30.0) as client:
            resp = await client.request(
                method=task.method,
                url=task.url,
                json=task.body,
            )
            try:
                body = resp.json()
            except Exception:
                body = resp.text
            if 200 <= resp.status_code < 400:
                return True, TaskResult(status_code=resp.status_code, body=body)
            return False, TaskResult(status_code=resp.status_code, body=body)
    except Exception as exc:
        return False, TaskResult(error=str(exc))


async def run_single_task(task_id) -> None:
    task = await task_store.get_task(task_id)
    if task is None:
        return

    await task_store.update_task(
        task_id,
        status=TaskStatus.RUNNING,
        started_at=datetime.utcnow(),
    )

    success, result = await execute_task(task)
    now = datetime.utcnow()

    if success:
        await task_store.update_task(
            task_id,
            status=TaskStatus.SUCCESS,
            result=result,
            completed_at=now,
        )
        return

    current_retry = task.current_retry
    max_retries = task.retry_config.max_retries

    if current_retry >= max_retries:
        await task_store.update_task(
            task_id,
            status=TaskStatus.FAILED,
            result=result,
            completed_at=now,
        )
        return

    next_retry = current_retry + 1
    delay = compute_backoff_delay(next_retry - 1, task.retry_config.backoff_strategy)
    next_retry_at = now + timedelta(seconds=delay)

    await task_store.update_task(
        task_id,
        status=TaskStatus.RETRYING,
        current_retry=next_retry,
        next_retry_at=next_retry_at,
    )
    await task_store.add_retrying(task_id)

    asyncio.create_task(schedule_retry(task_id, delay))


async def schedule_retry(task_id, delay: float) -> None:
    await asyncio.sleep(delay)
    await task_store.remove_retrying(task_id)
    await task_store.mark_pending(task_id)


async def scheduler_worker() -> None:
    while True:
        task_id = await task_store.get_next_pending()
        if task_id is None:
            await asyncio.sleep(0.1)
            continue
        asyncio.create_task(run_single_task(task_id))


async def start_workers(worker_count: int = 4) -> list[asyncio.Task]:
    return [asyncio.create_task(scheduler_worker()) for _ in range(worker_count)]
