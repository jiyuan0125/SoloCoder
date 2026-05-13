import threading
import time
from collections import deque
from typing import Dict, Deque, List, Optional
from .models import Span


class TraceStorage:
    MAX_SPANS_PER_TRACE = 1000
    DEFAULT_TTL_SECONDS = 3600

    def __init__(self, ttl_seconds: int = DEFAULT_TTL_SECONDS):
        self._traces: Dict[str, Deque[Span]] = {}
        self._trace_last_access: Dict[str, float] = {}
        self._ttl_seconds = ttl_seconds
        self._lock = threading.RLock()

    def _check_ttl(self):
        current_time = time.time()
        expired_traces = [
            trace_id
            for trace_id, last_access in self._trace_last_access.items()
            if current_time - last_access > self._ttl_seconds
        ]
        for trace_id in expired_traces:
            del self._traces[trace_id]
            del self._trace_last_access[trace_id]

    def _touch_trace(self, trace_id: str):
        self._trace_last_access[trace_id] = time.time()

    def add_span(self, span: Span) -> None:
        with self._lock:
            self._check_ttl()
            trace_id = span.trace_id

            if trace_id not in self._traces:
                self._traces[trace_id] = deque(maxlen=self.MAX_SPANS_PER_TRACE)

            self._traces[trace_id].append(span)
            self._touch_trace(trace_id)

    def add_spans(self, spans: List[Span]) -> None:
        with self._lock:
            self._check_ttl()
            for span in spans:
                self.add_span(span)

    def get_trace(self, trace_id: str) -> Optional[List[Span]]:
        with self._lock:
            self._check_ttl()
            if trace_id not in self._traces:
                return None

            spans = list(self._traces[trace_id])
            spans.sort(key=lambda s: s.start_time)
            self._touch_trace(trace_id)
            return spans

    def get_spans_by_service(
        self,
        service_name: str,
        start_time: float = None,
        end_time: float = None
    ) -> List[Span]:
        with self._lock:
            self._check_ttl()
            result = []

            for trace_spans in self._traces.values():
                for span in trace_spans:
                    if span.service_name == service_name:
                        if start_time is not None and span.start_time < start_time:
                            continue
                        if end_time is not None and span.end_time > end_time:
                            continue
                        result.append(span)

            result.sort(key=lambda s: s.start_time)
            return result

    def get_orphan_spans(self) -> List[Span]:
        with self._lock:
            self._check_ttl()
            result = []

            for trace_spans in self._traces.values():
                span_ids = {s.span_id for s in trace_spans}
                for span in trace_spans:
                    if span.parent_span_id and span.parent_span_id not in span_ids:
                        result.append(span)

            result.sort(key=lambda s: s.start_time)
            return result

    def get_all_spans(self) -> List[Span]:
        with self._lock:
            self._check_ttl()
            result = []
            for trace_spans in self._traces.values():
                result.extend(trace_spans)
            return result

    def clear(self):
        with self._lock:
            self._traces.clear()
            self._trace_last_access.clear()
