from datetime import datetime, timedelta
from typing import Optional, Any
from threading import Lock
import hashlib
import json

from decimal import Decimal


CACHE_DURATION_SECONDS: int = 24 * 60 * 60


class DecimalEncoder(json.JSONEncoder):
    def default(self, o: Any) -> Any:
        if isinstance(o, Decimal):
            return str(o)
        return super().default(o)


def _normalize_for_hash(obj: Any) -> Any:
    if isinstance(obj, dict):
        return {k: _normalize_for_hash(v) for k, v in sorted(obj.items())}
    if isinstance(obj, list):
        return [_normalize_for_hash(item) for item in obj]
    if isinstance(obj, Decimal):
        return str(obj)
    return obj


class CacheItem:
    def __init__(self, key: str, value: Any, expires_at: datetime) -> None:
        self.key: str = key
        self.value: Any = value
        self.expires_at: datetime = expires_at


class FreightCache:
    def __init__(self) -> None:
        self._cache: dict[str, CacheItem] = {}
        self._lock: Lock = Lock()

    def _generate_key(self, data: dict[str, Any]) -> str:
        normalized = _normalize_for_hash(data)
        json_str = json.dumps(normalized, sort_keys=True, cls=DecimalEncoder)
        return hashlib.sha256(json_str.encode("utf-8")).hexdigest()

    def get(self, key_data: dict[str, Any]) -> Optional[Any]:
        key = self._generate_key(key_data)
        with self._lock:
            item = self._cache.get(key)
            if item is None:
                return None
            if datetime.now() > item.expires_at:
                del self._cache[key]
                return None
            return item.value

    def set(self, key_data: dict[str, Any], value: Any) -> str:
        key = self._generate_key(key_data)
        expires_at = datetime.now() + timedelta(seconds=CACHE_DURATION_SECONDS)
        with self._lock:
            self._cache[key] = CacheItem(key, value, expires_at)
        return key

    def cleanup_expired(self) -> int:
        removed_count = 0
        now = datetime.now()
        keys_to_remove: list[str] = []
        
        with self._lock:
            for key, item in self._cache.items():
                if now > item.expires_at:
                    keys_to_remove.append(key)
            
            for key in keys_to_remove:
                del self._cache[key]
                removed_count += 1
        
        return removed_count

    def clear_all(self) -> int:
        with self._lock:
            count = len(self._cache)
            self._cache.clear()
        return count

    def get_stats(self) -> dict[str, Any]:
        with self._lock:
            return {
                "total_items": len(self._cache),
            }


freight_cache: FreightCache = FreightCache()
