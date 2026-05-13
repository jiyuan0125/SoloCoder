from typing import Dict, List, Optional
from collections import defaultdict
from .models import Span, SpanTreeNode, ServiceStats
from .storage import TraceStorage


class TraceService:
    def __init__(self, storage: TraceStorage):
        self._storage = storage

    def build_trace_tree(self, trace_id: str) -> Optional[List[SpanTreeNode]]:
        spans = self._storage.get_trace(trace_id)
        if not spans:
            return None

        if not spans:
            return []

        min_start = min(s.start_time for s in spans)
        max_end = max(s.end_time for s in spans)
        total_duration = max_end - min_start

        span_map: Dict[str, SpanTreeNode] = {}
        children_map: Dict[str, List[SpanTreeNode]] = defaultdict(list)

        for span in spans:
            duration = span.end_time - span.start_time
            percentage = (duration / total_duration * 100) if total_duration > 0 else 0

            tree_node = SpanTreeNode(
                trace_id=span.trace_id,
                span_id=span.span_id,
                parent_span_id=span.parent_span_id,
                service_name=span.service_name,
                operation=span.operation,
                start_time=span.start_time,
                end_time=span.end_time,
                status=span.status,
                duration=duration,
                duration_ms=duration * 1000,
                total_duration=total_duration,
                percentage=percentage,
                children=[]
            )
            span_map[span.span_id] = tree_node

            if span.parent_span_id:
                children_map[span.parent_span_id].append(tree_node)

        root_nodes: List[SpanTreeNode] = []

        for span_id, node in span_map.items():
            if node.parent_span_id and node.parent_span_id in span_map:
                parent = span_map[node.parent_span_id]
                parent.children.append(node)
            else:
                root_nodes.append(node)

        for node in span_map.values():
            node.children.sort(key=lambda c: c.start_time)

        root_nodes.sort(key=lambda n: n.start_time)

        return root_nodes

    def get_service_stats(self) -> List[ServiceStats]:
        all_spans = self._storage.get_all_spans()
        if not all_spans:
            return []

        service_data: Dict[str, List[Span]] = defaultdict(list)
        for span in all_spans:
            service_data[span.service_name].append(span)

        stats_list: List[ServiceStats] = []

        for service_name, spans in service_data.items():
            durations = [s.end_time - s.start_time for s in spans]
            error_spans = sum(1 for s in spans if s.status.lower() != "ok")

            total_duration = sum(durations)
            avg_duration = total_duration / len(durations) if durations else 0

            sorted_durations = sorted(durations)
            p99_index = int(len(sorted_durations) * 0.99)
            if sorted_durations:
                p99_duration = sorted_durations[min(p99_index, len(sorted_durations) - 1)]
            else:
                p99_duration = 0

            error_rate = error_spans / len(spans) if spans else 0

            stats_list.append(ServiceStats(
                service_name=service_name,
                average_duration_ms=avg_duration * 1000,
                error_rate=error_rate,
                p99_duration_ms=p99_duration * 1000,
                total_spans=len(spans),
                error_spans=error_spans
            ))

        stats_list.sort(key=lambda s: s.service_name)
        return stats_list
