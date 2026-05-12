import asyncio
import os
from datetime import datetime

from aiohttp import web

from .group import group_manager
from .models import Target
from .notify import notification_service
from .probe import probe_executor
from .routes import setup_routes
from .store import store


class Scheduler:
    def __init__(self):
        self._running = False
        self._tasks: dict[str, asyncio.Task] = {}
        self._target_last_run: dict[str, datetime] = {}

    async def start(self):
        self._running = True
        asyncio.create_task(self._monitor_loop())

    async def stop(self):
        self._running = False
        for task in self._tasks.values():
            task.cancel()
        self._tasks.clear()
        await probe_executor.close()
        await notification_service.close()

    async def _monitor_loop(self):
        while self._running:
            await self._check_and_spawn()
            await asyncio.sleep(1)

    async def _check_and_spawn(self):
        targets = store.get_all_targets()
        now = datetime.utcnow()

        for target in targets:
            if target.id in self._tasks and not self._tasks[target.id].done():
                continue

            last_run = self._target_last_run.get(target.id)
            if last_run:
                elapsed = (now - last_run).total_seconds()
                if elapsed < target.interval:
                    continue

            self._tasks[target.id] = asyncio.create_task(self._run_check(target))

    async def _run_check(self, target: Target):
        self._target_last_run[target.id] = datetime.utcnow()

        try:
            record, event = await probe_executor.execute(target)

            if event:
                await notification_service.notify_status_change(event)
                group_manager.recompute_all_affected_groups(target.group_id)
        except Exception:
            pass


async def on_startup(app: web.Application):
    app["scheduler"] = Scheduler()
    await app["scheduler"].start()


async def on_shutdown(app: web.Application):
    if "scheduler" in app:
        await app["scheduler"].stop()


def create_app() -> web.Application:
    app = web.Application()
    app.on_startup.append(on_startup)
    app.on_shutdown.append(on_shutdown)
    setup_routes(app)
    return app


def main():
    app = create_app()
    port = int(os.environ.get("PORT", 8080))
    web.run_app(app, port=port)


if __name__ == "__main__":
    main()
