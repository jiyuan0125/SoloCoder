from datetime import datetime
from typing import List, Optional

from shared.models import Order, OrderItem, OrderStatus
from server.app.exceptions import (
    OrderNotFoundException,
    OrderAlreadyShippedException,
    DuplicateOrderIdException,
)
from server.app.repositories.memory_store import memory_store


def get_order(order_id: str) -> Order:
    order = memory_store.get_order(order_id)
    if order is None:
        raise OrderNotFoundException(order_id)
    return order


def get_all_orders() -> List[Order]:
    return memory_store.get_all_orders()


def create_order(order_id: str, items: List[OrderItem]) -> Order:
    existing = memory_store.get_order(order_id)
    if existing is not None:
        raise DuplicateOrderIdException(order_id)

    for item in items:
        memory_store.save_product(item.product)

    order = Order(
        order_id=order_id,
        items=items,
        status=OrderStatus.CREATED,
    )
    memory_store.save_order(order)
    return order


def update_order(order_id: str, items: Optional[List[OrderItem]] = None) -> Order:
    order = get_order(order_id)

    if order.status == OrderStatus.SHIPPED:
        raise OrderAlreadyShippedException(order_id)

    if items is not None:
        old_quantities: dict[str, int] = {}
        for item in order.items:
            old_quantities[item.product.product_id] = item.quantity

        new_quantities: dict[str, int] = {}
        for item in items:
            new_quantities[item.product.product_id] = item.quantity
            memory_store.save_product(item.product)

        needs_recheck = False
        for prod_id, old_qty in old_quantities.items():
            new_qty = new_quantities.get(prod_id, 0)
            if new_qty < old_qty:
                needs_recheck = True
                break

        order.items = items
        order.status = OrderStatus.MODIFIED
        order.modified_at = datetime.now()

        if needs_recheck:
            _recheck_packed_packages(order_id)

        memory_store.save_order(order)

    return order


def _recheck_packed_packages(order_id: str) -> None:
    from shared.models import PackageStatus
    from server.app.services.packing_service import validate_package

    packing_list = memory_store.get_packing_list(order_id)
    if packing_list is None:
        return

    for package in packing_list.packages:
        if package.status == PackageStatus.SHIPPED:
            continue

        try:
            validate_package(package)
        except Exception:
            package.status = PackageStatus.NEED_REPACK
            memory_store.save_package(package)


def get_order_product(order_id: str, product_id: str) -> Optional[OrderItem]:
    order = get_order(order_id)
    for item in order.items:
        if item.product.product_id == product_id:
            return item
    return None
