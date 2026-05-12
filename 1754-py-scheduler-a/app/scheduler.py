import time
import threading
import requests
from datetime import datetime, timedelta, timezone
from typing import Optional
from croniter import croniter
from .models import Task, TaskStatus, ExecutionRecord, ExecutionStatus, TaskType
from .storage import storage


CALLBACK_TIMEOUT = 30
MAX_RETRIES = 3
RETRY_INTERVALS = [5, 15, 30]


def to_naive_utc(dt: Optional[datetime]) -> Optional[datetime]:
    if dt is None:
        return None
    if dt.tzinfo is not None:
        return dt.astimezone(timezone.utc).replace(tzinfo=None)
    return dt


def now_utc() -> datetime:
    return datetime.utcnow()


def calculate_next_run(task: Task, from_time: Optional[datetime] = None) -> Optional[datetime]:
    if task.task_type == TaskType.DELAYED:
        return to_naive_utc(task.execute_at)

    if task.task_type == TaskType.CRON and task.cron_expression:
        try:
            base_time = from_time or now_utc()
            if not croniter.is_valid(task.cron_expression):
                return None
            cron = croniter(task.cron_expression, base_time)
            next_dt = cron.get_next(datetime)
            return next_dt
        except Exception:
            return None
    return None


def execute_task(task: Task) -> ExecutionRecord:
    record = ExecutionRecord(
        task_id=task.id,
        task_name=task.name,
        status=ExecutionStatus.PENDING,
        retry_attempt=task.retry_count,
    )

    if not task.callback_url:
        record.status = ExecutionStatus.SUCCESS
        record.finished_at = now_utc()
        storage.add_execution_record(record)
        return record

    try:
        response = requests.post(
            task.callback_url,
            json=task.payload or {},
            timeout=CALLBACK_TIMEOUT,
        )
        record.response_status_code = response.status_code
        if 200 <= response.status_code < 300:
            record.status = ExecutionStatus.SUCCESS
        else:
            record.status = ExecutionStatus.FAILED
            record.error_message = f"HTTP {response.status_code}: {response.text[:200]}"
    except requests.exceptions.Timeout:
        record.status = ExecutionStatus.TIMEOUT
        record.error_message = "Callback request timed out after 30 seconds"
    except Exception as e:
        record.status = ExecutionStatus.FAILED
        record.error_message = str(e)[:200]
    finally:
        record.finished_at = now_utc()
        storage.add_execution_record(record)

    return record


def handle_task_result(task: Task, record: ExecutionRecord) -> None:
    task.last_run_at = record.started_at

    if record.status == ExecutionStatus.SUCCESS:
        task.retry_count = 0
        if task.task_type == TaskType.DELAYED:
            task.status = TaskStatus.SUCCESS
            task.next_run_at = None
        else:
            task.status = TaskStatus.PENDING
            task.next_run_at = calculate_next_run(task)
    else:
        if task.retry_count < MAX_RETRIES:
            task.retry_count += 1
            interval_idx = task.retry_count - 1
            interval = RETRY_INTERVALS[interval_idx] if interval_idx < len(RETRY_INTERVALS) else 60
            task.next_run_at = now_utc() + timedelta(seconds=interval)
            task.status = TaskStatus.PENDING
        else:
            task.status = TaskStatus.DEAD_LETTER
            task.next_run_at = None

    storage.update_task(task)


def run_task(task: Task) -> None:
    current_status = task.status
    task.status = TaskStatus.RUNNING
    storage.update_task(task)

    try:
        record = execute_task(task)
        handle_task_result(task, record)
    except Exception:
        task.status = current_status
        storage.update_task(task)


def scheduler_loop() -> None:
    while True:
        now = now_utc()
        due_tasks = storage.get_due_tasks(now)

        for task in due_tasks:
            worker = threading.Thread(target=run_task, args=(task,), daemon=True)
            worker.start()

        time.sleep(1)


_scheduler_thread: Optional[threading.Thread] = None


def start_scheduler() -> None:
    global _scheduler_thread
    if _scheduler_thread and _scheduler_thread.is_alive():
        return
    _scheduler_thread = threading.Thread(target=scheduler_loop, daemon=True)
    _scheduler_thread.start()
