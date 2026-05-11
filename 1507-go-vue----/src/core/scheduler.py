from typing import List, Tuple
from .state import SystemState
from .models import OrderStatus


class Scheduler:
    def __init__(self, state: SystemState):
        self.state = state

    def check_timeouts(self) -> List[str]:
        timed_out_orders: List[str] = []
        for order in self.state.get_all_orders():
            if order.status in (OrderStatus.PENDING, OrderStatus.DELIVERING):
                if order.is_timeout:
                    self.state.mark_order_timeout(order.id)
                    timed_out_orders.append(order.id)
        return timed_out_orders

    def check_offline_riders(self) -> List[Tuple[str, List[str]]]:
        offline_events: List[Tuple[str, List[str]]] = []
        for rider_id in self.state.get_riders_needing_offline_check():
            affected_orders = self.state.handle_offline_rider(rider_id)
            offline_events.append((rider_id, affected_orders))
        return offline_events

    def run_dispatch(self) -> List[Tuple[str, str]]:
        assignments: List[Tuple[str, str]] = []
        pending_orders = self.state.get_pending_orders_sorted()
        idle_riders = self.state.get_idle_riders()

        for order in pending_orders:
            if not idle_riders:
                break
            rider = idle_riders.pop(0)
            self.state.assign_order_to_rider(order.id, rider.id)
            assignments.append((order.id, rider.id))

        return assignments

    def tick(self) -> dict:
        timed_out = self.check_timeouts()
        offline_events = self.check_offline_riders()
        assignments = self.run_dispatch()

        return {
            "timed_out_orders": timed_out,
            "offline_riders": [rider_id for rider_id, _ in offline_events],
            "assigned_orders": assignments,
        }
