import logging
from datetime import datetime, timedelta
from typing import Optional

from apscheduler.schedulers.asyncio import AsyncIOScheduler
from sqlalchemy import select, func, delete

from app.config import settings
from app.database import SessionLocal, engine
from app.models import LogEntry, TraceMetadata

logger = logging.getLogger("log_cleanup")


class LogCleanupService:
    def __init__(self):
        self.scheduler: Optional[AsyncIOScheduler] = None

    async def run_cleanup(self) -> dict:
        retention_date = datetime.utcnow() - timedelta(
            days=settings.LOG_RETENTION_DAYS
        )
        result = {
            "logs_deleted": 0,
            "traces_deleted": 0,
            "cutoff": retention_date.isoformat(),
        }

        async with SessionLocal() as session:
            async with session.begin():
                count_stmt = (
                    select(func.count(LogEntry.id))
                    .where(LogEntry.timestamp < retention_date)
                )
                count_result = await session.execute(count_stmt)
                logs_to_delete = count_result.scalar() or 0

                delete_stmt = delete(LogEntry).where(
                    LogEntry.timestamp < retention_date
                )
                delete_result = await session.execute(delete_stmt)
                result["logs_deleted"] = delete_result.rowcount or logs_to_delete

                trace_stmt = delete(TraceMetadata).where(
                    TraceMetadata.last_seen < retention_date
                )
                trace_result = await session.execute(trace_stmt)
                result["traces_deleted"] = trace_result.rowcount or 0

        logger.info(
            f"Log cleanup completed: {result['logs_deleted']} logs, "
            f"{result['traces_deleted']} traces deleted. "
            f"Cutoff: {retention_date.isoformat()}"
        )
        return result

    async def manual_cleanup(self, older_than_days: int) -> dict:
        original_retention = settings.LOG_RETENTION_DAYS
        settings.LOG_RETENTION_DAYS = older_than_days
        try:
            return await self.run_cleanup()
        finally:
            settings.LOG_RETENTION_DAYS = original_retention

    def start(self):
        if self.scheduler and self.scheduler.running:
            return

        self.scheduler = AsyncIOScheduler()
        self.scheduler.add_job(
            self.run_cleanup,
            "interval",
            hours=settings.CLEANUP_INTERVAL_HOURS,
            id="log_cleanup",
            replace_existing=True,
            misfire_grace_time=300,
        )
        self.scheduler.start()
        logger.info(
            f"Log cleanup scheduler started. Interval: "
            f"{settings.CLEANUP_INTERVAL_HOURS}h, Retention: "
            f"{settings.LOG_RETENTION_DAYS}d"
        )

    def stop(self):
        if self.scheduler and self.scheduler.running:
            self.scheduler.shutdown(wait=False)
            logger.info("Log cleanup scheduler stopped")


cleanup_service = LogCleanupService()
