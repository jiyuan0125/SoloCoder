from typing import Dict, List, Optional
from datetime import datetime
from models import Node, NodeStats, Callback, NodeStatus
import uuid


class Store:
    def __init__(self):
        self._nodes: Dict[str, Node] = {}
        self._stats: Dict[str, NodeStats] = {}
        self._callbacks: Dict[str, Callback] = {}

    def add_node(self, address: str, weight: int) -> Node:
        node_id = str(uuid.uuid4())
        node = Node(id=node_id, address=address, weight=weight)
        self._nodes[node_id] = node
        self._stats[node_id] = NodeStats(node_id=node_id)
        return node

    def get_node(self, node_id: str) -> Optional[Node]:
        return self._nodes.get(node_id)

    def get_all_nodes(self) -> List[Node]:
        return list(self._nodes.values())

    def update_node(self, node_id: str, **kwargs) -> Optional[Node]:
        node = self._nodes.get(node_id)
        if node:
            for key, value in kwargs.items():
                if hasattr(node, key):
                    setattr(node, key, value)
            node.updated_at = datetime.now()
            return node
        return None

    def delete_node(self, node_id: str) -> bool:
        if node_id in self._nodes:
            del self._nodes[node_id]
            return True
        return False

    def get_stats(self, node_id: str) -> Optional[NodeStats]:
        return self._stats.get(node_id)

    def get_all_stats(self) -> List[NodeStats]:
        return list(self._stats.values())

    def record_request(self, node_id: str, response_time_ms: int):
        if node_id not in self._stats:
            self._stats[node_id] = NodeStats(node_id=node_id)
        self._stats[node_id].add_request(response_time_ms)

    def add_callback(self, url: str) -> Callback:
        callback_id = str(uuid.uuid4())
        callback = Callback(url=url)
        self._callbacks[callback_id] = callback
        return callback

    def get_all_callbacks(self) -> List[Callback]:
        return list(self._callbacks.values())


store = Store()
