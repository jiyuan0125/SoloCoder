from datetime import datetime
from enum import Enum
from typing import Optional
from pydantic import BaseModel


class RiderStatus(str, Enum):
    IDLE = "idle"
    DELIVERING = "delivering"
    RESTING = "resting"
    OFFLINE = "offline"


class OrderStatus(str, Enum):
    PENDING = "pending"
    ASSIGNED = "assigned"
    DELIVERING = "delivering"
    COMPLETED = "completed"
    TIMED_OUT = "timed_out"


class Rider(BaseModel):
    id: str
    status: RiderStatus = RiderStatus.IDLE
    last_position_update: datetime
    current_order_id: Optional[str] = None


class Order(BaseModel):
    id: str
    created_at: datetime
    expected_delivery_at: datetime
    status: OrderStatus = OrderStatus.PENDING
    assigned_rider_id: Optional[str] = None
    accepted_at: Optional[datetime] = None
    completed_at: Optional[datetime] = None
    needs_manual_reassign: bool = False


class Metrics(BaseModel):
    online_riders: int
    pending_orders: int
    delivering_orders: int
    today_completed: int
    avg_delivery_time_minutes: float
    timeout_rate: float
