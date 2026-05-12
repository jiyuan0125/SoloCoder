import time
import threading
import requests
from datetime import datetime, timedelta
from typing import Optional
from crontab import CronTab
from .models import Task, TaskStatus, ExecutionRecord, ExecutionStatus, TaskType
from .storage import storage


CALLBACK_TIMEOUT = 30
MAX_RETRIES = 3
RETRY_INTERVALS = [5, 15, 30]


def calculate_next_run(task: Task, from_time: Optional[datetime] = None) -> Optional[datetime]:
    if task.task_type == TaskType.DELAYED:
        return task.execute_at

    if task.task_type == TaskType.CRON and task.cron_expression:
        try:
            cron = CronTab(task.cron_expression)
            base_time = from_time or datetime.utcnow()
            next_ts = cron.next(base_time, default_utc=True)
            if next_ts is None or next_ts <= 0:
                next_ts = cron.next(datetime.utcnow() + timedelta(seconds=1), default_utc=True)
            if next_ts is not None:
                return datetime.utcnow() + timedelta(seconds=next_ts)
        except Exception:
            pass
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
        record.finished_at = datetime.utcnow()
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
        record.finished_at = datetime.utcnow()
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
            interval = RETRY_INTERVALS[task.retry_count - 1] if task.retry_count <= len(RETRY_INTERVALS) else 60
            task.next_run_at = datetime.utcnow() + timedelta(seconds=interval)
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
    except Exception as e:
        task.status = current_status
        storage.update_task(task)


def scheduler_loop() -> None:
    while True:
        now = datetime.utcnow()
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
