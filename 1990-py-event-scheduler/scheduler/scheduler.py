import asyncio
import logging
from datetime import datetime, timedelta
from typing import Dict, List, Optional

import aiohttp
from croniter import croniter
from sqlalchemy.orm import Session

from .models import (
    ExecutionHistory,
    SessionLocal,
    Task,
    TaskStatus,
    TaskType,
)

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)
logger = logging.getLogger(__name__)

MAX_HISTORY = 1000


class TaskScheduler:
    def __init__(self):
        self.running = False
        self._scheduler_task: Optional[asyncio.Task] = None
        self._running_tasks: Dict[int, List[asyncio.Task]] = {}
        self._task_queues: Dict[int, List[asyncio.Event]] = {}

    def start(self):
        self.running = True
        self._scheduler_task = asyncio.create_task(self._scheduler_loop())
        logger.info("调度器已启动")

    def stop(self):
        self.running = False
        if self._scheduler_task:
            self._scheduler_task.cancel()
        for task_list in self._running_tasks.values():
            for task in task_list:
                task.cancel()
        logger.info("调度器已停止")

    async def _scheduler_loop(self):
        while self.running:
            try:
                await self._check_and_run_due_tasks()
                await asyncio.sleep(1)
            except asyncio.CancelledError:
                break
            except Exception as e:
                logger.error(f"调度循环异常: {e}")
                await asyncio.sleep(5)

    async def _check_and_run_due_tasks(self):
        db = SessionLocal()
        try:
            now = datetime.utcnow()
            due_tasks = (
                db.query(Task)
                .filter(Task.next_run_at <= now, Task.status != TaskStatus.RUNNING)
                .all()
            )

            for task in due_tasks:
                await self._enqueue_or_run_task(task, db)
        finally:
            db.close()

    async def _enqueue_or_run_task(self, task: Task, db: Session):
        if task.id not in self._running_tasks:
            self._running_tasks[task.id] = []
            self._task_queues[task.id] = []

        running_count = len(self._running_tasks.get(task.id, []))
        queue_count = len(self._task_queues.get(task.id, []))

        if running_count > 0:
            if queue_count < task.max_queue_size:
                event = asyncio.Event()
                self._task_queues[task.id].append(event)
                logger.warning(f"任务 {task.id} ({task.name}) 已排队，队列长度: {queue_count + 1}")

                async def wait_and_run():
                    await event.wait()
                    await self._execute_task(task.id, 0)

                asyncio.create_task(wait_and_run())
            else:
                logger.warning(f"任务 {task.id} ({task.name}) 队列已满，跳过本次执行")
                self._update_next_run(task, db)
                return
        else:
            asyncio.create_task(self._execute_task(task.id, 0))

        self._update_next_run(task, db)

    def _update_next_run(self, task: Task, db: Session):
        if task.task_type == TaskType.CRON:
            if task.cron_expression:
                now = datetime.utcnow()
                cron = croniter(task.cron_expression, now)
                task.next_run_at = cron.get_next(datetime)
        elif task.task_type == TaskType.DELAYED:
            task.next_run_at = None
            task.status = TaskStatus.SUCCESS

        db.commit()

    async def _execute_task(self, task_id: int, retry_attempt: int):
        db = SessionLocal()
        try:
            task = db.query(Task).filter(Task.id == task_id).first()
            if not task:
                return

            task.status = TaskStatus.RUNNING
            task.retry_count = retry_attempt
            db.commit()

            if task.id not in self._running_tasks:
                self._running_tasks[task.id] = []

            current_task = asyncio.current_task()
            if current_task:
                self._running_tasks[task.id].append(current_task)

            history = ExecutionHistory(
                task_id=task.id,
                status=TaskStatus.RUNNING,
                start_time=datetime.utcnow(),
                retry_attempt=retry_attempt,
            )
            db.add(history)
            db.commit()

            try:
                start_time = datetime.utcnow()
                async with aiohttp.ClientSession() as session:
                    async with session.post(
                        task.callback_url,
                        json=task.callback_payload,
                        timeout=aiohttp.ClientTimeout(total=task.timeout_seconds),
                    ) as response:
                        end_time = datetime.utcnow()
                        duration = int((end_time - start_time).total_seconds())

                        history.end_time = end_time
                        history.duration_seconds = duration
                        history.response_status = response.status

                        if 200 <= response.status < 300:
                            task.status = TaskStatus.SUCCESS
                            task.last_run_at = end_time
                            history.status = TaskStatus.SUCCESS
                            logger.info(f"任务 {task.id} ({task.name}) 执行成功，耗时: {duration}s")
                        else:
                            response_text = await response.text()
                            raise Exception(f"HTTP 状态码 {response.status}: {response_text[:500]}")

            except asyncio.TimeoutError:
                error_msg = f"任务超时（{task.timeout_seconds}秒）"
                history.error_message = error_msg
                history.end_time = datetime.utcnow()
                history.duration_seconds = task.timeout_seconds
                raise Exception(error_msg)

            except aiohttp.ClientError as e:
                error_msg = f"回调 URL 不可达: {str(e)}"
                history.error_message = error_msg
                history.end_time = datetime.utcnow()
                raise Exception(error_msg)

            except Exception as e:
                if history.error_message is None:
                    history.error_message = str(e)
                if history.end_time is None:
                    history.end_time = datetime.utcnow()
                raise

            finally:
                db.commit()

        except Exception as e:
            logger.error(f"任务 {task_id} 执行失败: {e}")
            task = db.query(Task).filter(Task.id == task_id).first()
            if task:
                history.status = TaskStatus.FAILED
                task.status = TaskStatus.FAILED
                db.commit()

                if retry_attempt < task.max_retries:
                    backoff = self._get_backoff_time(task, retry_attempt)
                    logger.info(f"任务 {task_id} 将在 {backoff} 秒后重试（第 {retry_attempt + 1} 次）")
                    asyncio.get_event_loop().call_later(
                        backoff,
                        lambda: asyncio.create_task(self._execute_task(task_id, retry_attempt + 1)),
                    )

        finally:
            task = db.query(Task).filter(Task.id == task_id).first()
            if task and current_task in self._running_tasks.get(task.id, []):
                self._running_tasks[task.id].remove(current_task)

            db.commit()
            db.close()

            await self._process_queue(task_id)
            self._cleanup_old_history()

    def _get_backoff_time(self, task: Task, retry_attempt: int) -> int:
        backoff = getattr(task, "retry_backoff", [10, 30, 60])
        if retry_attempt < len(backoff):
            return backoff[retry_attempt]
        return backoff[-1] if backoff else 60

    async def _process_queue(self, task_id: int):
        if task_id in self._task_queues and self._task_queues[task_id]:
            event = self._task_queues[task_id].pop(0)
            event.set()

    def _cleanup_old_history(self):
        db = SessionLocal()
        try:
            count = db.query(ExecutionHistory).count()
            if count > MAX_HISTORY:
                excess = count - MAX_HISTORY
                old_records = (
                    db.query(ExecutionHistory)
                    .order_by(ExecutionHistory.id.asc())
                    .limit(excess)
                    .all()
                )
                for record in old_records:
                    db.delete(record)
                db.commit()
                logger.info(f"清理了 {excess} 条历史记录")
        finally:
            db.close()

    def calculate_next_run(self, task: Task) -> Optional[datetime]:
        if task.task_type == TaskType.CRON and task.cron_expression:
            now = datetime.utcnow()
            cron = croniter(task.cron_expression, now)
            return cron.get_next(datetime)
        elif task.task_type == TaskType.DELAYED and task.delay_seconds:
            return datetime.utcnow() + timedelta(seconds=task.delay_seconds)
        return None

    def restore_on_startup(self):
        db = SessionLocal()
        try:
            logger.info("正在恢复待调度任务...")
            now = datetime.utcnow()

            tasks = db.query(Task).filter(Task.next_run_at.isnot(None)).all()

            for task in tasks:
                if task.task_type == TaskType.CRON:
                    if task.cron_expression:
                        if task.next_run_at and task.next_run_at < now:
                            logger.info(f"任务 {task.id} ({task.name}) 错过了上次执行，立即触发")
                            task.next_run_at = now
                        else:
                            cron = croniter(task.cron_expression, now)
                            task.next_run_at = cron.get_next(datetime)

                elif task.task_type == TaskType.DELAYED:
                    if task.next_run_at and task.next_run_at < now:
                        logger.info(f"延迟任务 {task.id} ({task.name}) 错过执行时间，立即触发")
                        task.next_run_at = now

            db.commit()
            logger.info(f"已恢复 {len(tasks)} 个待调度任务")
        finally:
            db.close()


scheduler = TaskScheduler()
