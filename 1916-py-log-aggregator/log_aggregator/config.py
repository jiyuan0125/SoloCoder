import asyncio
import random


DEFAULT_CONFIG = {
    "storage": "memory",
    "warn_sample_rate": 1.0,
    "info_keywords": []
}


class ServiceConfigManager:
    def __init__(self):
        self._configs = {}
        self._lock = asyncio.Lock()
    
    async def get_config(self, service_name):
        async with self._lock:
            return self._configs.get(service_name, dict(DEFAULT_CONFIG))
    
    async def set_config(self, service_name, config):
        async with self._lock:
            existing = self._configs.get(service_name, dict(DEFAULT_CONFIG))
            merged = {**existing, **config}
            
            if "warn_sample_rate" in merged:
                rate = merged["warn_sample_rate"]
                if not (0.0 <= rate <= 1.0):
                    raise ValueError("warn_sample_rate must be between 0 and 1")
            
            if "storage" in merged:
                storage = merged["storage"]
                if storage not in ("memory", "file"):
                    raise ValueError("storage must be 'memory' or 'file'")
            
            self._configs[service_name] = merged
            return merged
    
    async def get_all_configs(self):
        async with self._lock:
            return dict(self._configs)
