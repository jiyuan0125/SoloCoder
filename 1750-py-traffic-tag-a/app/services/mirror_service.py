import asyncio
import logging
from typing import Optional

import httpx

from app.config import settings

logger = logging.getLogger(__name__)


class MirrorService:
    def __init__(self):
        self._client: Optional[httpx.AsyncClient] = None

    async def _get_client(self) -> httpx.AsyncClient:
        if self._client is None:
            self._client = httpx.AsyncClient(timeout=5.0)
        return self._client

    async def mirror_request(
        self,
        method: str,
        url: str,
        headers: dict,
        body: bytes,
        tag: str,
    ) -> None:
        if not settings.gateway_config.mirror.enabled:
            return

        if tag not in settings.gateway_config.mirror.tags:
            return

        mirror_upstream = settings.gateway_config.mirror.upstream
        if not mirror_upstream:
            return

        try:
            client = await self._get_client()
            mirror_headers = {k: v for k, v in headers.items() if k.lower() != "host"}
            mirror_headers["X-Color-Tag"] = tag
            mirror_headers["X-Is-Mirror"] = "true"

            await client.request(
                method=method,
                url=f"{mirror_upstream}{url}",
                headers=mirror_headers,
                content=body,
            )
            logger.info(f"Mirrored request to {mirror_upstream}{url} (tag={tag})")
        except Exception as e:
            logger.warning(f"Failed to mirror request: {e}")

    async def close(self) -> None:
        if self._client:
            await self._client.aclose()
            self._client = None


mirror_service = MirrorService()
