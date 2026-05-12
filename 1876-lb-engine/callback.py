import asyncio
import httpx
from typing import Optional
from models import Node, NodeStatus
from store import store


class CallbackNotifier:
    def __init__(self):
        pass

    async def _notify_callback(self, callback_url: str, node: Node, old_status: NodeStatus):
        try:
            async with httpx.AsyncClient(timeout=5.0) as client:
                payload = {
                    "node_id": node.id,
                    "address": node.address,
                    "old_status": old_status.value,
                    "new_status": node.status.value,
                    "timestamp": node.updated_at.isoformat()
                }
                await client.post(callback_url, json=payload)
        except Exception:
            pass

    def notify_status_change(self, node: Node, old_status: NodeStatus):
        callbacks = store.get_all_callbacks()
        for callback in callbacks:
            asyncio.create_task(self._notify_callback(callback.url, node, old_status))


notifier = CallbackNotifier()
