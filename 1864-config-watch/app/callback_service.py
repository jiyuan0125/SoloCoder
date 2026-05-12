import asyncio
import httpx
from sqlalchemy.orm import Session
from app.models import Watch
from app.schemas import CallbackPayload

MAX_FAILURES = 5
MAX_RETRIES = 2
RETRY_INTERVAL = 5


async def notify_watches(db: Session, key: str, environment: str, value: str, version: int):
    watches = db.query(Watch).filter(
        Watch.key == key,
        Watch.environment == environment,
        Watch.is_active == "active"
    ).all()

    payload = CallbackPayload(
        key=key,
        environment=environment,
        value=value,
        version=version
    )

    tasks = [push_notification(db, watch, payload) for watch in watches]
    await asyncio.gather(*tasks)


async def push_notification(db: Session, watch: Watch, payload: CallbackPayload):
    async with httpx.AsyncClient() as client:
        for attempt in range(MAX_RETRIES + 1):
            try:
                response = await client.post(
                    watch.callback_url,
                    json=payload.model_dump(),
                    timeout=10.0
                )
                if response.status_code >= 200 and response.status_code < 300:
                    if watch.failure_count > 0:
                        watch.failure_count = 0
                        db.commit()
                    return
            except Exception:
                pass
            
            if attempt < MAX_RETRIES:
                await asyncio.sleep(RETRY_INTERVAL)
        
        watch.failure_count += 1
        if watch.failure_count >= MAX_FAILURES:
            watch.is_active = "inactive"
        db.commit()
