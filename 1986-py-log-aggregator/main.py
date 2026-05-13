import os
import asyncio
from aiohttp import web
from log_aggregator import LogAggregator
from handlers import setup_routes


async def main():
    app = web.Application()
    log_aggregator = LogAggregator()
    app["log_aggregator"] = log_aggregator
    setup_routes(app)
    port = int(os.environ.get("PORT", 8080))
    runner = web.AppRunner(app)
    await runner.setup()
    site = web.TCPSite(runner, "0.0.0.0", port)
    await site.start()
    print(f"Log Aggregator service started on port {port}")
    while True:
        await asyncio.sleep(3600)


if __name__ == "__main__":
    asyncio.run(main())
