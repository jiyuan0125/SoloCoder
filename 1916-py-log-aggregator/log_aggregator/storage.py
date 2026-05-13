from abc import ABC, abstractmethod
from collections import deque
from datetime import datetime
import asyncio
import os
import threading


class BaseStorage(ABC):
    @abstractmethod
    async def store(self, log_entry):
        pass

    @abstractmethod
    async def query(self, filters=None, sort=None):
        pass


class MemoryStorage(BaseStorage):
    def __init__(self, capacity=10000):
        self.capacity = capacity
        self.logs = deque(maxlen=capacity)
        self._lock = asyncio.Lock()

    async def store(self, log_entry):
        async with self._lock:
            self.logs.append(log_entry)

    async def query(self, filters=None, sort=None):
        logs = list(self.logs)
        if filters:
            logs = self._filter_logs(logs, filters)
        if sort:
            logs = self._sort_logs(logs, sort)
        return logs

    @staticmethod
    def _filter_logs(logs, filters):
        filtered = []
        for log in logs:
            match = True
            if "from_timestamp" in filters:
                if log.get("timestamp") < filters["from_timestamp"]:
                    match = False
            if "to_timestamp" in filters:
                if log.get("timestamp") > filters["to_timestamp"]:
                    match = False
            if "level" in filters:
                if log.get("level") != filters["level"]:
                    match = False
            if "service" in filters:
                if log.get("service") != filters["service"]:
                    match = False
            if match:
                filtered.append(log)
        return filtered

    @staticmethod
    def _sort_logs(logs, sort):
        if not sort:
            return logs
        
        field, direction = sort[0], sort[1]
        
        def sort_key(log):
            return log.get(field, "")
        
        return sorted(logs, key=sort_key, reverse=(direction == "desc"))


class FileStorage(BaseStorage):
    def __init__(self, service_name, log_dir="./logs"):
        self.service_name = service_name
        self.log_dir = log_dir
        self._current_hour = None
        self._current_file = None
        self._lock = asyncio.Lock()
        self._file_write_lock = threading.Lock()
        os.makedirs(log_dir, exist_ok=True)

    def _get_log_filename(self, dt):
        date_str = dt.strftime("%Y%m%d")
        hour_str = dt.strftime("%H")
        return f"logs-{self.service_name}-{date_str}-{hour_str}.log"

    def _get_file_path(self, dt):
        return os.path.join(self.log_dir, self._get_log_filename(dt))

    def _ensure_file_for_hour(self, dt):
        hour = dt.replace(minute=0, second=0, microsecond=0)
        if self._current_hour != hour:
            if self._current_file:
                self._current_file.close()
            file_path = self._get_file_path(dt)
            self._current_file = open(file_path, "a", encoding="utf-8")
            self._current_hour = hour

    async def store(self, log_entry):
        async with self._lock:
            ts = log_entry.get("timestamp")
            if isinstance(ts, str):
                try:
                    dt = datetime.fromisoformat(ts)
                except ValueError:
                    dt = datetime.now()
            else:
                try:
                    dt = datetime.fromtimestamp(ts)
                except (ValueError, TypeError):
                    dt = datetime.now()
            
            def _write():
                with self._file_write_lock:
                    self._ensure_file_for_hour(dt)
                    import json
                    self._current_file.write(json.dumps(log_entry, ensure_ascii=False) + "\n")
                    self._current_file.flush()
            
            loop = asyncio.get_event_loop()
            await loop.run_in_executor(None, _write)

    async def query(self, filters=None, sort=None):
        all_logs = []
        
        log_files = self._find_log_files(filters)
        
        for file_path in log_files:
            logs_from_file = await self._read_logs_from_file(file_path)
            all_logs.extend(logs_from_file)
        
        if filters:
            all_logs = self._filter_logs(all_logs, filters)
        
        if sort:
            all_logs = self._sort_logs(all_logs, sort)
        
        return all_logs

    def _find_log_files(self, filters):
        import glob
        
        pattern = os.path.join(self.log_dir, f"logs-{self.service_name}-*.log")
        all_files = glob.glob(pattern)
        
        if not filters or ("from_timestamp" not in filters and "to_timestamp" not in filters):
            return sorted(all_files)
        
        filtered_files = []
        for file_path in all_files:
            filename = os.path.basename(file_path)
            try:
                parts = filename.split("-")
                date_str = parts[-2]
                hour_str = parts[-1].split(".")[0]
                file_dt = datetime.strptime(f"{date_str}{hour_str}", "%Y%m%d%H")
                
                should_include = True
                if "from_timestamp" in filters:
                    from_dt = self._parse_timestamp(filters["from_timestamp"])
                    if file_dt < from_dt.replace(minute=0, second=0, microsecond=0):
                        should_include = False
                if "to_timestamp" in filters:
                    to_dt = self._parse_timestamp(filters["to_timestamp"])
                    if file_dt > to_dt.replace(minute=0, second=0, microsecond=0):
                        should_include = False
                
                if should_include:
                    filtered_files.append(file_path)
            except (ValueError, IndexError):
                continue
        
        return sorted(filtered_files)

    @staticmethod
    def _parse_timestamp(ts):
        if isinstance(ts, str):
            return datetime.fromisoformat(ts)
        try:
            return datetime.fromtimestamp(ts)
        except (ValueError, TypeError):
            return datetime.now()

    async def _read_logs_from_file(self, file_path):
        import json
        
        def _read():
            logs = []
            with open(file_path, "r", encoding="utf-8") as f:
                for line in f:
                    line = line.strip()
                    if line:
                        try:
                            logs.append(json.loads(line))
                        except json.JSONDecodeError:
                            continue
            return logs
        
        loop = asyncio.get_event_loop()
        return await loop.run_in_executor(None, _read)

    @staticmethod
    def _filter_logs(logs, filters):
        filtered = []
        for log in logs:
            match = True
            if "from_timestamp" in filters:
                from_ts = filters["from_timestamp"]
                if isinstance(from_ts, str):
                    from_ts = datetime.fromisoformat(from_ts).isoformat()
                if log.get("timestamp") < from_ts:
                    match = False
            if "to_timestamp" in filters:
                to_ts = filters["to_timestamp"]
                if isinstance(to_ts, str):
                    to_ts = datetime.fromisoformat(to_ts).isoformat()
                if log.get("timestamp") > to_ts:
                    match = False
            if "level" in filters:
                if log.get("level") != filters["level"]:
                    match = False
            if "service" in filters:
                if log.get("service") != filters["service"]:
                    match = False
            if match:
                filtered.append(log)
        return filtered

    @staticmethod
    def _sort_logs(logs, sort):
        if not sort:
            return logs
        
        field, direction = sort[0], sort[1]
        
        def sort_key(log):
            return log.get(field, "")
        
        return sorted(logs, key=sort_key, reverse=(direction == "desc"))

    def __del__(self):
        if self._current_file:
            self._current_file.close()
