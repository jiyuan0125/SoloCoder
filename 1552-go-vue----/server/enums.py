from enum import Enum


class AgentServiceStage(str, Enum):
    ORDER_RECEIVED = "order_received"
    DECLARATION = "declaration"
    BERTHING_ARRANGEMENT = "berthing_arrangement"
    OPERATION_EXECUTION = "operation_execution"
    FEE_SETTLEMENT = "fee_settlement"
    DEPARTURE_CONFIRMED = "departure_confirmed"


class StageStatus(str, Enum):
    PENDING = "pending"
    COMPLETED = "completed"


class BerthStatus(str, Enum):
    AVAILABLE = "available"
    OCCUPIED = "occupied"


class BerthApplicationStatus(str, Enum):
    PENDING = "pending"
    APPROVED = "approved"
    REJECTED = "rejected"


class SupplyType(str, Enum):
    FUEL = "fuel"
    FRESH_WATER = "fresh_water"


class SupplyStatus(str, Enum):
    SCHEDULED = "scheduled"
    DELIVERED = "delivered"
    CANCELLED = "cancelled"


class WasteStatus(str, Enum):
    SCHEDULED = "scheduled"
    COMPLETED = "completed"
    CANCELLED = "cancelled"


class TodoStatus(str, Enum):
    PENDING = "pending"
    COMPLETED = "completed"


class SettlementStatus(str, Enum):
    PENDING = "pending"
    SETTLED = "settled"
