from datetime import datetime, timedelta
from typing import Dict, List, Optional
from .models import Rider, Order, RiderStatus, OrderStatus


class SystemState:
    def __init__(self):
        self.riders: Dict[str, Rider] = {}
        self.orders: Dict[str, Order] = {}
        self.offline_riders_orders: Dict[str, List[str]] = {}

    def add_rider(self, rider_id: str) -> Rider:
        if rider_id in self.riders:
            raise ValueError(f"Rider {rider_id} already exists")
        rider = Rider(id=rider_id)
        self.riders[rider_id] = rider
        return rider

    def get_rider(self, rider_id: str) -> Optional[Rider]:
        return self.riders.get(rider_id)

    def get_all_riders(self) -> List[Rider]:
        return list(self.riders.values())

    def update_rider_status(self, rider_id: str, status: RiderStatus) -> Rider:
        rider = self.get_rider(rider_id)
        if not rider:
            raise ValueError(f"Rider {rider_id} not found")
        rider.status = status
        return rider

    def update_rider_position(self, rider_id: str) -> Rider:
        rider = self.get_rider(rider_id)
        if not rider:
            raise ValueError(f"Rider {rider_id} not found")
        rider.last_position_update = datetime.now()
        return rider

    def add_order(self, order_id: str, merchant_name: str) -> Order:
        if order_id in self.orders:
            raise ValueError(f"Order {order_id} already exists")
        now = datetime.now()
        order = Order(
            id=order_id,
            merchant_name=merchant_name,
            created_at=now,
            timeout_at=now + timedelta(minutes=30),
        )
        self.orders[order_id] = order
        return order

    def get_order(self, order_id: str) -> Optional[Order]:
        return self.orders.get(order_id)

    def get_all_orders(self) -> List[Order]:
        return list(self.orders.values())

    def get_orders_by_status(self, status: OrderStatus) -> List[Order]:
        return [order for order in self.orders.values() if order.status == status]

    def get_pending_orders_sorted(self) -> List[Order]:
        pending = self.get_orders_by_status(OrderStatus.PENDING)
        return sorted(pending, key=lambda o: o.created_at)

    def get_idle_riders(self) -> List[Rider]:
        return [
            rider
            for rider in self.riders.values()
            if rider.status == RiderStatus.IDLE and rider.current_order_id is None
        ]

    def assign_order_to_rider(self, order_id: str, rider_id: str) -> None:
        order = self.get_order(order_id)
        rider = self.get_rider(rider_id)
        if not order or not rider:
            raise ValueError("Order or rider not found")
        if order.status != OrderStatus.PENDING:
            raise ValueError("Order is not pending")
        if rider.status != RiderStatus.IDLE or rider.current_order_id is not None:
            raise ValueError("Rider is not available")

        order.status = OrderStatus.DELIVERING
        order.rider_id = rider_id
        order.accepted_at = datetime.now()

        rider.status = RiderStatus.DELIVERING
        rider.current_order_id = order_id

    def complete_order(self, order_id: str) -> None:
        order = self.get_order(order_id)
        if not order:
            raise ValueError(f"Order {order_id} not found")
        if order.status != OrderStatus.DELIVERING:
            raise ValueError("Order is not delivering")

        order.status = OrderStatus.COMPLETED
        order.completed_at = datetime.now()

        if order.rider_id:
            rider = self.get_rider(order.rider_id)
            if rider:
                rider.status = RiderStatus.IDLE
                rider.current_order_id = None

    def mark_order_timeout(self, order_id: str) -> None:
        order = self.get_order(order_id)
        if not order:
            raise ValueError(f"Order {order_id} not found")
        if order.status in (OrderStatus.COMPLETED, OrderStatus.TIMEOUT):
            return

        order.status = OrderStatus.TIMEOUT

        if order.rider_id:
            rider = self.get_rider(order.rider_id)
            if rider:
                rider.status = RiderStatus.IDLE
                rider.current_order_id = None

    def handle_offline_rider(self, rider_id: str) -> List[str]:
        rider = self.get_rider(rider_id)
        if not rider:
            raise ValueError(f"Rider {rider_id} not found")
        if rider.status == RiderStatus.OFFLINE:
            return []

        rider.status = RiderStatus.OFFLINE
        affected_orders: List[str] = []

        if rider.current_order_id:
            order = self.get_order(rider.current_order_id)
            if order and order.status == OrderStatus.DELIVERING:
                order.status = OrderStatus.PENDING
                order.rider_id = None
                order.accepted_at = None
                affected_orders.append(rider.current_order_id)

            if rider_id not in self.offline_riders_orders:
                self.offline_riders_orders[rider_id] = []
            self.offline_riders_orders[rider_id].extend(affected_orders)

        rider.current_order_id = None
        return affected_orders

    def handle_rider_back_online(self, rider_id: str) -> None:
        rider = self.get_rider(rider_id)
        if not rider:
            raise ValueError(f"Rider {rider_id} not found")
        if rider.status != RiderStatus.OFFLINE:
            return

        rider.status = RiderStatus.IDLE
        rider.last_position_update = datetime.now()

    def get_riders_needing_offline_check(self, timeout_minutes: int = 10) -> List[str]:
        now = datetime.now()
        return [
            rider_id
            for rider_id, rider in self.riders.items()
            if rider.status != RiderStatus.OFFLINE
            and (now - rider.last_position_update).total_seconds() / 60 >= timeout_minutes
        ]
