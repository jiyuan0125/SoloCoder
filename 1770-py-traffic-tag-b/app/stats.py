import threading
import time
from typing import Dict, Any, Optional
from collections import defaultdict
from app.models import TrafficType


class StatsManager:
    def __init__(self):
        self._lock = threading.RLock()
        self._reset_stats()

    def _reset_stats(self):
        self._request_count = {
            TrafficType.NORMAL: 0,
            TrafficType.GRAY: 0
        }
        self._response_count = {
            TrafficType.NORMAL: defaultdict(int),
            TrafficType.GRAY: defaultdict(int)
        }
        self._path_stats = {
            TrafficType.NORMAL: defaultdict(lambda: {
                "count": 0,
                "total_duration": 0.0,
                "error_count": 0
            }),
            TrafficType.GRAY: defaultdict(lambda: {
                "count": 0,
                "total_duration": 0.0,
                "error_count": 0
            })
        }
        self._method_stats = {
            TrafficType.NORMAL: defaultdict(int),
            TrafficType.GRAY: defaultdict(int)
        }
        self._status_code_stats = {
            TrafficType.NORMAL: defaultdict(int),
            TrafficType.GRAY: defaultdict(int)
        }
        self._total_duration = {
            TrafficType.NORMAL: 0.0,
            TrafficType.GRAY: 0.0
        }
        self._start_time = time.time()

    def record_request(self, traffic_type: TrafficType, method: str, path: str):
        with self._lock:
            self._request_count[traffic_type] += 1
            self._method_stats[traffic_type][method] += 1

    def record_response(
        self,
        traffic_type: TrafficType,
        method: str,
        path: str,
        status_code: int,
        duration: float
    ):
        with self._lock:
            self._response_count[traffic_type][status_code] += 1
            self._status_code_stats[traffic_type][status_code] += 1
            self._total_duration[traffic_type] += duration

            path_stat = self._path_stats[traffic_type][path]
            path_stat["count"] += 1
            path_stat["total_duration"] += duration
            if status_code >= 400:
                path_stat["error_count"] += 1

    def get_stats(self, traffic_type: Optional[TrafficType] = None) -> Dict[str, Any]:
        with self._lock:
            if traffic_type:
                return self._get_single_stats(traffic_type)
            
            return {
                "normal": self._get_single_stats(TrafficType.NORMAL),
                "gray": self._get_single_stats(TrafficType.GRAY),
                "combined": self._get_combined_stats(),
                "uptime_seconds": time.time() - self._start_time
            }

    def _get_single_stats(self, traffic_type: TrafficType) -> Dict[str, Any]:
        total_requests = self._request_count[traffic_type]
        if total_requests == 0:
            return {
                "total_requests": 0,
                "total_responses": 0,
                "avg_duration_ms": 0,
                "error_rate": 0,
                "by_method": {},
                "by_status_code": {},
                "by_path": {}
            }

        total_responses = sum(self._response_count[traffic_type].values())
        avg_duration = 0
        error_count = 0

        if total_responses > 0:
            avg_duration = (self._total_duration[traffic_type] / total_responses) * 1000
        
        for status in self._status_code_stats[traffic_type]:
            if status >= 400:
                error_count += self._status_code_stats[traffic_type][status]

        error_rate = (error_count / total_requests) * 100 if total_requests > 0 else 0

        by_path = {}
        for path, stat in self._path_stats[traffic_type].items():
            path_avg = 0
            if stat["count"] > 0:
                path_avg = (stat["total_duration"] / stat["count"]) * 1000
            by_path[path] = {
                "count": stat["count"],
                "avg_duration_ms": round(path_avg, 2),
                "error_count": stat["error_count"]
            }

        return {
            "total_requests": total_requests,
            "total_responses": total_responses,
            "avg_duration_ms": round(avg_duration, 2),
            "error_rate": round(error_rate, 2),
            "by_method": dict(self._method_stats[traffic_type]),
            "by_status_code": dict(self._status_code_stats[traffic_type]),
            "by_path": by_path
        }

    def _get_combined_stats(self) -> Dict[str, Any]:
        normal_stats = self._get_single_stats(TrafficType.NORMAL)
        gray_stats = self._get_single_stats(TrafficType.GRAY)
        
        total_requests = normal_stats["total_requests"] + gray_stats["total_requests"]
        total_responses = normal_stats["total_responses"] + gray_stats["total_responses"]
        
        return {
            "total_requests": total_requests,
            "total_responses": total_responses,
            "gray_ratio": round(
                (gray_stats["total_requests"] / total_requests * 100)
                if total_requests > 0 else 0,
                2
            )
        }

    def reset(self):
        with self._lock:
            self._reset_stats()


stats_manager = StatsManager()
