import asyncio
import threading
import time
from typing import Dict, Optional, List
from .models import (
    DatasourceType,
    DatasourceCreate,
    DatasourceStats,
    DatasourceOverview,
)
from .pool import ConnectionPool


class DatasourceManager:
    def __init__(self):
        self._pools: Dict[str, ConnectionPool] = {}
        self._lock = threading.RLock()
        self._background_task = None
        self._loop: Optional[asyncio.AbstractEventLoop] = None

    def register_datasource(self, ds: DatasourceCreate) -> bool:
        with self._lock:
            if ds.name in self._pools:
                return False
            pool = ConnectionPool(
                name=ds.name,
                type=ds.type,
                max_connections=ds.max_connections,
                idle_timeout=ds.idle_timeout,
            )
            self._pools[ds.name] = pool
            return True

    def unregister_datasource(self, name: str) -> bool:
        with self._lock:
            if name in self._pools:
                pool = self._pools.pop(name)
                for conn in pool.get_all_connections():
                    conn.close()
                return True
            return False

    def get_pool(self, name: str) -> Optional[ConnectionPool]:
        with self._lock:
            return self._pools.get(name)

    def get_datasource_stats(self, name: str) -> Optional[DatasourceStats]:
        pool = self.get_pool(name)
        if not pool:
            return None
        stats = pool.get_stats()
        return DatasourceStats(
            name=pool.name,
            type=pool.type,
            max_connections=pool.max_connections,
            current_pool_size=pool.current_size(),
            connections=stats,
        )

    def list_datasources(self) -> List[DatasourceOverview]:
        with self._lock:
            result = []
            for name, pool in self._pools.items():
                stats = pool.get_stats()
                result.append(
                    DatasourceOverview(
                        name=name,
                        type=pool.type,
                        max_connections=pool.max_connections,
                        current_pool_size=pool.current_size(),
                        idle_connections=stats.idle,
                        in_use_connections=stats.in_use,
                    )
                )
            return result

    def update_max_connections(self, name: str, new_max: int) -> bool:
        pool = self.get_pool(name)
        if pool:
            pool.update_max_connections(new_max)
            return True
        return False

    def _run_background_checks(self):
        while True:
            with self._lock:
                pools = list(self._pools.values())
            for pool in pools:
                try:
                    pool.validate_idle_connections()
                    pool.check_leaks()
                except Exception:
                    pass
            time.sleep(60)

    def start_background_task(self):
        if self._background_task is None:
            thread = threading.Thread(target=self._run_background_checks, daemon=True)
            thread.start()
            self._background_task = thread


_manager: Optional[DatasourceManager] = None


def get_manager() -> DatasourceManager:
    global _manager
    if _manager is None:
        _manager = DatasourceManager()
    return _manager
