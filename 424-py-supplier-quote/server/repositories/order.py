from typing import Optional
from uuid import UUID

from shared.models.enums import OrderStatus
from shared.models.order import PurchaseOrder
from server.repositories.base import BaseRepository


class OrderRepository(BaseRepository[PurchaseOrder]):
    def get_by_order_number(self, order_number: str) -> Optional[PurchaseOrder]:
        return self.find_one(lambda o: o.order_number == order_number)

    def get_by_purchase(self, purchase_id: UUID) -> Optional[PurchaseOrder]:
        return self.find_one(lambda o: o.purchase_id == purchase_id)

    def list_by_supplier(self, supplier_id: UUID) -> list[PurchaseOrder]:
        return sorted(
            self.find(lambda o: o.supplier_id == supplier_id),
            key=lambda o: o.created_at,
            reverse=True,
        )

    def list_by_status(self, status: OrderStatus) -> list[PurchaseOrder]:
        return sorted(
            self.find(lambda o: o.status == status),
            key=lambda o: o.created_at,
            reverse=True,
        )

    def list_active(self) -> list[PurchaseOrder]:
        return self.find(
            lambda o: o.status in (OrderStatus.DRAFT, OrderStatus.CONFIRMED, OrderStatus.IN_PROGRESS)
        )
