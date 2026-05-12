from __future__ import annotations

import httpx

from models import HealthStatus
from storage import store


class Notifier:
    @staticmethod
    async def notify_group_state_change(
        group_id, group_name: str, old_status: HealthStatus, new_status: HealthStatus
    ) -> None:
        callbacks = await store.get_all_callbacks()
        if not callbacks:
            return

        payload = {
            "group_id": str(group_id),
            "group_name": group_name,
            "old_status": old_status.value,
            "new_status": new_status.value,
        }

        for callback in callbacks:
            try:
                async with httpx.AsyncClient() as client:
                    await client.post(callback.url, json=payload, timeout=5)
            except Exception:
                pass
