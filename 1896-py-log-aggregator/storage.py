import threading
import queue
import time
import os
import json
from collections import defaultdict
from typing import Dict, List, Optional, Set
from dataclasses import dataclass, field

from models import LogEntry


@dataclass
class WriteFailure:
    timestamp: float
    count: int
    reason: str


class LogStore:
    def __init__(self, data_dir: str = "./log_data", retention_days: int = 7):
        self.data_dir = data_dir
        self.retention_days = retention_days
        self.service_retention: Dict[str, int] = {}
        
        self.all_logs: List[LogEntry] = []
        self._logs_lock = threading.RLock()
        
        self._index_by_service: Dict[str, List[int]] = defaultdict(list)
        self._index_by_level: Dict[str, List[int]] = defaultdict(list)
        self._index_by_trace: Dict[str, List[int]] = defaultdict(list)
        self._index_lock = threading.RLock()
        
        self.write_queue: queue.Queue = queue.Queue()
        self._write_failures: List[WriteFailure] = []
        self._total_failed = 0
        self._failures_lock = threading.Lock()
        
        self._running = False
        self._writer_thread: Optional[threading.Thread] = None
        self._cleaner_thread: Optional[threading.Thread] = None
        
        os.makedirs(data_dir, exist_ok=True)
        self._load_existing_logs()
    
    def _load_existing_logs(self):
        log_file = os.path.join(self.data_dir, "logs.jsonl")
        if not os.path.exists(log_file):
            return
        
        with self._logs_lock:
            with open(log_file, 'r', encoding='utf-8') as f:
                for line in f:
                    try:
                        data = json.loads(line)
                        entry = LogEntry(
                            timestamp=data['timestamp'],
                            service=data['service'],
                            level=data['level'],
                            message=data['message'],
                            trace_id=data.get('trace_id'),
                            raw_data=data
                        )
                        self._add_to_index(entry)
                        self.all_logs.append(entry)
                    except Exception:
                        continue
        
        self._sort_logs()
    
    def _sort_logs(self):
        self.all_logs.sort(key=lambda e: e.timestamp, reverse=True)
        self._rebuild_indices()
    
    def _rebuild_indices(self):
        with self._index_lock:
            self._index_by_service.clear()
            self._index_by_level.clear()
            self._index_by_trace.clear()
            
            for idx, entry in enumerate(self.all_logs):
                self._index_by_service[entry.service].append(idx)
                self._index_by_level[entry.level].append(idx)
                if entry.trace_id:
                    self._index_by_trace[entry.trace_id].append(idx)
    
    def _add_to_index(self, entry: LogEntry):
        pass
    
    def _add_entry_to_indices(self, entry: LogEntry, idx: int):
        with self._index_lock:
            self._index_by_service[entry.service].append(idx)
            self._index_by_level[entry.level].append(idx)
            if entry.trace_id:
                self._index_by_trace[entry.trace_id].append(idx)
    
    def _persist_entries(self, entries: List[LogEntry]):
        log_file = os.path.join(self.data_dir, "logs.jsonl")
        try:
            with open(log_file, 'a', encoding='utf-8') as f:
                for entry in entries:
                    f.write(json.dumps(entry.to_dict(), ensure_ascii=False) + '\n')
        except Exception as e:
            with self._failures_lock:
                self._write_failures.append(
                    WriteFailure(timestamp=time.time(), count=len(entries), reason=str(e))
                )
                self._total_failed += len(entries)
            raise
    
    def enqueue_entries(self, entries: List[LogEntry]):
        for entry in entries:
            self.write_queue.put(entry)
    
    def _writer_loop(self):
        batch = []
        last_write = time.time()
        
        while self._running:
            try:
                entry = self.write_queue.get(timeout=1.0)
                batch.append(entry)
                
                if len(batch) >= 100 or (time.time() - last_write) >= 1.0:
                    self._flush_batch(batch)
                    batch = []
                    last_write = time.time()
            except queue.Empty:
                if batch:
                    self._flush_batch(batch)
                    batch = []
                    last_write = time.time()
        
        if batch:
            self._flush_batch(batch)
    
    def _flush_batch(self, batch: List[LogEntry]):
        if not batch:
            return
        
        try:
            self._persist_entries(batch)
            with self._logs_lock:
                for entry in batch:
                    self.all_logs.append(entry)
                self._sort_logs()
        except Exception:
            pass
    
    def _cleaner_loop(self):
        while self._running:
            try:
                self._cleanup_expired()
            except Exception:
                pass
            
            for _ in range(600):
                if not self._running:
                    return
                time.sleep(1.0)
    
    def _cleanup_expired(self):
        now = time.time()
        with self._logs_lock:
            retention_map = self.service_retention.copy()
            default_cutoff = now - (self.retention_days * 24 * 3600)
            
            new_logs = []
            removed = 0
            
            for entry in self.all_logs:
                retention_days = retention_map.get(entry.service, self.retention_days)
                cutoff = now - (retention_days * 24 * 3600)
                
                if entry.timestamp >= cutoff:
                    new_logs.append(entry)
                else:
                    removed += 1
            
            if removed > 0:
                self.all_logs = new_logs
                self._rebuild_indices()
                self._rewrite_persisted_logs()
    
    def _rewrite_persisted_logs(self):
        log_file = os.path.join(self.data_dir, "logs.jsonl")
        temp_file = log_file + ".tmp"
        
        try:
            with open(temp_file, 'w', encoding='utf-8') as f:
                for entry in self.all_logs:
                    f.write(json.dumps(entry.to_dict(), ensure_ascii=False) + '\n')
            
            os.replace(temp_file, log_file)
        except Exception:
            if os.path.exists(temp_file):
                os.remove(temp_file)
    
    def start(self):
        self._running = True
        self._writer_thread = threading.Thread(target=self._writer_loop, daemon=True)
        self._writer_thread.start()
        self._cleaner_thread = threading.Thread(target=self._cleaner_loop, daemon=True)
        self._cleaner_thread.start()
    
    def stop(self):
        self._running = False
    
    def get_failure_stats(self) -> Dict:
        with self._failures_lock:
            return {
                "total_failed": self._total_failed,
                "recent_failures": [
                    {"count": f.count, "reason": f.reason, "timestamp": f.timestamp}
                    for f in self._write_failures[-10:]
                ]
            }
    
    def set_service_retention(self, service: str, days: int):
        self.service_retention[service] = days
    
    def get_services(self) -> Set[str]:
        with self._index_lock:
            return set(self._index_by_service.keys())
