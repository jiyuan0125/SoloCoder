from apscheduler.schedulers.background import BackgroundScheduler
from apscheduler.triggers.cron import CronTrigger
from datetime import datetime
import httpx
import logging
from sqlalchemy.orm import Session
from app.models.models import TaskConfig, ExecutionAudit, TaskStatus
from app.core.database import SessionLocal

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

scheduler = BackgroundScheduler()


async def execute_task(task_id: int, retry_count: int = 0):
    db: Session = SessionLocal()
    try:
        task = db.query(TaskConfig).filter(TaskConfig.id == task_id).first()
        if not task:
            logger.error(f"Task {task_id} not found")
            return

        trigger_time = datetime.now()
        audit = ExecutionAudit(
            task_id=task.id,
            task_name=task.name,
            trigger_time=trigger_time,
            callback_url=task.callback_url,
            retry_count=retry_count,
            status=TaskStatus.RUNNING,
            executor="system"
        )
        db.add(audit)
        db.commit()
        db.refresh(audit)

        start_time = datetime.now()
        audit.start_time = start_time
        db.commit()

        try:
            async with httpx.AsyncClient(timeout=task.timeout) as client:
                response = await client.post(
                    task.callback_url,
                    json={
                        "task_id": task.id,
                        "task_name": task.name,
                        "trigger_time": trigger_time.isoformat(),
                        "retry_count": retry_count
                    }
                )
                result_text = response.text[:500] if response.text else ""
                audit.callback_result = result_text

                if response.status_code >= 400:
                    raise Exception(f"HTTP {response.status_code}: {result_text}")

                audit.status = TaskStatus.SUCCESS
                logger.info(f"Task {task.name} executed successfully")

        except httpx.TimeoutException:
            audit.status = TaskStatus.TIMEOUT
            audit.error_message = "Request timeout"
            logger.error(f"Task {task.name} timeout")

            if retry_count < task.max_retry:
                scheduler.add_job(
                    execute_task,
                    'date',
                    run_date=datetime.now(),
                    args=[task_id, retry_count + 1]
                )

        except Exception as e:
            audit.status = TaskStatus.FAILED
            audit.error_message = str(e)[:500]
            logger.error(f"Task {task.name} failed: {e}")

            if retry_count < task.max_retry:
                scheduler.add_job(
                    execute_task,
                    'date',
                    run_date=datetime.now(),
                    args=[task_id, retry_count + 1]
                )

        finally:
            audit.end_time = datetime.now()
            db.commit()

    except Exception as e:
        logger.error(f"Error in execute_task: {e}")
    finally:
        db.close()


def sync_execute_task(task_id: int, retry_count: int = 0):
    import asyncio
    loop = asyncio.new_event_loop()
    asyncio.set_event_loop(loop)
    try:
        loop.run_until_complete(execute_task(task_id, retry_count))
    finally:
        loop.close()


def add_task_to_scheduler(task: TaskConfig):
    job_id = f"task_{task.id}"
    if scheduler.get_job(job_id):
        scheduler.remove_job(job_id)

    if task.is_active:
        try:
            trigger = CronTrigger.from_crontab(task.cron_expression)
            scheduler.add_job(
                sync_execute_task,
                trigger=trigger,
                id=job_id,
                args=[task.id, 0],
                replace_existing=True
            )
            logger.info(f"Task {task.name} added to scheduler with cron: {task.cron_expression}")
        except Exception as e:
            logger.error(f"Failed to add task {task.name} to scheduler: {e}")


def remove_task_from_scheduler(task_id: int):
    job_id = f"task_{task_id}"
    if scheduler.get_job(job_id):
        scheduler.remove_job(job_id)
        logger.info(f"Task {task_id} removed from scheduler")


def init_scheduler():
    db: Session = SessionLocal()
    try:
        tasks = db.query(TaskConfig).filter(TaskConfig.is_active == True).all()
        for task in tasks:
            add_task_to_scheduler(task)
        scheduler.start()
        logger.info("Scheduler started successfully")
    except Exception as e:
        logger.error(f"Failed to initialize scheduler: {e}")
    finally:
        db.close()


def shutdown_scheduler():
    if scheduler.running:
        scheduler.shutdown()
        logger.info("Scheduler shutdown")
