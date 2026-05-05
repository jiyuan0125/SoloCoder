from datetime import datetime
from typing import Dict, List, Optional, Union

from shared.models import (
    OutboundOrder,
    ScrapRecord,
    DefectiveItem,
    OrderStatus,
    ReturnStatus,
    OutboundOrderItem,
    ReturnOrderItem,
    DisposalReason,
)
from server.models import ReturnOrder
from shared.constants import ReturnConstants


class DataStore:
    _instance: Optional["DataStore"] = None
    _initialized: bool = False

    def __new__(cls) -> "DataStore":
        if cls._instance is None:
            cls._instance = super().__new__(cls)
            cls._instance._initialized = False
        return cls._instance

    def __init__(self) -> None:
        if self._initialized:
            return
        self._outbound_orders: Dict[str, OutboundOrder] = {}
        self._return_orders: Dict[str, ReturnOrder] = {}
        self._scrap_records: Dict[str, ScrapRecord] = {}
        self._defective_items: Dict[str, DefectiveItem] = {}
        self._normal_inventory: Dict[str, Dict[str, Union[int, float]]] = {}
        self._return_tracking: Dict[str, Dict[str, int]] = {}
        self._defective_id_counter: int = 1
        self._scrap_id_counter: int = 1
        self._initialized = True

    def get_outbound_order(self, order_id: str) -> Optional[OutboundOrder]:
        return self._outbound_orders.get(order_id)

    def create_outbound_order(self, order: OutboundOrder) -> OutboundOrder:
        self._outbound_orders[order.order_id] = order
        self._return_tracking[order.order_id] = {}
        for item in order.items:
            self._return_tracking[order.order_id][item.sku] = 0
        return order

    def list_outbound_orders(self) -> List[OutboundOrder]:
        return list(self._outbound_orders.values())

    def get_return_order(self, return_id: str) -> Optional[ReturnOrder]:
        return self._return_orders.get(return_id)

    def create_return_order(self, return_order: ReturnOrder) -> ReturnOrder:
        self._return_orders[return_order.return_order_id] = return_order
        return return_order

    def update_return_order(self, return_order: ReturnOrder) -> ReturnOrder:
        self._return_orders[return_order.return_order_id] = return_order
        return return_order

    def list_return_orders(self) -> List[ReturnOrder]:
        return list(self._return_orders.values())

    def get_returned_quantity(self, outbound_id: str, sku: str) -> int:
        tracking = self._return_tracking.get(outbound_id, {})
        return tracking.get(sku, 0)

    def add_returned_quantity(self, outbound_id: str, sku: str, quantity: int) -> None:
        if outbound_id not in self._return_tracking:
            self._return_tracking[outbound_id] = {}
        if sku not in self._return_tracking[outbound_id]:
            self._return_tracking[outbound_id][sku] = 0
        self._return_tracking[outbound_id][sku] += quantity

    def get_defective_item(self, defective_id: str) -> Optional[DefectiveItem]:
        return self._defective_items.get(defective_id)

    def create_defective_item(self, item: DefectiveItem) -> DefectiveItem:
        self._defective_items[item.defective_id] = item
        return item

    def update_defective_item(self, item: DefectiveItem) -> DefectiveItem:
        self._defective_items[item.defective_id] = item
        return item

    def list_defective_items(self) -> List[DefectiveItem]:
        return list(self._defective_items.values())

    def get_scrap_record(self, scrap_id: str) -> Optional[ScrapRecord]:
        return self._scrap_records.get(scrap_id)

    def create_scrap_record(self, record: ScrapRecord) -> ScrapRecord:
        self._scrap_records[record.scrap_id] = record
        return record

    def list_scrap_records(self) -> List[ScrapRecord]:
        return list(self._scrap_records.values())

    def add_normal_inventory(self, sku: str, quantity: int, price: float) -> None:
        if sku not in self._normal_inventory:
            self._normal_inventory[sku] = {"quantity": 0, "total_value": 0.0}
        self._normal_inventory[sku]["quantity"] += quantity
        self._normal_inventory[sku]["total_value"] += quantity * price

    def get_normal_inventory(self, sku: str) -> int:
        if sku not in self._normal_inventory:
            return 0
        return int(self._normal_inventory[sku].get("quantity", 0))

    def generate_defective_id(self) -> str:
        self._defective_id_counter += 1
        return f"DEF_{datetime.now().strftime('%Y%m%d')}_{self._defective_id_counter:04d}"

    def generate_scrap_id(self) -> str:
        self._scrap_id_counter += 1
        return f"SCR_{datetime.now().strftime('%Y%m%d')}_{self._scrap_id_counter:04d}"


def get_data_store() -> DataStore:
    return DataStore()
