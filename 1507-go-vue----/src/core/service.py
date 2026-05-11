from datetime import datetime, timedelta
from typing import List, Optional
from uuid import uuid4

from .models import Rider, Order, RiderStatus, OrderStatus, Metrics
from .store import store


EXPECTED_DELIVERY_MINUTES = 30
OFFLINE_TIMEOUT_MINUTES = 10


def create_order() -> Order:
    now = datetime.now()
    order = Order(
        id=str(uuid4()),
        created_at=now,
        expected_delivery_at=now + timedelta(minutes=EXPECTED_DELIVERY_MINUTES),
        status=OrderStatus.PENDING
    )
    return store.add_order(order)


def create_rider(rider_id: str) -> Rider:
    return store.add_rider(rider_id)


def update_rider_position(rider_id: str) -> Optional[Rider]:
    rider = store.get_rider(rider_id)
    if not rider:
        return None
    rider.last_position_update = datetime.now()
    return store.update_rider(rider)


def set_rider_status(rider_id: str, status: RiderStatus) -> Optional[Rider]:
    rider = store.get_rider(rider_id)
    if not rider:
        return None
    if status == RiderStatus.DELIVERING and rider.current_order_id:
        rider.status = RiderStatus.DELIVERING
    elif status == RiderStatus.IDLE or status == RiderStatus.RESTING:
        if rider.status == RiderStatus.OFFLINE:
            rider.last_position_update = datetime.now()
        rider.status = status
    rider.last_position_update = datetime.now()
    return store.update_rider(rider)


def accept_order(rider_id: str, order_id: str) -> Optional[Order]:
    order = store.get_order(order_id)
    rider = store.get_rider(rider_id)
    
    if not order or not rider:
        return None
    
    if order.status != OrderStatus.ASSIGNED:
        return None
    
    if order.assigned_rider_id != rider_id:
        return None
    
    if rider.status != RiderStatus.IDLE or rider.current_order_id:
        return None
    
    order.status = OrderStatus.DELIVERING
    order.accepted_at = datetime.now()
    store.update_order(order)
    
    rider.status = RiderStatus.DELIVERING
    rider.current_order_id = order_id
    store.update_rider(rider)
    
    return order


def complete_order(order_id: str) -> Optional[Order]:
    order = store.get_order(order_id)
    if not order:
        return None
    
    if order.status != OrderStatus.DELIVERING:
        return None
    
    order.status = OrderStatus.COMPLETED
    order.completed_at = datetime.now()
    store.update_order(order)
    
    if order.assigned_rider_id:
        rider = store.get_rider(order.assigned_rider_id)
        if rider:
            rider.status = RiderStatus.IDLE
            rider.current_order_id = None
            store.update_rider(rider)
    
    return order


def manually_reassign_order(order_id: str, rider_id: str) -> Optional[Order]:
    order = store.get_order(order_id)
    rider = store.get_rider(rider_id)
    
    if not order or not rider:
        return None
    
    if rider.status != RiderStatus.IDLE or rider.current_order_id:
        return None
    
    if order.assigned_rider_id:
        old_rider = store.get_rider(order.assigned_rider_id)
        if old_rider and old_rider.current_order_id == order_id:
            old_rider.current_order_id = None
            if old_rider.status == RiderStatus.DELIVERING:
                old_rider.status = RiderStatus.IDLE
            store.update_rider(old_rider)
    
    order.assigned_rider_id = rider_id
    order.status = OrderStatus.ASSIGNED
    order.needs_manual_reassign = False
    store.update_order(order)
    
    return order


def dispatch_orders() -> List[Order]:
    assigned_orders = []
    
    pending_orders = sorted(
        store.get_orders_by_status(OrderStatus.PENDING),
        key=lambda o: o.created_at
    )
    
    idle_riders = store.get_riders_by_status(RiderStatus.IDLE)
    idle_riders_without_order = [
        r for r in idle_riders if not r.current_order_id
    ]
    
    for order in pending_orders:
        if not idle_riders_without_order:
            break
        
        rider = idle_riders_without_order.pop(0)
        
        order.assigned_rider_id = rider.id
        order.status = OrderStatus.ASSIGNED
        store.update_order(order)
        assigned_orders.append(order)
    
    return assigned_orders


