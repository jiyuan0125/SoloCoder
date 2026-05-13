import asyncio
import time
from typing import Dict, Optional
from dataclasses import dataclass, field

from app.config import config


@dataclass
class KeyLockState:
    lock: asyncio.Lock = field(default_factory=asyncio.Lock)
    is_loading: bool = False
    result_available: asyncio.Event = field(default_factory=asyncio.Event)


class BreakdownProtector:
    def __init__(self, timeout_ms: int = 500):
        self._key_locks: Dict[str, KeyLockState] = {}
        self._lock = asyncio.Lock()
        self._timeout = timeout_ms / 1000.0
    
    def _get_key_lock(self, key: str) -> KeyLockState:
        if key not in self._key_locks:
            self._key_locks[key] = KeyLockState()
        return self._key_locks[key]
    
    def _cleanup(self, key: str):
        if key in self._key_locks:
            del self._key_locks[key]
    
    async def acquire_for_load(self, key: str) -> bool:
        """
        Try to acquire the right to load data for this key.
        Returns True if this caller should load from backend,
        False if another caller is already loading and this one should wait.
        """
        key_lock = self._get_key_lock(key)
        
        if key_lock.is_loading:
            return False
        
        await key_lock.lock.acquire()
        
        if key_lock.is_loading:
            key_lock.lock.release()
            return False
        
        key_lock.is_loading = True
        key_lock.result_available.clear()
        key_lock.lock.release()
        
        return True
    
    async def wait_for_result(self, key: str) -> bool:
        """
        Wait for another request to load the data.
        Returns True if result became available within timeout,
        False if timeout occurred (should fallback to loading).
        """
        key_lock = self._get_key_lock(key)
        event = key_lock.result_available
        
        try:
            await asyncio.wait_for(event.wait(), timeout=self._timeout)
            return True
        except asyncio.TimeoutError:
            return False
    
    async def release_load(self, key: str, success: bool):
        """
        Notify waiting requests that loading is complete.
        """
        key_lock = self._get_key_lock(key)
        
        await key_lock.lock.acquire()
        try:
            key_lock.is_loading = False
            key_lock.result_available.set()
        finally:
            key_lock.lock.release()
        
        if not success:
            self._cleanup(key)


breakdown_protector = BreakdownProtector(timeout_ms=config.breakdown_timeout_ms)
