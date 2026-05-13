from typing import Dict, Any
from dataclasses import dataclass, field
from collections import defaultdict


@dataclass
class NamespaceStats:
    hits: int = 0
    misses: int = 0
    breakdown_protected: int = 0
    
    def hit(self):
        self.hits += 1
    
    def miss(self):
        self.misses += 1
    
    def breakdown_protect(self):
        self.breakdown_protected += 1
    
    def get_hit_rate(self) -> float:
        total = self.hits + self.misses
        if total == 0:
            return 0.0
        return round(self.hits / total * 100, 2)
    
    def to_dict(self) -> Dict[str, Any]:
        return {
            "hits": self.hits,
            "misses": self.misses,
            "breakdown_protected": self.breakdown_protected,
            "hit_rate": self.get_hit_rate()
        }


class StatsManager:
    def __init__(self):
        self._stats: Dict[str, NamespaceStats] = defaultdict(NamespaceStats)
    
    def hit(self, namespace: str):
        self._stats[namespace].hit()
    
    def miss(self, namespace: str):
        self._stats[namespace].miss()
    
    def breakdown_protect(self, namespace: str):
        self._stats[namespace].breakdown_protect()
    
    def get_stats(self, namespace: str = None) -> Dict[str, Any]:
        if namespace:
            return self._stats[namespace].to_dict()
        
        return {
            ns: stats.to_dict()
            for ns, stats in self._stats.items()
        }
    
    def get_all_stats(self) -> Dict[str, Any]:
        totals = NamespaceStats()
        for stats in self._stats.values():
            totals.hits += stats.hits
            totals.misses += stats.misses
            totals.breakdown_protected += stats.breakdown_protected
        
        return {
            "namespaces": self.get_stats(),
            "total": totals.to_dict()
        }
    
    def reset(self, namespace: str = None):
        if namespace:
            if namespace in self._stats:
                del self._stats[namespace]
        else:
            self._stats.clear()


stats_manager = StatsManager()