def check_timeouts() -> List[Order]:
    now = datetime.now()
    timed_out_orders = []
    
    orders_to_check = [
        o for o in store.get_all_orders()
        if o.status in [OrderStatus.PENDING, OrderStatus.ASSIGNED, OrderStatus.DELIVERING]
    ]
    
    for order in orders_to_check:
        if now >= order.expected_delivery_at:
            order.status = OrderStatus.TIMED_OUT
            store.update_order(order)
            timed_out_orders.append(order)
            
            if order.assigned_rider_id:
                rider = store.get_rider(order.assigned_rider_id)
                if rider and rider.current_order_id == order.id:
                    rider.status = RiderStatus.IDLE
                    rider.current_order_id = None
                    store.update_rider(rider)
    
    return timed_out_orders


def check_offline_riders() -> List[Rider]:
    now = datetime.now()
    offline_riders = []
    
    for rider in store.get_all_riders():
        if rider.status == RiderStatus.OFFLINE:
            continue
        
        time_since_update = now - rider.last_position_update
        if time_since_update >= timedelta(minutes=OFFLINE_TIMEOUT_MINUTES):
            rider.status = RiderStatus.OFFLINE
            
            if rider.current_order_id:
                order = store.get_order(rider.current_order_id)
                if order and order.status in [OrderStatus.ASSIGNED, OrderStatus.DELIVERING]:
                    order.status = OrderStatus.PENDING
                    order.needs_manual_reassign = True
                    order.assigned_rider_id = None
                    order.accepted_at = None
                    store.update_order(order)
                
                rider.current_order_id = None
            
            store.update_rider(rider)
            offline_riders.append(rider)
    
    return offline_riders


def get_metrics() -> Metrics:
    all_riders = store.get_all_riders()
    all_orders = store.get_all_orders()
    
    online_riders = sum(1 for r in all_riders if r.status != RiderStatus.OFFLINE)
    
    pending_orders = sum(1 for o in all_orders if o.status == OrderStatus.PENDING)
    
    delivering_orders = sum(1 for o in all_orders if o.status == OrderStatus.DELIVERING)
    
    today = datetime.now().date()
    today_completed = sum(
        1 for o in all_orders
        if o.status == OrderStatus.COMPLETED
        and o.completed_at
        and o.completed_at.date() == today
    )
    
    completed_orders = [
        o for o in all_orders
        if o.status == OrderStatus.COMPLETED
        and o.accepted_at
        and o.completed_at
    ]
    
    if completed_orders:
        total_time = sum(
            (o.completed_at - o.accepted_at).total_seconds()
            for o in completed_orders
        )
        avg_delivery_time_seconds = total_time / len(completed_orders)
        avg_delivery_time_minutes = avg_delivery_time_seconds / 60
    else:
        avg_delivery_time_minutes = 0.0
    
    orders_ever_assigned = [
        o for o in all_orders
        if o.status in [OrderStatus.COMPLETED, OrderStatus.TIMED_OUT]
    ]
    
    if orders_ever_assigned:
        timed_out_count = sum(1 for o in orders_ever_assigned if o.status == OrderStatus.TIMED_OUT)
        timeout_rate = timed_out_count / len(orders_ever_assigned)
    else:
        timeout_rate = 0.0
    
    return Metrics(
        online_riders=online_riders,
        pending_orders=pending_orders,
        delivering_orders=delivering_orders,
        today_completed=today_completed,
        avg_delivery_time_minutes=round(avg_delivery_time_minutes, 2),
        timeout_rate=round(timeout_rate, 4)
    )


def get_rider(rider_id: str) -> Optional[Rider]:
    return store.get_rider(rider_id)


def get_all_riders() -> List[Rider]:
    return store.get_all_riders()


def get_order(order_id: str) -> Optional[Order]:
    return store.get_order(order_id)


def get_all_orders() -> List[Order]:
    return store.get_all_orders()
