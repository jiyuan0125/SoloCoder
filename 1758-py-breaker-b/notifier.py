import asyncio
import json
import logging
from typing import List, Optional

import aiohttp

from circuit_breaker import StateChangeRecord


logger = logging.getLogger(__name__)


class Notifier:
    def __init__(self, webhook_url: Optional[str] = None):
        self.webhook_url = webhook_url
        self._session: Optional[aiohttp.ClientSession] = None

    async def _get_session(self) -> aiohttp.ClientSession:
        if self._session is None or self._session.closed:
            self._session = aiohttp.ClientSession()
        return self._session

    async def notify(self, record: StateChangeRecord):
        message = self._format_message(record)
        logger.info(f"[熔断器通知] {message}")
        await self._send_to_webhook(record, message)

    def _format_message(self, record: StateChangeRecord) -> str:
        return (
            f"服务 [{record.service}] 状态变更: "
            f"{record.from_state.value} -> {record.to_state.value} | "
            f"原因: {record.reason}"
        )

    async def _send_to_webhook(self, record: StateChangeRecord, message: str):
        if not self.webhook_url:
            return

        try:
            session = await self._get_session()
            payload = {
                "service": record.service,
                "from_state": record.from_state.value,
                "to_state": record.to_state.value,
                "reason": record.reason,
                "timestamp": record.timestamp,
                "message": message,
            }
            async with session.post(
                self.webhook_url,
                json=payload,
                timeout=aiohttp.ClientTimeout(total=5),
            ) as response:
                if response.status >= 400:
                    text = await response.text()
                    logger.warning(f"Webhook 通知失败: {response.status} - {text}")
        except Exception as e:
            logger.warning(f"Webhook 通知异常: {e}")

    async def close(self):
        if self._session and not self._session.closed:
            await self._session.close()
