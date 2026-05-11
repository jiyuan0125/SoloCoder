from datetime import datetime
from enum import Enum
from typing import Optional
from pydantic import BaseModel, Field


class RiderStatus(str, Enum):
    IDLE = "idle"
    DELIVERING = "delivering"
    RESTING = "resting"
    OFFLINE = "offline"


class OrderStatus(str, Enum):
    PENDING = "pending"
    DELIVERING = "delivering"
    COMPLETED = "completed"
    TIMEOUT = "timeout"


class Rider(BaseModel):
    id: str
    status: RiderStatus = RiderStatus.IDLE
    current_order_id: Optional[str] = None
    last_position_update: datetime = Field(default_factory=datetime.now)
    created_at: datetime = Field(default_factory=datetime.now)


class Order(BaseModel):
    id: str
    merchant_name: str
    status: OrderStatus = OrderStatus.PENDING
    rider_id: Optional[str] = None
    created_at: datetime = Field(default_factory=datetime.now)
    accepted_at: Optional[datetime] = None
    completed_at: Optional[datetime] = None
    timeout_at: datetime = Field(default_factory=lambda: datetime.now())

    @property
    def is_timeout(self) -> bool:
        if self.status in (OrderStatus.COMPLETED, OrderStatus.TIMEOUT):
            return self.status == OrderStatus.TIMEOUT
        return datetime.now() >= self.timeout_at

    @property
    def delivery_duration_minutes(self) -> Optional[float]:
        if self.accepted_at and self.completed_at:
            return (self.completed_at - self.accepted_at).total_seconds() / 60
        return None
