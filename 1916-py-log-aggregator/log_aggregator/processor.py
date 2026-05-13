import asyncio
import random
from datetime import datetime

from .storage import MemoryStorage, FileStorage
from .config import ServiceConfigManager


class LogProcessor:
    def __init__(self, log_dir="./logs"):
        self.config_manager = ServiceConfigManager()
        self._storages = {}
        self._storages_lock = asyncio.Lock()
        self._log_dir = log_dir
    
    async def _get_or_create_storage(self, service_name, storage_type):
        async with self._storages_lock:
            if service_name in self._storages:
                existing = self._storages[service_name]
                if storage_type == "memory" and isinstance(existing, MemoryStorage):
                    return existing
                if storage_type == "file" and isinstance(existing, FileStorage):
                    return existing
            
            if storage_type == "memory":
                storage = MemoryStorage(capacity=10000)
            else:
                storage = FileStorage(service_name, log_dir=self._log_dir)
            
            self._storages[service_name] = storage
            return storage
    
    async def process_log(self, log_entry):
        level = log_entry.get("level", "").upper()
        service = log_entry.get("service")
        
        if not service:
            return False
        
        config = await self.config_manager.get_config(service)
        storage = await self._get_or_create_storage(service, config["storage"])
        
        should_store = self._should_store_log(log_entry, level, config)
        
        if should_store:
            await storage.store(log_entry)
            return True
        
        return False
    
    async def process_logs(self, log_entries):
        results = []
        for log_entry in log_entries:
            stored = await self.process_log(log_entry)
            results.append(stored)
        return results
    
    @staticmethod
    def _should_store_log(log_entry, level, config):
        if level in ("ERROR", "FATAL"):
            return True
        
        if level == "WARN":
            sample_rate = config.get("warn_sample_rate", 1.0)
            if sample_rate <= 0:
                return False
            if sample_rate >= 1:
                return True
            return random.random() < sample_rate
        
        if level == "INFO":
            keywords = config.get("info_keywords", [])
            if not keywords:
                return False
            message = log_entry.get("message", "")
            message_lower = message.lower()
            return any(keyword.lower() in message_lower for keyword in keywords)
        
        return False
    
    async def query_logs(self, filters=None, sort=None):
        storages_to_query = []
        
        async with self._storages_lock:
            storages_copy = dict(self._storages)
        
        if filters and "service" in filters:
            service = filters["service"]
            if service in storages_copy:
                storages_to_query.append((service, storages_copy[service]))
        else:
            storages_to_query = list(storages_copy.items())
        
        all_logs = []
        for service, storage in storages_to_query:
            service_logs = await storage.query(filters=filters, sort=None)
            all_logs.extend(service_logs)
        
        if sort:
            all_logs = self._sort_logs(all_logs, sort)
        
        return all_logs
    
    @staticmethod
    def _sort_logs(logs, sort):
        if not sort:
            return logs
        
        field, direction = sort[0], sort[1]
        
        def sort_key(log):
            return log.get(field, "")
        
        return sorted(logs, key=sort_key, reverse=(direction == "desc"))
