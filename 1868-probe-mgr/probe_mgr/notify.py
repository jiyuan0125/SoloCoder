import json
from typing import Optional

import aiohttp

from .models import StatusChangeEvent
from .store import store


class NotificationService:
    def __init__(self):
        self._session: Optional[aiohttp.ClientSession] = None

    async def _get_session(self) -> aiohttp.ClientSession:
        if self._session is None or self._session.closed:
            self._session = aiohttp.ClientSession()
        return self._session

    async def close(self):
        if self._session and not self._session.closed:
            await self._session.close()

    async def notify_status_change(self, event: StatusChangeEvent):
        callbacks = store.get_all_callbacks()
        if not callbacks:
            return

        payload = {
            "target_id": event.target_id,
            "old_status": event.old_status.value,
            "new_status": event.new_status.value,
            "timestamp": event.changed_at.isoformat(),
        }

        for callback in callbacks:
            await self._send_callback(callback.url, payload)

    async def _send_callback(self, url: str, payload: dict):
        try:
            session = await self._get_session()
            async with session.post(
                url,
                json=payload,
                headers={"Content-Type": "application/json"},
                timeout=aiohttp.ClientTimeout(total=10),
            ) as response:
                pass
        except Exception:
            pass


notification_service = NotificationService()
