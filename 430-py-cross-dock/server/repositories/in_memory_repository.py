from typing import Dict, List, Optional
from datetime import datetime

from shared.models import CrossDockOrder


class InMemoryRepository:
    def __init__(self) -> None:
        self._orders: Dict[str, CrossDockOrder] = {}
        self._order_number_to_id: Dict[str, str] = {}
        self._outbound_to_inbounds: Dict[str, List[str]] = {}

    def save(self, order: CrossDockOrder) -> CrossDockOrder:
        self._orders[order.id] = order
        self._order_number_to_id[order.order_number] = order.id
        if order.outbound_order_number:
            if order.outbound_order_number not in self._outbound_to_inbounds:
                self._outbound_to_inbounds[order.outbound_order_number] = []
            if order.id not in self._outbound_to_inbounds[order.outbound_order_number]:
                self._outbound_to_inbounds[order.outbound_order_number].append(order.id)
        return order

    def get_by_id(self, order_id: str) -> Optional[CrossDockOrder]:
        return self._orders.get(order_id)

    def get_by_order_number(self, order_number: str) -> Optional[CrossDockOrder]:
        order_id = self._order_number_to_id.get(order_number)
        if order_id:
            return self._orders.get(order_id)
        return None

    def get_all(self) -> List[CrossDockOrder]:
        return list(self._orders.values())

    def get_by_outbound_order_number(self, outbound_order_number: str) -> List[CrossDockOrder]:
        order_ids = self._outbound_to_inbounds.get(outbound_order_number, [])
        return [self._orders[oid] for oid in order_ids if oid in self._orders]

    def delete(self, order_id: str) -> bool:
        if order_id in self._orders:
            order = self._orders.pop(order_id)
            self._order_number_to_id.pop(order.order_number, None)
            if order.outbound_order_number:
                if order.outbound_order_number in self._outbound_to_inbounds:
                    self._outbound_to_inbounds[order.outbound_order_number].remove(order.id)
                    if not self._outbound_to_inbounds[order.outbound_order_number]:
                        del self._outbound_to_inbounds[order.outbound_order_number]
            return True
        return False

    def exists_by_order_number(self, order_number: str) -> bool:
        return order_number in self._order_number_to_id

    def exists_outbound_order(self, outbound_order_number: str) -> bool:
        return outbound_order_number in self._outbound_to_inbounds
