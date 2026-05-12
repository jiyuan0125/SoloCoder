from datetime import datetime
from typing import Optional

from .models import CheckRecord
from .store import store


class HistoryService:
    def add_record(
        self,
        target_id: str,
        success: bool,
        duration_ms: float,
        error_message: Optional[str] = None,
    ) -> CheckRecord:
        record = CheckRecord(
            target_id=target_id,
            success=success,
            duration_ms=duration_ms,
            error_message=error_message,
            checked_at=datetime.utcnow(),
        )
        return store.add_check_record(record)

    def get_history(
        self,
        target_id: Optional[str] = None,
        limit: int = 100,
    ) -> list[CheckRecord]:
        return store.get_history(target_id=target_id, limit=limit)


history_service = HistoryService()
