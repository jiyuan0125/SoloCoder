from datetime import datetime
from typing import Dict, List, Optional
from threading import Lock

from .models import Rider, Order, RiderStatus, OrderStatus


class Store:
    def __init__(self):
        self._riders: Dict[str, Rider] = {}
        self._orders: Dict[str, Order] = {}
        self._lock = Lock()

    def add_rider(self, rider_id: str) -> Rider:
        with self._lock:
            rider = Rider(
                id=rider_id,
                status=RiderStatus.IDLE,
                last_position_update=datetime.now()
            )
            self._riders[rider_id] = rider
            return rider.model_copy()

    def get_rider(self, rider_id: str) -> Optional[Rider]:
        with self._lock:
            rider = self._riders.get(rider_id)
            return rider.model_copy() if rider else None

    def get_all_riders(self) -> List[Rider]:
        with self._lock:
            return [r.model_copy() for r in self._riders.values()]

    def update_rider(self, rider: Rider) -> Optional[Rider]:
        with self._lock:
            if rider.id not in self._riders:
                return None
            self._riders[rider.id] = rider.model_copy()
            return self._riders[rider.id].model_copy()

    def add_order(self, order: Order) -> Order:
        with self._lock:
            self._orders[order.id] = order
            return order.model_copy()

    def get_order(self, order_id: str) -> Optional[Order]:
        with self._lock:
            order = self._orders.get(order_id)
            return order.model_copy() if order else None

    def get_all_orders(self) -> List[Order]:
        with self._lock:
            return [o.model_copy() for o in self._orders.values()]

    def update_order(self, order: Order) -> Optional[Order]:
        with self._lock:
            if order.id not in self._orders:
                return None
            self._orders[order.id] = order.model_copy()
            return self._orders[order.id].model_copy()

    def get_orders_by_status(self, status: OrderStatus) -> List[Order]:
        with self._lock:
            return [
                o.model_copy()
                for o in self._orders.values()
                if o.status == status
            ]

    def get_riders_by_status(self, status: RiderStatus) -> List[Rider]:
        with self._lock:
            return [
                r.model_copy()
                for r in self._riders.values()
                if r.status == status
            ]


store = Store()
