import threading
from collections import deque, defaultdict
from enum import Enum
from typing import Optional, List, Dict, Any, Tuple


class LogLevel(Enum):
    DEBUG = 10
    INFO = 20
    WARN = 30
    ERROR = 40


LEVEL_MAP = {
    "DEBUG": LogLevel.DEBUG,
    "INFO": LogLevel.INFO,
    "WARN": LogLevel.WARN,
    "ERROR": LogLevel.ERROR,
    "WARNING": LogLevel.WARN,
}

MAX_LOGS_PER_SERVICE = 100000
BATCH_SIZE_LIMIT = 1000
DEFAULT_TOP_N = 20
MAX_TOP_N = 100


class LogEntry:
    def __init__(
        self,
        timestamp: float,
        service_name: str,
        level: str,
        message: str,
        tags: Optional[Dict[str, Any]] = None,
        trace_id: Optional[str] = None,
    ):
        self.timestamp = timestamp
        self.service_name = service_name
        self.level = level.upper()
        self.message = message
        self.tags = tags or {}
        self.trace_id = trace_id
        level_enum = LEVEL_MAP.get(self.level, LogLevel.DEBUG)
        self._level_value = level_enum.value

    def to_dict(self) -> Dict[str, Any]:
        result = {
            "timestamp": self.timestamp,
            "service_name": self.service_name,
            "level": self.level,
            "message": self.message,
            "tags": self.tags,
        }
        if self.trace_id:
            result["trace_id"] = self.trace_id
        return result


class LogAggregator:
    def __init__(self):
        self._lock = threading.RLock()
        self._logs: Dict[str, deque] = defaultdict(lambda: deque(maxlen=MAX_LOGS_PER_SERVICE))
        self._level_config: Dict[str, LogLevel] = {}
        self._default_level = LogLevel.DEBUG
        self._trace_index: Dict[str, List[LogEntry]] = defaultdict(list)
        self._tag_counts: Dict[str, Dict[str, int]] = defaultdict(lambda: defaultdict(int))

    def _should_keep(self, entry: LogEntry) -> bool:
        with self._lock:
            min_level = self._level_config.get(entry.service_name, self._default_level)
        return entry._level_value >= min_level.value

    def _update_tag_counts(self, tags: Dict[str, Any]):
        for key, value in tags.items():
            str_value = str(value)
            self._tag_counts[key][str_value] += 1

    def add_log(self, entry: LogEntry) -> bool:
        if not self._should_keep(entry):
            return False
        with self._lock:
            self._logs[entry.service_name].append(entry)
            if entry.trace_id:
                self._trace_index[entry.trace_id].append(entry)
            if entry.tags:
                self._update_tag_counts(entry.tags)
        return True

    def add_batch(self, entries: List[LogEntry]) -> Tuple[int, int, bool]:
        if len(entries) > BATCH_SIZE_LIMIT:
            entries = entries[:BATCH_SIZE_LIMIT]
            truncated = True
        else:
            truncated = False
        added = 0
        filtered = 0
        for entry in entries:
            if self.add_log(entry):
                added += 1
            else:
                filtered += 1
        return added, filtered, truncated

    def set_service_level(self, service_name: str, level: str):
        with self._lock:
            level_enum = LEVEL_MAP.get(level.upper())
            if level_enum is None:
                raise ValueError(f"Invalid level: {level}")
            self._level_config[service_name] = level_enum

    def get_service_level(self, service_name: str) -> str:
        with self._lock:
            level = self._level_config.get(service_name, self._default_level)
        return level.name

    def query(
        self,
        start_time: Optional[float] = None,
        end_time: Optional[float] = None,
        level: Optional[str] = None,
        service_name: Optional[str] = None,
        tags: Optional[Dict[str, Any]] = None,
        offset: int = 0,
        limit: int = 100,
    ) -> Dict[str, Any]:
        level_enum = LEVEL_MAP.get(level.upper()) if level else None
        level_value = level_enum.value if level_enum else None
        all_entries: List[LogEntry] = []
        with self._lock:
            if service_name:
                services = [service_name] if service_name in self._logs else []
            else:
                services = list(self._logs.keys())
            for svc in services:
                for entry in self._logs[svc]:
                    if start_time is not None and entry.timestamp < start_time:
                        continue
                    if end_time is not None and entry.timestamp > end_time:
                        continue
                    if level_value is not None and entry._level_value < level_value:
                        continue
                    if tags:
                        match = True
                        for k, v in tags.items():
                            if entry.tags.get(k) != v:
                                match = False
                                break
                        if not match:
                            continue
                    all_entries.append(entry)
        all_entries.sort(key=lambda e: e.timestamp, reverse=True)
        total = len(all_entries)
        results = all_entries[offset:offset + limit]
        return {
            "total": total,
            "offset": offset,
            "limit": limit,
            "logs": [e.to_dict() for e in results],
        }

    def get_trace(self, trace_id: str) -> Optional[Dict[str, Any]]:
        with self._lock:
            spans = list(self._trace_index.get(trace_id, []))
        if not spans:
            return None
        spans.sort(key=lambda e: e.timestamp)
        min_time = spans[0].timestamp
        max_time = spans[-1].timestamp
        trace_spans = []
        for i, span in enumerate(spans):
            next_time = spans[i + 1].timestamp if i + 1 < len(spans) else span.timestamp
            duration_ms = int((next_time - span.timestamp) * 1000)
            trace_spans.append({
                "service_name": span.service_name,
                "level": span.level,
                "message": span.message,
                "timestamp": span.timestamp,
                "duration_ms": duration_ms,
                "tags": span.tags,
            })
        total_duration_ms = int((max_time - min_time) * 1000)
        return {
            "trace_id": trace_id,
            "total_duration_ms": total_duration_ms,
            "span_count": len(spans),
            "spans": trace_spans,
        }

    def get_tag_stats(self, key: str, top_n: int = DEFAULT_TOP_N) -> Optional[Dict[str, Any]]:
        n = max(1, min(top_n, MAX_TOP_N))
        with self._lock:
            counts = self._tag_counts.get(key)
            if counts is None:
                return None
            sorted_items = sorted(counts.items(), key=lambda x: x[1], reverse=True)
        top_items = sorted_items[:n]
        return {
            "key": key,
            "top_n": n,
            "values": [{"value": v, "count": c} for v, c in top_items],
        }

    def get_all_service_levels(self) -> Dict[str, str]:
        with self._lock:
            return {svc: level.name for svc, level in self._level_config.items()}
