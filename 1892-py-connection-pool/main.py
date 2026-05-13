import asyncio
import logging
import signal
import sys
from typing import Optional

from aiohttp import web

from app.config import AppConfig
from app.manager import DatasourceManager
from app.api import create_app

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)
logger = logging.getLogger(__name__)


def main() -> None:
    app_config = AppConfig()
    manager = DatasourceManager()
    http_app = create_app(manager)

    port = app_config.port
    runner = web.AppRunner(http_app)

    async def start() -> None:
        await runner.setup()
        site = web.TCPSite(runner, "0.0.0.0", port)
        await site.start()
        logger.info("Server started on http://0.0.0.0:%d", port)

    async def shutdown() -> None:
        logger.info("Shutting down...")
        await manager.close_all()
        await runner.cleanup()
        logger.info("Shutdown complete")

    loop = asyncio.new_event_loop()
    asyncio.set_event_loop(loop)

    stop_event = asyncio.Event()

    def handle_signal(signal_num, frame):
        logger.info("Received signal %s", signal_num)
        loop.call_soon_threadsafe(stop_event.set)

    signal.signal(signal.SIGINT, handle_signal)
    signal.signal(signal.SIGTERM, handle_signal)

    try:
        loop.run_until_complete(start())
        loop.run_until_complete(stop_event.wait())
        loop.run_until_complete(shutdown())
    finally:
        loop.close()


if __name__ == "__main__":
    main()
