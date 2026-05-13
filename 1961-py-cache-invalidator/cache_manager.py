from typing import Dict, Set, Optional, List
from collections import defaultdict
from dataclasses import dataclass, field
import time
from enum import Enum


class InvalidationReason(str, Enum):
    MANUAL = "手动失效"
    CASCADE = "依赖级联"
    PREFIX = "前缀匹配"
    DELETE = "手动删除"


@dataclass
class CacheEntry:
    key: str
    ttl: int
    created_at: float
    is_invalidated: bool = False


@dataclass
class InvalidationLog:
    keys: List[str]
    reason: InvalidationReason
    operator: Optional[str]
    timestamp: float
    id: int


class CacheManager:
    def __init__(self, max_logs: int = 5000):
        self.cache: Dict[str, CacheEntry] = {}
        self.parent_to_children: Dict[str, Set[str]] = defaultdict(set)
        self.child_to_parents: Dict[str, Set[str]] = defaultdict(set)
        self.prefix_index: Dict[str, Set[str]] = defaultdict(set)
        self.logs: List[InvalidationLog] = []
        self.max_logs = max_logs
        self._log_id_counter = 0

    def register(self, key: str, ttl: int) -> CacheEntry:
        if key in self.cache:
            entry = self.cache[key]
            entry.ttl = ttl
            entry.is_invalidated = False
        else:
            entry = CacheEntry(key=key, ttl=ttl, created_at=time.time())
            self.cache[key] = entry
        self._update_prefix_index(key)
        return entry

    def _update_prefix_index(self, key: str):
        for i in range(1, len(key) + 1):
            prefix = key[:i]
            self.prefix_index[prefix].add(key)

    def get(self, key: str) -> Optional[CacheEntry]:
        return self.cache.get(key)

    def register_dependency(self, parent_key: str, child_key: str):
        if child_key not in self.cache:
            self.register(child_key, ttl=3600)
        self.parent_to_children[parent_key].add(child_key)
        self.child_to_parents[child_key].add(parent_key)

    def get_dependencies(self, key: str) -> List[str]:
        return list(self.child_to_parents.get(key, set()))

    def get_children(self, key: str) -> List[str]:
        return list(self.parent_to_children.get(key, set()))

    def _collect_cascade_keys(self, start_key: str) -> Set[str]:
        collected = set()
        stack = [start_key]
        while stack:
            current = stack.pop()
            if current in collected:
                continue
            collected.add(current)
            for child in self.parent_to_children.get(current, set()):
                stack.append(child)
        return collected

    def invalidate(self, key: str, operator: Optional[str] = None) -> List[str]:
        all_keys = self._collect_cascade_keys(key)
        invalidated_keys = []
        for k in all_keys:
            if k in self.cache and not self.cache[k].is_invalidated:
                self.cache[k].is_invalidated = True
                invalidated_keys.append(k)
        if invalidated_keys:
            cascade_keys = [k for k in invalidated_keys if k != key]
            reason = InvalidationReason.MANUAL if key in invalidated_keys else InvalidationReason.CASCADE
            self._log(invalidated_keys, reason, operator)
        return invalidated_keys

    def invalidate_by_prefix(self, prefix: str, operator: Optional[str] = None) -> List[str]:
        keys = list(self.prefix_index.get(prefix, set()))
        invalidated = []
        for key in keys:
            if key in self.cache and not self.cache[key].is_invalidated:
                self.cache[key].is_invalidated = True
                invalidated.append(key)
        if invalidated:
            self._log(invalidated, InvalidationReason.PREFIX, operator)
        return invalidated

    def _log(self, keys: List[str], reason: InvalidationReason, operator: Optional[str]):
        self._log_id_counter += 1
        log = InvalidationLog(
            id=self._log_id_counter,
            keys=keys,
            reason=reason,
            operator=operator,
            timestamp=time.time()
        )
        self.logs.append(log)
        if len(self.logs) > self.max_logs:
            self.logs = self.logs[-self.max_logs:]

    def query_logs(
        self,
        start_time: Optional[float] = None,
        end_time: Optional[float] = None,
        reason: Optional[InvalidationReason] = None
    ) -> List[InvalidationLog]:
        result = []
        for log in reversed(self.logs):
            if start_time is not None and log.timestamp < start_time:
                continue
            if end_time is not None and log.timestamp > end_time:
                continue
            if reason is not None and log.reason != reason:
                continue
            result.append(log)
        return result
