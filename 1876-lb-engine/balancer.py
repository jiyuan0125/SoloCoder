from typing import List, Optional
from models import Node, NodeStatus
from store import store


class WeightedRoundRobinBalancer:
    def __init__(self):
        self._current_weight_map: dict = {}

    def _get_eligible_nodes(self) -> List[Node]:
        nodes = store.get_all_nodes()
        return [
            node for node in nodes
            if node.status == NodeStatus.ACTIVE and node.weight > 0
        ]

    def select_node(self) -> Optional[Node]:
        eligible_nodes = self._get_eligible_nodes()
        if not eligible_nodes:
            return None

        if len(eligible_nodes) == 1:
            return eligible_nodes[0]

        total_weight = sum(node.weight for node in eligible_nodes)
        max_weight = max(node.weight for node in eligible_nodes)

        for node_id in list(self._current_weight_map.keys()):
            if not any(n.id == node_id for n in eligible_nodes):
                del self._current_weight_map[node_id]

        while True:
            for node in eligible_nodes:
                if node.id not in self._current_weight_map:
                    self._current_weight_map[node.id] = 0
                self._current_weight_map[node.id] += node.weight

                if self._current_weight_map[node.id] >= max_weight:
                    self._current_weight_map[node.id] -= total_weight
                    return node

            max_weight = max(self._current_weight_map.values())


balancer = WeightedRoundRobinBalancer()
