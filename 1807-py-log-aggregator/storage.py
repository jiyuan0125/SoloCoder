from datetime import datetime, timedelta, timezone
from typing import Dict, List, Optional
import sys
import json
import logging
import threading

from models import LogEntry, LogLevel, ServiceStats, StorageStats

logger = logging.getLogger(__name__)


def to_utc(dt: datetime) -> datetime:
    if dt.tzinfo is None:
        return dt.replace(tzinfo=timezone.utc)
    return dt.astimezone(timezone.utc)


class LogStorage:
    def __init__(self):
        self._logs: List[LogEntry] = []
        self._lock = threading.Lock()
        self._service_index: Dict[str, List[LogEntry]] = {}
        self._time_index: Dict[str, List[LogEntry]] = {}

    def add(self, entry: LogEntry):
        if entry.timestamp.tzinfo is None:
            entry.timestamp = entry.timestamp.replace(tzinfo=timezone.utc)
        with self._lock:
            self._logs.append(entry)
            if entry.service not in self._service_index:
                self._service_index[entry.service] = []
            self._service_index[entry.service].append(entry)
            hour_key = entry.timestamp.strftime("%Y%m%d%H")
            if hour_key not in self._time_index:
                self._time_index[hour_key] = []
            self._time_index[hour_key].append(entry)

    def query(self,
              service: Optional[str] = None,
              level: Optional[LogLevel] = None,
              start_time: Optional[datetime] = None,
              end_time: Optional[datetime] = None,
              page: int = 1,
              page_size: int = 50) -> tuple[List[LogEntry], int]:
        with self._lock:
            candidates = list(self._logs)

        if service:
            candidates = [e for e in candidates if e.service == service]
        if level:
            candidates = [e for e in candidates if e.level == level]
        if start_time:
            start_utc = to_utc(start_time)
            candidates = [e for e in candidates if to_utc(e.timestamp) >= start_utc]
        if end_time:
            end_utc = to_utc(end_time)
            candidates = [e for e in candidates if to_utc(e.timestamp) <= end_utc]

        candidates.sort(key=lambda e: to_utc(e.timestamp), reverse=True)

        total = len(candidates)
        start_idx = (page - 1) * page_size
        end_idx = start_idx + page_size
        return candidates[start_idx:end_idx], total

    def cleanup_expired(self, retention_hours: int = 24) -> int:
        cutoff = datetime.now(timezone.utc) - timedelta(hours=retention_hours)
        with self._lock:
            initial_count = len(self._logs)
            self._logs = [e for e in self._logs if to_utc(e.timestamp) >= cutoff]
            self._service_index = {
                svc: [e for e in entries if to_utc(e.timestamp) >= cutoff]
                for svc, entries in self._service_index.items()
            }
            expired_hour_keys = []
            for hour_key, entries in self._time_index.items():
                self._time_index[hour_key] = [e for e in entries if to_utc(e.timestamp) >= cutoff]
                if not self._time_index[hour_key]:
                    expired_hour_keys.append(hour_key)
            for key in expired_hour_keys:
                del self._time_index[key]

            deleted_count = initial_count - len(self._logs)
            if deleted_count > 0:
                logger.warning(f"清理过期日志: 删除了 {deleted_count} 条")
            return deleted_count

    def get_stats(self) -> StorageStats:
        with self._lock:
            total_memory = self._estimate_object_size(self._logs)
            services_stats = []

            for service, entries in self._service_index.items():
                if not entries:
                    continue
                timestamps = [to_utc(e.timestamp) for e in entries]
                service_stats = ServiceStats(
                    service=service,
                    count=len(entries),
                    earliest=min(timestamps),
                    latest=max(timestamps),
                    estimated_memory_bytes=self._estimate_object_size(entries)
                )
                services_stats.append(service_stats)

            services_stats.sort(key=lambda s: s.service)

            return StorageStats(
                total_logs=len(self._logs),
                total_estimated_memory_bytes=total_memory,
                services=services_stats
            )

    def _estimate_object_size(self, obj) -> int:
        try:
            return len(json.dumps([e.model_dump() for e in obj], default=str))
        except Exception:
            return sys.getsizeof(obj)


storage = LogStorage()
