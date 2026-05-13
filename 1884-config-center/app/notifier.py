import asyncio
from typing import Optional
from sqlalchemy.orm import Session
from httpx import AsyncClient, HTTPError
from app.models import Watch


MAX_RETRIES = 2
RETRY_INTERVAL = 5
MAX_FAILED_COUNT = 5


async def notify_watchers(
    db: Session,
    project: str,
    env: str,
    key: str,
    value: Optional[str],
    version: int
):
    watches = db.query(Watch).filter(
        Watch.project == project,
        Watch.env == env,
        Watch.key == key,
        Watch.status == "active"
    ).all()

    if not watches:
        return

    payload = {
        "project": project,
        "env": env,
        "key": key,
        "value": value,
        "version": version
    }

    async with AsyncClient(timeout=10.0) as client:
        for watch in watches:
            asyncio.create_task(_push_notification(client, db, watch, payload))


async def _push_notification(
    client: AsyncClient,
    db: Session,
    watch: Watch,
    payload: dict
):
    attempts = 0
    success = False
    max_attempts = MAX_RETRIES + 1

    while attempts < max_attempts and not success:
        try:
            response = await client.post(watch.callback_url, json=payload)
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
        if watch.failed_count > 0:
            watch.failed_count = 0
    else:
        watch.failed_count += 1
        if watch.failed_count >= MAX_FAILED_COUNT:
            watch.status = "inactive"

    db.commit()
