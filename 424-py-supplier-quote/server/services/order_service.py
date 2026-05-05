from datetime import datetime, timedelta
from decimal import Decimal
from typing import Optional
from uuid import UUID

from shared.constants.error_codes import ErrorCode
from shared.models.enums import OrderStatus, PurchaseStatus
from shared.models.order import OrderItem, PurchaseOrder
from server.repositories.database import get_repository
from server.repositories.order import OrderRepository
from server.services.exceptions import BusinessException
from server.services.purchase_service import PurchaseService
from server.services.quote_service import QuoteService
from server.services.supplier_service import SupplierService


class OrderService:
    def __init__(self) -> None:
        self._order_repo = get_repository(OrderRepository)
        self._purchase_service = PurchaseService()
        self._quote_service = QuoteService()
        self._supplier_service = SupplierService()

    def create_order_from_quote(self, quote_id: UUID) -> PurchaseOrder:
        quote = self._quote_service.get_quote(quote_id)
        purchase = self._purchase_service.get_purchase(quote.purchase_id)
        supplier = self._supplier_service.get_supplier(quote.supplier_id)

        if quote.latest_version is None:
            raise BusinessException(ErrorCode.INVALID_INPUT, "报价无有效版本")

        if purchase.status == PurchaseStatus.AWARDED:
            existing_order = self._order_repo.get_by_purchase(purchase.id)
            if existing_order is not None:
                return existing_order

        items: list[OrderItem] = []
        for purchase_item in purchase.items:
            total_amount = purchase_item.quantity * quote.latest_version.unit_price
            order_item = OrderItem(
                product_name=purchase_item.product_name,
                product_code=purchase_item.product_code,
                quantity=purchase_item.quantity,
                unit=purchase_item.unit,
                unit_price=quote.latest_version.unit_price,
                total_amount=total_amount,
            )
            items.append(order_item)

        total_amount = sum((item.total_amount for item in items), Decimal("0"))
        order_number = self._generate_order_number()

        order = PurchaseOrder(
            order_number=order_number,
            purchase_id=purchase.id,
            quote_id=quote.id,
            supplier_id=supplier.id,
            items=items,
            total_amount=total_amount,
            status=OrderStatus.DRAFT,
            delivery_days=quote.latest_version.delivery_days,
            expected_delivery_date=datetime.utcnow() + timedelta(days=quote.latest_version.delivery_days),
        )

        self._purchase_service.mark_awarded(purchase.id, supplier.id)
        self._quote_service.mark_awarded(quote.id)

        all_quotes = self._quote_service.list_quotes_by_purchase(purchase.id)
        for q in all_quotes:
            if q.id != quote.id:
                self._quote_service.mark_lost(q.id)

        return self._order_repo.create(order)

    def _generate_order_number(self) -> str:
        now = datetime.utcnow()
        count = self._order_repo.count() + 1
        return f"PO-{now.strftime('%Y%m%d')}-{count:06d}"

    def get_order(self, order_id: UUID) -> PurchaseOrder:
        order = self._order_repo.get_by_id(order_id)
        if order is None:
            raise BusinessException(ErrorCode.ORDER_NOT_FOUND)
        return order

    def get_order_by_number(self, order_number: str) -> PurchaseOrder:
        order = self._order_repo.get_by_order_number(order_number)
        if order is None:
            raise BusinessException(ErrorCode.ORDER_NOT_FOUND)
        return order

    def confirm_order(self, order_id: UUID, remarks: str = "") -> PurchaseOrder:
        order = self.get_order(order_id)

        if order.status != OrderStatus.DRAFT:
            raise BusinessException(ErrorCode.ORDER_ALREADY_CONFIRMED)

        order.status = OrderStatus.CONFIRMED
        order.confirmed_at = datetime.utcnow()
        order.remarks = remarks if remarks else order.remarks
        order.updated_at = datetime.utcnow()

        return self._order_repo.update(order)

    def start_order(self, order_id: UUID) -> PurchaseOrder:
        order = self.get_order(order_id)

        if order.status != OrderStatus.CONFIRMED:
            raise BusinessException(ErrorCode.INVALID_INPUT, "只能确认已确认状态的订单")

        order.status = OrderStatus.IN_PROGRESS
        order.updated_at = datetime.utcnow()

        return self._order_repo.update(order)

    def complete_order(self, order_id: UUID) -> PurchaseOrder:
        order = self.get_order(order_id)

        if order.status != OrderStatus.IN_PROGRESS:
            raise BusinessException(ErrorCode.INVALID_INPUT, "只能完成进行中的订单")

        order.status = OrderStatus.COMPLETED
        order.completed_at = datetime.utcnow()
        order.updated_at = datetime.utcnow()

        return self._order_repo.update(order)

    def cancel_order(self, order_id: UUID) -> PurchaseOrder:
        order = self.get_order(order_id)

        if order.status in (OrderStatus.COMPLETED, OrderStatus.CANCELLED):
            return order

        order.status = OrderStatus.CANCELLED
        order.updated_at = datetime.utcnow()

        return self._order_repo.update(order)

    def list_orders_by_supplier(self, supplier_id: UUID) -> list[PurchaseOrder]:
        self._supplier_service.get_supplier(supplier_id)
        return self._order_repo.list_by_supplier(supplier_id)

    def list_orders(self, status: Optional[OrderStatus] = None) -> list[PurchaseOrder]:
        if status is not None:
            return self._order_repo.list_by_status(status)
        return self._order_repo.list_all()

    def list_active_orders(self) -> list[PurchaseOrder]:
        return self._order_repo.list_active()
