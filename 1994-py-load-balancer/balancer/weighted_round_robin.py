import asyncio
from typing import Optional, List
from .models import Node, ServiceGroup


class WeightedRoundRobin:
    def __init__(self):
        self._lock = asyncio.Lock()

    async def select_node(
        self, service_group: ServiceGroup, exclude_urls: List[str] = None
    ) -> Optional[Node]:
        exclude_urls = exclude_urls or []
        nodes = await service_group.get_available_nodes()
        nodes = [n for n in nodes if n.url not in exclude_urls]
        
        if not nodes:
            return None
        
        async with self._lock:
            return self._select(nodes)

    def _select(self, nodes: List[Node]) -> Node:
        total_weight = sum(n.effective_weight for n in nodes if n.effective_weight > 0)
        if total_weight <= 0:
            total_weight = sum(n.weight for n in nodes if n.weight > 0)
        
        best_node = None
        max_weight = -1

        for node in nodes:
            if node.effective_weight <= 0:
                continue
            
            node.current_weight += node.effective_weight
            
            if node.current_weight > max_weight:
                max_weight = node.current_weight
                best_node = node

        if best_node is not None:
            best_node.current_weight -= total_weight
        
        return best_node

    def reset_weights(self, nodes: List[Node]):
        for node in nodes:
            node.current_weight = 0
