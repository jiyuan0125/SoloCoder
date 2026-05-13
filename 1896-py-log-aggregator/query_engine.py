from typing import Optional, List, Dict, Any
from dataclasses import dataclass
from collections import defaultdict

from storage import LogStore
from models import LogEntry


@dataclass
class QueryResult:
    logs: List[LogEntry]
    total: int
    page: int
    page_size: int


class QueryEngine:
    def __init__(self, store: LogStore):
        self.store = store
    
    def query(
        self,
        service: Optional[str] = None,
        level: Optional[str] = None,
        start_time: Optional[float] = None,
        end_time: Optional[float] = None,
        message_contains: Optional[str] = None,
        page: int = 1,
        page_size: int = 50,
    ) -> QueryResult:
        page_size = min(page_size, 200)
        
        with self.store._logs_lock:
            candidates = self._get_filtered_indices(service, level, start_time, end_time, message_contains)
            total = len(candidates)
            
            start = (page - 1) * page_size
            end = start + page_size
            
            result_indices = candidates[start:end]
            logs = [self.store.all_logs[idx] for idx in result_indices]
            
            return QueryResult(
                logs=logs,
                total=total,
                page=page,
                page_size=page_size,
            )
    
    def _get_filtered_indices(
        self,
        service: Optional[str],
        level: Optional[str],
        start_time: Optional[float],
        end_time: Optional[float],
        message_contains: Optional[str],
    ) -> List[int]:
        candidates = None
        
        if service:
            with self.store._index_lock:
                candidates = set(self.store._index_by_service.get(service, []))
        
        if level:
            with self.store._index_lock:
                level_indices = set(self.store._index_by_level.get(level, []))
            if candidates is None:
                candidates = level_indices
            else:
                candidates &= level_indices
        
        if candidates is None:
            candidates = set(range(len(self.store.all_logs)))
        
        filtered = []
        msg_lower = message_contains.lower() if message_contains else None
        
        for idx in candidates:
            entry = self.store.all_logs[idx]
            
            if start_time is not None and entry.timestamp < start_time:
                continue
            if end_time is not None and entry.timestamp > end_time:
                continue
            if msg_lower and msg_lower not in entry.message.lower():
                continue
            
            filtered.append(idx)
        
        filtered.sort(key=lambda i: self.store.all_logs[i].timestamp, reverse=True)
        return filtered
    
    def query_by_trace(self, trace_id: str) -> List[LogEntry]:
        with self.store._index_lock:
            indices = self.store._index_by_trace.get(trace_id, [])
        
        with self.store._logs_lock:
            logs = [self.store.all_logs[idx] for idx in indices]
        
        logs.sort(key=lambda e: e.timestamp)
        return logs
    
    def get_statistics(self) -> Dict[str, Any]:
        stats = defaultdict(lambda: {"total": 0, "levels": defaultdict(int)})
        
        with self.store._logs_lock:
            for entry in self.store.all_logs:
                stats[entry.service]["total"] += 1
                stats[entry.service]["levels"][entry.level] += 1
        
        result = {}
        for service, data in stats.items():
            result[service] = {
                "total": data["total"],
                "levels": dict(data["levels"]),
            }
        
        return result
