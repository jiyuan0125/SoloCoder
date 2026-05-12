import asyncio
import time
from datetime import datetime
from typing import Optional

import aiohttp

from .models import (
    CheckRecord,
    HealthStatus,
    ProbeType,
    StatusChangeEvent,
    Target,
)
from .store import store


class ProbeExecutor:
    def __init__(self):
        self._session: Optional[aiohttp.ClientSession] = None

    async def _get_session(self) -> aiohttp.ClientSession:
        if self._session is None or self._session.closed:
            self._session = aiohttp.ClientSession()
        return self._session

    async def close(self):
        if self._session and not self._session.closed:
            await self._session.close()

    async def _check_http(self, target: Target) -> tuple[bool, Optional[str], float]:
        if not target.http_config:
            return False, "HTTP config missing", 0.0

        config = target.http_config
        start_time = time.monotonic()
        error_msg: Optional[str] = None
        success = False

        try:
            session = await self._get_session()
            async with session.get(
                config.url,
                timeout=aiohttp.ClientTimeout(total=target.timeout),
            ) as response:
                if response.status == config.expected_status:
                    success = True
                else:
                    error_msg = f"Unexpected status: {response.status}"
        except asyncio.TimeoutError:
            error_msg = "Request timed out"
        except aiohttp.ClientError as e:
            error_msg = f"Client error: {str(e)}"
        except Exception as e:
            error_msg = f"Unexpected error: {str(e)}"

        duration_ms = (time.monotonic() - start_time) * 1000
        return success, error_msg, duration_ms

    async def _check_tcp(self, target: Target) -> tuple[bool, Optional[str], float]:
        if not target.tcp_config:
            return False, "TCP config missing", 0.0

        config = target.tcp_config
        start_time = time.monotonic()
        error_msg: Optional[str] = None
        success = False

        try:
            _, writer = await asyncio.wait_for(
                asyncio.open_connection(config.host, config.port),
                timeout=target.timeout,
            )
            writer.close()
            await writer.wait_closed()
            success = True
        except asyncio.TimeoutError:
            error_msg = "Connection timed out"
        except OSError as e:
            error_msg = f"Connection error: {str(e)}"
        except Exception as e:
            error_msg = f"Unexpected error: {str(e)}"

        duration_ms = (time.monotonic() - start_time) * 1000
        return success, error_msg, duration_ms

    async def execute(self, target: Target) -> tuple[CheckRecord, Optional[StatusChangeEvent]]:
        success, error_msg, duration_ms = await self._run_check(target)
        record = self._create_record(target, success, duration_ms, error_msg)
        event = self._update_target_state(target, success)
        store.add_check_record(record)
        return record, event

    async def _run_check(self, target: Target) -> tuple[bool, Optional[str], float]:
        if target.probe_type == ProbeType.HTTP:
            return await self._check_http(target)
        elif target.probe_type == ProbeType.TCP:
            return await self._check_tcp(target)
        else:
            return False, f"Unknown probe type: {target.probe_type}", 0.0

    def _create_record(
        self,
        target: Target,
        success: bool,
        duration_ms: float,
        error_msg: Optional[str],
    ) -> CheckRecord:
        return CheckRecord(
            target_id=target.id,
            success=success,
            duration_ms=duration_ms,
            error_message=error_msg,
            checked_at=datetime.utcnow(),
        )

    def _update_target_state(
        self,
        target: Target,
        success: bool,
    ) -> Optional[StatusChangeEvent]:
        old_status = target.status

        if success:
            target.consecutive_successes += 1
            target.consecutive_failures = 0
        else:
            target.consecutive_failures += 1
            target.consecutive_successes = 0

        new_status = old_status
        if old_status == HealthStatus.UNKNOWN:
            if success:
                if target.consecutive_successes >= target.success_threshold:
                    new_status = HealthStatus.HEALTHY
            else:
                if target.consecutive_failures >= target.failure_threshold:
                    new_status = HealthStatus.UNHEALTHY
        elif old_status == HealthStatus.HEALTHY:
            if target.consecutive_failures >= target.failure_threshold:
                new_status = HealthStatus.UNHEALTHY
        elif old_status == HealthStatus.UNHEALTHY:
            if target.consecutive_successes >= target.success_threshold:
                new_status = HealthStatus.HEALTHY

        event: Optional[StatusChangeEvent] = None
        if new_status != old_status:
            target.status = new_status
            target.last_status_change_at = datetime.utcnow()
            event = StatusChangeEvent(
                target_id=target.id,
                old_status=old_status,
                new_status=new_status,
                changed_at=target.last_status_change_at,
            )

        target.last_checked_at = datetime.utcnow()
        return event


probe_executor = ProbeExecutor()
