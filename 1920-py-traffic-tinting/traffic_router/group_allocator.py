import bisect
import hashlib
from threading import RLock
from typing import Dict, Optional


class GroupAllocator:
    def __init__(self, initial_ratios: Dict[str, int]):
        self._ratios: Dict[str, int] = dict(initial_ratios)
        self._history: Dict[str, str] = {}
        self._sorted_points: list = []
        self._point_map: Dict[int, str] = {}
        self._lock = RLock()
        self._rebalance_ring()

    def get_group(self, user_id: str) -> Optional[str]:
        with self._lock:
            if user_id in self._history:
                return self._history[user_id]
            assigned_group = self._get_from_ring(user_id)
            if assigned_group is not None:
                self._history[user_id] = assigned_group
            return assigned_group

    def update_ratios(self, new_ratios: Dict[str, int]):
        with self._lock:
            self._ratios = dict(new_ratios)
            self._rebalance_ring()

    def get_ratios(self) -> Dict[str, int]:
        with self._lock:
            return dict(self._ratios)

    def _rebalance_ring(self):
        ratio_range = 10000
        ring_points: list = []
        total_ratio = sum(self._ratios.values())
        if total_ratio <= 0:
            self._sorted_points = []
            self._point_map = {}
            return
        for group, ratio in self._ratios.items():
            num_points = max(1, int((ratio / total_ratio) * ratio_range))
            for i in range(num_points):
                point_hash = self._hash(f"{group}:{i}")
                ring_points.append((point_hash, group))
        ring_points.sort(key=lambda x: x[0])
        self._sorted_points = [p[0] for p in ring_points]
        self._point_map = {p[0]: p[1] for p in ring_points}

    def _get_from_ring(self, user_id: str) -> Optional[str]:
        if not self._sorted_points:
            return None
        user_hash = self._hash(user_id)
        idx = bisect.bisect_left(self._sorted_points, user_hash)
        if idx == len(self._sorted_points):
            idx = 0
        return self._point_map[self._sorted_points[idx]]

    @staticmethod
    def _hash(key: str) -> int:
        digest = hashlib.sha1(key.encode("utf-8")).hexdigest()
        return int(digest, 16)
