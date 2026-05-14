import hashlib
import time
import asyncio
from typing import Dict, Optional, Any
from collections import OrderedDict

from app.config import config


class CacheEntry:
    def __init__(self, value: Any, status_code: int, ttl: int, path: str = ""):
        self.value = value
        self.status_code = status_code
        self.ttl = ttl
        self.path = path
        self.expiry_time = time.time() + ttl
        self.access_time = time.time()


class NamespaceCache:
    def __init__(self, namespace: str):
        self.namespace = namespace
        self._data: Dict[str, CacheEntry] = {}
        self._order: OrderedDict = OrderedDict()
        self._lock = asyncio.Lock()
    
    @property
    def size(self) -> int:
        return len(self._data)
    
    def _evict_if_needed(self, max_size: int):
        while len(self._data) > max_size:
            if self._order:
                oldest_key, _ = self._order.popitem(last=False)
                if oldest_key in self._data:
                    del self._data[oldest_key]
    
    def _remove_expired(self):
        now = time.time()
        expired_keys = [key for key, entry in self._data.items() if now > entry.expiry_time]
        for key in expired_keys:
            if key in self._data:
                del self._data[key]
            if key in self._order:
                del self._order[key]
    
    async def get(self, key: str) -> Optional[CacheEntry]:
        self._remove_expired()
        
        if key in self._data:
            entry = self._data[key]
            entry.access_time = time.time()
            
            if key in self._order:
                self._order.move_to_end(key)
            
            return entry
        return None
    
    async def set(self, key: str, value: Any, status_code: int, ttl: int, path: str = ""):
        ns_config = config.get_namespace_config(self.namespace)
        
        self._remove_expired()
        
        self._data[key] = CacheEntry(value=value, status_code=status_code, ttl=ttl, path=path)
        self._order[key] = time.time()
        self._order.move_to_end(key)
        
        self._evict_if_needed(ns_config.max_size)
    
    async def delete(self, key: str):
        if key in self._data:
            del self._data[key]
        if key in self._order:
            del self._order[key]
    
    async def invalidate_by_prefix(self, prefix: str):
        keys_to_delete = []
        for key, entry in self._data.items():
            if entry.path.startswith(prefix):
                keys_to_delete.append(key)
        for key in keys_to_delete:
            await self.delete(key)
    
    async def clear(self):
        self._data.clear()
        self._order.clear()
    
    def get_stats(self) -> Dict:
        return {
            "size": self.size,
            "max_size": config.get_namespace_config(self.namespace).max_size,
            "ttl": config.get_namespace_config(self.namespace).ttl
        }


class CacheManager:
    def __init__(self):
        self._namespaces: Dict[str, NamespaceCache] = {}
        self._lock = asyncio.Lock()
    
    def _get_namespace(self, namespace: str) -> NamespaceCache:
        if namespace not in self._namespaces:
            self._namespaces[namespace] = NamespaceCache(namespace)
        return self._namespaces[namespace]
    
    @staticmethod
    def generate_key(path: str, query_params: Dict[str, Any] = None) -> str:
        query_str = ""
        if query_params:
            sorted_items = sorted(query_params.items())
            query_str = "&".join(f"{k}={v}" for k, v in sorted_items)
        
        key_base = f"{path}?{query_str}" if query_str else path
        return hashlib.sha256(key_base.encode()).hexdigest()
    
    @staticmethod
    def get_namespace_from_path(path: str) -> str:
        parts = path.strip("/").split("/")
        if parts and parts[0]:
            return parts[0]
        return "default"
    
    async def get(self, namespace: str, key: str) -> Optional[CacheEntry]:
        ns = self._get_namespace(namespace)
        return await ns.get(key)
    
    async def set(self, namespace: str, key: str, value: Any, status_code: int, ttl: Optional[int] = None, path: str = ""):
        if ttl is None:
            ttl = config.get_namespace_config(namespace).ttl
        
        ns = self._get_namespace(namespace)
        await ns.set(key, value, status_code, ttl, path)
    
    async def delete(self, namespace: str, key: str):
        if namespace in self._namespaces:
            await self._namespaces[namespace].delete(key)
    
    async def invalidate_by_prefix(self, namespace: str, prefix: str):
        if namespace in self._namespaces:
            await self._namespaces[namespace].invalidate_by_prefix(prefix)
    
    async def clear_namespace(self, namespace: str):
        if namespace in self._namespaces:
            await self._namespaces[namespace].clear()
    
    async def clear_all(self):
        for namespace in list(self._namespaces.keys()):
            await self._namespaces[namespace].clear()
    
    def get_all_stats(self) -> Dict[str, Dict]:
        return {
            ns: cache.get_stats()
            for ns, cache in self._namespaces.items()
        }


cache_manager = CacheManager()
