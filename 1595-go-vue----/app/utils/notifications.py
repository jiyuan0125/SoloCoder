import logging
from datetime import datetime

logger = logging.getLogger(__name__)


def send_notification(phone: str, message: str) -> bool:
    logger.info(f"Sending notification to {phone}: {message}")
    return True


def write_system_log(message: str) -> None:
    logger.info(f"SYSTEM LOG: {message}")
