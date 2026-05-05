from typing import Dict, List, Optional
from uuid import UUID

from shared.models import (
    Order,
    Package,
    PackingList,
    BoxType,
    PackagingMaterialInventory,
    Product,
)


class MemoryStore:
    _instance: Optional["MemoryStore"] = None

    def __new__(cls) -> "MemoryStore":
        if cls._instance is None:
            cls._instance = super().__new__(cls)
            cls._instance._initialized = False
        return cls._instance

    def __init__(self) -> None:
        if self._initialized:
            return
        self._initialized: bool = True

        self.orders: Dict[str, Order] = {}
        self.packages: Dict[UUID, Package] = {}
        self.packing_lists: Dict[str, PackingList] = {}
        self.packing_templates: Dict[UUID, PackingList] = {}
        self.box_types: Dict[str, BoxType] = {}
        self.inventories: Dict[str, PackagingMaterialInventory] = {}
        self.products: Dict[str, Product] = {}

    def get_order(self, order_id: str) -> Optional[Order]:
        return self.orders.get(order_id)

    def save_order(self, order: Order) -> None:
        self.orders[order.order_id] = order

    def get_all_orders(self) -> List[Order]:
        return list(self.orders.values())

    def delete_order(self, order_id: str) -> None:
        self.orders.pop(order_id, None)

    def get_package(self, package_id: UUID) -> Optional[Package]:
        return self.packages.get(package_id)

    def save_package(self, package: Package) -> None:
        self.packages[package.package_id] = package

    def get_packages_by_order(self, order_id: str) -> List[Package]:
        packing_list = self.packing_lists.get(order_id)
        if packing_list:
            return packing_list.packages
        return []

    def get_packing_list(self, order_id: str) -> Optional[PackingList]:
        return self.packing_lists.get(order_id)

    def save_packing_list(self, packing_list: PackingList) -> None:
        self.packing_lists[packing_list.order_id] = packing_list

    def get_packing_template(self, template_id: UUID) -> Optional[PackingList]:
        return self.packing_templates.get(template_id)

    def save_packing_template(self, template: PackingList) -> None:
        self.packing_templates[template.packing_id] = template

    def get_all_templates(self) -> List[PackingList]:
        return list(self.packing_templates.values())

    def get_box_type(self, box_type_id: str) -> Optional[BoxType]:
        return self.box_types.get(box_type_id)

    def save_box_type(self, box_type: BoxType) -> None:
        self.box_types[box_type.box_type_id] = box_type

    def get_all_box_types(self) -> List[BoxType]:
        return list(self.box_types.values())

    def get_inventory(self, box_type_id: str) -> Optional[PackagingMaterialInventory]:
        return self.inventories.get(box_type_id)

    def save_inventory(self, inventory: PackagingMaterialInventory) -> None:
        self.inventories[inventory.box_type.box_type_id] = inventory

    def get_all_inventories(self) -> List[PackagingMaterialInventory]:
        return list(self.inventories.values())

    def get_product(self, product_id: str) -> Optional[Product]:
        return self.products.get(product_id)

    def save_product(self, product: Product) -> None:
        self.products[product.product_id] = product


memory_store = MemoryStore()
