from datetime import datetime, time
from typing import List
from .state import SystemState
from .models import RiderStatus, OrderStatus, Order


class MetricsCalculator:
    def __init__(self, state: SystemState):
        self.state = state

    def get_online_riders_count(self) -> int:
        return sum(
            1 for rider in self.state.get_all_riders()
            if rider.status != RiderStatus.OFFLINE
        )

    def get_pending_orders_count(self) -> int:
        return len(self.state.get_orders_by_status(OrderStatus.PENDING))

    def get_delivering_orders_count(self) -> int:
        return len(self.state.get_orders_by_status(OrderStatus.DELIVERING))

    def get_today_completed_orders_count(self) -> int:
        today = datetime.now().date()
        return sum(
            1 for order in self.state.get_orders_by_status(OrderStatus.COMPLETED)
            if order.completed_at and order.completed_at.date() == today
        )

    def get_today_completed_orders(self) -> List[Order]:
        today = datetime.now().date()
        return [
            order for order in self.state.get_orders_by_status(OrderStatus.COMPLETED)
            if order.completed_at and order.completed_at.date() == today
        ]

    def get_average_delivery_duration_minutes(self) -> float:
        today_orders = self.get_today_completed_orders()
        if not today_orders:
            return 0.0
        durations = [
            order.delivery_duration_minutes
            for order in today_orders
            if order.delivery_duration_minutes is not None
        ]
        if not durations:
            return 0.0
        return sum(durations) / len(durations)

    def get_timeout_rate(self) -> float:
        today = datetime.now().date()
        today_orders = [
            order for order in self.state.get_all_orders()
            if order.created_at.date() == today
        ]
        if not today_orders:
            return 0.0
        timed_out = sum(
            1 for order in today_orders
            if order.status == OrderStatus.TIMEOUT
        )
        return timed_out / len(today_orders)

    def get_all_metrics(self) -> dict:
        return {
            "online_riders": self.get_online_riders_count(),
            "pending_orders": self.get_pending_orders_count(),
            "delivering_orders": self.get_delivering_orders_count(),
            "today_completed": self.get_today_completed_orders_count(),
            "avg_delivery_duration_minutes": round(self.get_average_delivery_duration_minutes(), 2),
            "timeout_rate": round(self.get_timeout_rate(), 4),
        }
