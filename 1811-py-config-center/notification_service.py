import asyncio
import logging
from datetime import datetime
from typing import List
import httpx
from sqlalchemy.orm import Session

from models import Notification, Subscriber

logger = logging.getLogger(__name__)

RETRY_INTERVALS = [1, 5, 30]  # 秒


async def send_notification(callback_url: str, payload: dict) -> bool:
    try:
        async with httpx.AsyncClient(timeout=5.0) as client:
            response = await client.post(callback_url, json=payload)
            return 200 <= response.status_code < 300
    except Exception as e:
        logger.warning(f"通知失败: {callback_url}, 错误: {e}")
        return False


async def notify_subscribers_with_retry(
    db: Session,
    config_id: int,
    app_name: str,
    key: str,
    value: str,
    version: int
):
    payload = {
        "app_name": app_name,
        "key": key,
        "value": value,
        "version": version
    }

    subscribers = db.query(Subscriber).filter(
        Subscriber.app_name == app_name
    ).all()

    if not subscribers:
        return

    for subscriber in subscribers:
        notification = Notification(
            config_id=config_id,
            app_name=app_name,
            key=key,
            callback_url=subscriber.callback_url,
            version=version,
            status="pending",
            retry_count=0
        )
        db.add(notification)
        db.commit()

        asyncio.create_task(
            process_notification(
                notification.id,
                subscriber.callback_url,
                payload
            )
        )


async def process_notification(
    notification_id: int,
    callback_url: str,
    payload: dict
):
    from database import SessionLocal

    db = SessionLocal()
    try:
        notification = db.query(Notification).filter(
            Notification.id == notification_id
        ).first()

        if not notification:
            return

        success = False
        for attempt in range(len(RETRY_INTERVALS) + 1):
            success = await send_notification(callback_url, payload)
            notification.retry_count = attempt + 1

            if success:
                notification.status = "success"
                notification.completed_at = datetime.utcnow()
                db.commit()
                logger.info(
                    f"通知成功: {callback_url}, key={payload['key']}, "
                    f"version={payload['version']}, 尝试次数: {attempt + 1}"
                )
                break

            if attempt < len(RETRY_INTERVALS):
                wait_time = RETRY_INTERVALS[attempt]
                logger.info(
                    f"通知尝试 {attempt + 1} 失败，等待 {wait_time} 秒后重试: "
                    f"{callback_url}"
                )
                await asyncio.sleep(wait_time)
            else:
                notification.status = "failed"
                notification.completed_at = datetime.utcnow()
                db.commit()
                logger.error(
                    f"通知全部失败: {callback_url}, key={payload['key']}, "
                    f"version={payload['version']}"
                )
    except Exception as e:
        logger.error(f"处理通知时出错: {e}")
    finally:
        db.close()
