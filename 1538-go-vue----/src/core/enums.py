from enum import Enum


class WasteType(str, Enum):
    GENERAL = "general"
    HAZARDOUS = "hazardous"


class WaybillStatus(str, Enum):
    PENDING_SHIPMENT = "pending_shipment"
    IN_TRANSIT = "in_transit"
    ARRIVED = "arrived"
    DISPOSED = "disposed"
    REJECTED = "rejected"


class UnitType(str, Enum):
    PRODUCER = "producer"
    DISPOSER = "disposer"
