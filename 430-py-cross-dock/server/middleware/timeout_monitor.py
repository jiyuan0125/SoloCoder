import threading
import time
from datetime import datetime, timezone
from typing import Optional

from shared.models import CrossDockStatus
from server.services.cross_dock_service import CrossDockService


class TimeoutMonitor:
    def __init__(self, service: CrossDockService, check_interval: int = 60) -> None:
        self._service = service
        self._check_interval = check_interval
        self._stop_event = threading.Event()
        self._thread: Optional[threading.Thread] = None

    def start(self) -> None:
        if self._thread is not None and self._thread.is_alive():
            return
        self._stop_event.clear()
        self._thread = threading.Thread(target=self._monitor_loop, daemon=True)
        self._thread.start()

    def stop(self) -> None:
        self._stop_event.set()
        if self._thread is not None:
            self._thread.join(timeout=10)

    def _monitor_loop(self) -> None:
        while not self._stop_event.is_set():
            try:
                self._check_timeouts()
            except Exception as e:
                import logging

                logger = logging.getLogger(__name__)
                logger.error(f"超时监控检查出错: {e}")
            self._stop_event.wait(self._check_interval)

    def _check_timeouts(self) -> None:
        orders = self._service.get_all_orders()
        current_time = datetime.now(timezone.utc)

        for order in orders:
            if order.status == CrossDockStatus.INBOUND_COMPLETED:
                if self._service.check_timeout(order, current_time):
                    order.status = CrossDockStatus.TIMEOUT_ALERT
                    self._service.send_timeout_alert(order)
