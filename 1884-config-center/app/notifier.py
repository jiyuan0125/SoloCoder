import asyncio
from typing import Optional
from httpx import AsyncClient, HTTPError

MAX_RETRIES = 2
RETRY_INTERVAL = 5
MAX_FAILED_COUNT = 5


async def notify_watchers(
    watches_data: list,
    project: str,
    env: str,
    key: str,
    value: Optional[str],
    version: int
):
    if not watches_data:
        return

    payload = {
        "project": project,
        "env": env,
        "key": key,
        "value": value,
        "version": version
    }

    results = []

    async with AsyncClient(timeout=10.0) as client:
        for watch in watches_data:
            success, failed_count, status = await _push_notification(
                client, watch["id"], watch["callback_url"], watch["failed_count"], payload
            )
            results.append({
                "id": watch["id"],
                "success": success,
                "failed_count": failed_count,
                "status": status
            })

    return results


async def _push_notification(
    client: AsyncClient,
    watch_id: int,
    callback_url: str,
    current_failed_count: int,
    payload: dict
):
    attempts = 0
    success = False
    max_attempts = MAX_RETRIES + 1

    while attempts < max_attempts and not success:
        try:
            response = await client.post(callback_url, json=payload)
            if 200 <= response.status_code < 300:
                success = True
        except HTTPError:
            pass
        except Exception:
            pass

        attempts += 1
        if not success and attempts < max_attempts:
            await asyncio.sleep(RETRY_INTERVAL)

    if success:
        new_failed_count = 0
        new_status = "active"
    else:
        new_failed_count = current_failed_count + 1
        new_status = "inactive" if new_failed_count >= MAX_FAILED_COUNT else "active"

    return success, new_failed_count, new_status
