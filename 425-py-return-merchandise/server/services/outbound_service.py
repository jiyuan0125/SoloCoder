from datetime import datetime
from typing import List, Optional

from shared.models import (
    OutboundOrder,
    OutboundOrderCreate,
    OrderStatus,
)
from shared.constants import ErrorCode
from server.data_store import DataStore, get_data_store


class OutboundService:
    def __init__(self, data_store: Optional[DataStore] = None) -> None:
        self._data_store = data_store or get_data_store()

    def create_order(self, request: OutboundOrderCreate) -> OutboundOrder:
        now = datetime.now()
        total_amount = sum(item.unit_price * item.quantity for item in request.items)

        order = OutboundOrder(
            order_id=request.order_id,
            customer_id=request.customer_id,
            items=request.items,
            total_amount=total_amount,
            status=OrderStatus.PENDING,
            created_at=now,
            updated_at=now,
        )
        self._data_store.create_outbound_order(order)
        return order

    def get_order(self, order_id: str) -> Optional[OutboundOrder]:
        return self._data_store.get_outbound_order(order_id)

    def list_orders(self) -> List[OutboundOrder]:
        return self._data_store.list_outbound_orders()

    def ship_order(self, order_id: str) -> OutboundOrder:
        order = self._data_store.get_outbound_order(order_id)
        if not order:
            raise ValueError(f"出库单 {order_id} 不存在")

        if order.status != OrderStatus.PENDING:
            raise ValueError(f"出库单 {order_id} 状态不是待发货，无法发货")

        now = datetime.now()
        order.status = OrderStatus.SHIPPED
        order.shipped_at = now
        order.updated_at = now
        return self._data_store.create_outbound_order(order)

    def deliver_order(self, order_id: str) -> OutboundOrder:
        order = self._data_store.get_outbound_order(order_id)
        if not order:
            raise ValueError(f"出库单 {order_id} 不存在")

        if order.status != OrderStatus.SHIPPED:
            raise ValueError(f"出库单 {order_id} 状态不是已发货，无法确认收货")

        now = datetime.now()
        order.status = OrderStatus.DELIVERED
        order.delivered_at = now
        order.updated_at = now
        return self._data_store.create_outbound_order(order)

    def complete_order(self, order_id: str) -> OutboundOrder:
        order = self._data_store.get_outbound_order(order_id)
        if not order:
            raise ValueError(f"出库单 {order_id} 不存在")

        if order.status != OrderStatus.DELIVERED:
            raise ValueError(f"出库单 {order_id} 状态不是已签收，无法完成")

        now = datetime.now()
        order.status = OrderStatus.COMPLETED
        order.updated_at = now
        return self._data_store.create_outbound_order(order)
