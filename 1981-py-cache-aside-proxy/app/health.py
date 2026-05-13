import logging
from dataclasses import dataclass
from datetime import datetime
from typing import Optional
import threading

from app.config import config

logger = logging.getLogger(__name__)


@dataclass
class BackendHealth:
    is_healthy: bool = True
    consecutive_failures: int = 0
    last_failure_time: Optional[datetime] = None
    last_success_time: Optional[datetime] = None
    total_failures: int = 0
    total_requests: int = 0


class HealthManager:
    def __init__(self, max_consecutive_failures: int = 10):
        self._health = BackendHealth()
        self._max_consecutive_failures = max_consecutive_failures
        self._lock = threading.Lock()
    
    def record_success(self):
        with self._lock:
            self._health.is_healthy = True
            self._health.consecutive_failures = 0
            self._health.last_success_time = datetime.now()
            self._health.total_requests += 1
    
    def record_failure(self, error_message: str = ""):
        with self._lock:
            self._health.consecutive_failures += 1
            self._health.last_failure_time = datetime.now()
            self._health.total_failures += 1
            self._health.total_requests += 1
            
            if self._health.consecutive_failures >= self._max_consecutive_failures:
                self._health.is_healthy = False
                logger.error(
                    f"Backend marked as unhealthy! "
                    f"Consecutive failures: {self._health.consecutive_failures}, "
                    f"Error: {error_message}"
                )
    
    def is_healthy(self) -> bool:
        return self._health.is_healthy
    
    def get_status(self) -> dict:
        with self._lock:
            health = self._health
            return {
                "backend_url": config.backend_url,
                "is_healthy": health.is_healthy,
                "consecutive_failures": health.consecutive_failures,
                "max_consecutive_failures": self._max_consecutive_failures,
                "last_failure_time": health.last_failure_time.isoformat() if health.last_failure_time else None,
                "last_success_time": health.last_success_time.isoformat() if health.last_success_time else None,
                "total_failures": health.total_failures,
                "total_requests": health.total_requests
            }
    
    def reset(self):
        with self._lock:
            self._health = BackendHealth()


health_manager = HealthManager(max_consecutive_failures=config.max_consecutive_failures)
