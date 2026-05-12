import time
from threading import Lock
from typing import Dict, List, Optional

from app.models import ServiceStats, Stats, TagStats


class StatsService:
    def __init__(self):
        self._lock = Lock()
        self._stats: Dict[str, Dict[str, Stats]] = {}

    def record_request(
        self,
        service_name: str,
        tag: str,
        latency_ms: float,
        is_error: bool = False,
    ) -> None:
        with self._lock:
            if service_name not in self._stats:
                self._stats[service_name] = {}
            if tag not in self._stats[service_name]:
                self._stats[service_name][tag] = Stats()

            stats = self._stats[service_name][tag]
            stats.total_requests += 1
            stats.total_latency_ms += latency_ms
            if is_error:
                stats.error_requests += 1

    def get_service_stats(self, service_name: str) -> Optional[ServiceStats]:
        with self._lock:
            if service_name not in self._stats:
                return None

            tags_stats = []
            for tag, stats in self._stats[service_name].items():
                tags_stats.append(
                    TagStats(
                        tag=tag,
                        stats=stats.model_copy(),
                    )
                )
            return ServiceStats(service_name=service_name, tags=tags_stats)

    def list_service_stats(self) -> List[ServiceStats]:
        with self._lock:
            result = []
            for service_name in self._stats:
                tags_stats = []
                for tag, stats in self._stats[service_name].items():
                    tags_stats.append(
                        TagStats(
                            tag=tag,
                            stats=stats.model_copy(),
                        )
                    )
                result.append(
                    ServiceStats(service_name=service_name, tags=tags_stats)
                )
            return result

    def reset_stats(self, service_name: Optional[str] = None) -> None:
        with self._lock:
            if service_name:
                if service_name in self._stats:
                    self._stats[service_name] = {}
            else:
                self._stats = {}
