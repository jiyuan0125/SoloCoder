import threading
import time
from collections import defaultdict, OrderedDict
from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional, Tuple
from datetime import datetime


@dataclass
class CacheEntry:
    key: str
    value: Any
    frequency: int = 1
    inserted_time: float = field(default_factory=time.time)
    last_access_time: float = field(default_factory=time.time)
    ttl: Optional[int] = None  # TTL in seconds
    expire_time: Optional[float] = None
    is_hot: bool = False


@dataclass
class CacheStats:
    total_hits: int = 0
    total_misses: int = 0
    total_items: int = 0
    min_frequency: int = 0
    hit_rate: float = 0.0


class LFUCache:
    def __init__(
        self,
        max_size: int = 1000,
        default_ttl: int = 300,
        hot_ttl: int = 1800,
        hot_key_count: int = 20,
        hot_key_refresh_interval: int = 60,
    ):
        self.max_size = max_size
        self.default_ttl = default_ttl
        self.hot_ttl = hot_ttl
        self.hot_key_count = hot_key_count
        self.hot_key_refresh_interval = hot_key_refresh_interval
        
        self.lock = threading.RLock()
        self.entries: Dict[str, CacheEntry] = {}
        self.frequency_map: Dict[int, OrderedDict[str, CacheEntry]] = defaultdict(OrderedDict)
        self.min_frequency = 1
        
        self.total_hits = 0
        self.total_misses = 0
        self.hot_keys: List[Tuple[str, int]] = []
        self.last_hot_key_refresh = 0.0
        
        self._cleanup_thread = None
        self._stop_cleanup = threading.Event()
        self._start_cleanup_thread()

    def _start_cleanup_thread(self):
        self._cleanup_thread = threading.Thread(target=self._cleanup_worker, daemon=True)
        self._cleanup_thread.start()

    def _cleanup_worker(self):
        while not self._stop_cleanup.is_set():
            time.sleep(min(self.hot_key_refresh_interval, 10))
            self._cleanup_expired()
            self._refresh_hot_keys()

    def stop(self):
        self._stop_cleanup.set()

    def _cleanup_expired(self):
        with self.lock:
            current_time = time.time()
            keys_to_remove = [
                key for key, entry in self.entries.items()
                if entry.expire_time and current_time > entry.expire_time
            ]
            for key in keys_to_remove:
                self._delete_entry(key)

    def _refresh_hot_keys(self):
        with self.lock:
            current_time = time.time()
            if current_time - self.last_hot_key_refresh >= self.hot_key_refresh_interval:
                sorted_entries = sorted(
                    self.entries.values(),
                    key=lambda e: (-e.frequency, e.inserted_time)
                )
                top_entries = sorted_entries[:self.hot_key_count]
                
                hot_keys_set = {entry.key for entry in top_entries}
                current_hot_keys_set = {entry.key for entry in self.entries.values() if entry.is_hot}
                
                keys_to_promote = hot_keys_set - current_hot_keys_set
                keys_to_demote = current_hot_keys_set - hot_keys_set
                
                for key in keys_to_promote:
                    self._promote_to_hot(key)
                
                for key in keys_to_demote:
                    self._demote_from_hot(key)
                
                self.hot_keys = [(entry.key, entry.frequency) for entry in top_entries]
                self.last_hot_key_refresh = current_time

    def _promote_to_hot(self, key: str):
        if key in self.entries:
            entry = self.entries[key]
            entry.is_hot = True
            if entry.ttl is not None:
                self._update_expire_time(entry, self.hot_ttl)

    def _demote_from_hot(self, key: str):
        if key in self.entries:
            entry = self.entries[key]
            entry.is_hot = False

    def _update_expire_time(self, entry: CacheEntry, ttl: int):
        entry.ttl = ttl
        if ttl is not None and ttl > 0:
            entry.expire_time = time.time() + ttl
        else:
            entry.expire_time = None

    def _is_expired(self, entry: CacheEntry) -> bool:
        return entry.expire_time is not None and time.time() > entry.expire_time

    def _get_entry(self, key: str) -> Optional[CacheEntry]:
        with self.lock:
            entry = self.entries.get(key)
            if entry and self._is_expired(entry):
                self._delete_entry(key)
                return None
            return entry

    def _delete_entry(self, key: str):
        if key not in self.entries:
            return
        entry = self.entries.pop(key)
        if entry.frequency in self.frequency_map and key in self.frequency_map[entry.frequency]:
            del self.frequency_map[entry.frequency][key]
            if not self.frequency_map[entry.frequency]:
                del self.frequency_map[entry.frequency]
                if entry.frequency == self.min_frequency:
                    if self.frequency_map:
                        self.min_frequency = min(self.frequency_map.keys())
                    else:
                        self.min_frequency = 1

    def _increment_frequency(self, key: str):
        if key not in self.entries:
            return
        entry = self.entries[key]
        del self.frequency_map[entry.frequency][key]
        if not self.frequency_map[entry.frequency]:
            del self.frequency_map[entry.frequency]
            if entry.frequency == self.min_frequency:
                self.min_frequency += 1
        entry.frequency += 1
        entry.last_access_time = time.time()
        self.frequency_map[entry.frequency][key] = entry

    def _evict(self):
        if not self.frequency_map:
            return
        min_freq = self.min_frequency
        if min_freq in self.frequency_map and self.frequency_map[min_freq]:
            oldest_key, _ = next(iter(self.frequency_map[min_freq].items()))
            self._delete_entry(oldest_key)

    def get(self, key: str) -> Optional[Any]:
        with self.lock:
            entry = self._get_entry(key)
            if entry:
                self._increment_frequency(key)
                self.total_hits += 1
                
                if entry.is_hot:
                    self._update_expire_time(entry, self.hot_ttl)
                else:
                    if entry.ttl is not None:
                        self._update_expire_time(entry, self.default_ttl)
                
                self._check_and_promote_to_hot(key)
                return entry.value
            else:
                self.total_misses += 1
                return None

    def _check_and_promote_to_hot(self, key: str):
        if key not in self.entries:
            return
        entry = self.entries[key]
        if entry.is_hot:
            return
        
        sorted_entries = sorted(
            self.entries.values(),
            key=lambda e: (-e.frequency, e.inserted_time)
        )
        for i, e in enumerate(sorted_entries[:self.hot_key_count]):
            if e.key == key:
                self._promote_to_hot(key)
                break

    def put(self, key: str, value: Any, ttl: Optional[int] = None) -> None:
        with self.lock:
            if key in self.entries:
                self._delete_entry(key)
            
            while len(self.entries) >= self.max_size:
                self._evict()
            
            actual_ttl = ttl if ttl is not None else self.default_ttl
            entry = CacheEntry(
                key=key,
                value=value,
                ttl=actual_ttl,
            )
            if actual_ttl is not None and actual_ttl > 0:
                entry.expire_time = time.time() + actual_ttl
            
            self.entries[key] = entry
            self.frequency_map[1][key] = entry
            self.min_frequency = 1

    def delete(self, key: str) -> bool:
        with self.lock:
            if key in self.entries:
                self._delete_entry(key)
                return True
            return False

    def get_stats(self) -> CacheStats:
        with self.lock:
            total = self.total_hits + self.total_misses
            hit_rate = self.total_hits / total if total > 0 else 0.0
            return CacheStats(
                total_hits=self.total_hits,
                total_misses=self.total_misses,
                total_items=len(self.entries),
                min_frequency=self.min_frequency,
                hit_rate=hit_rate,
            )

    def get_hot_keys(self) -> List[Dict[str, Any]]:
        with self.lock:
            self._refresh_hot_keys()
            return [
                {
                    "key": key,
                    "frequency": freq,
                    "is_hot": self.entries.get(key, None) is not None and self.entries[key].is_hot
                }
                for key, freq in self.hot_keys
            ]

    def batch_put(self, items: List[Tuple[str, Any, Optional[int]]]) -> bool:
        with self.lock:
            original_entries = {k: v for k, v in self.entries.items()}
            original_frequency_map = defaultdict(OrderedDict)
            for freq, od in self.frequency_map.items():
                original_frequency_map[freq] = OrderedDict(od)
            original_min_frequency = self.min_frequency
            
            try:
                for key, value, ttl in items:
                    self.put(key, value, ttl)
                return True
            except Exception:
                self.entries = original_entries
                self.frequency_map = original_frequency_map
                self.min_frequency = original_min_frequency
                return False
