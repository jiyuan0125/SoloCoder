from datetime import datetime
from typing import List, Optional
from uuid import UUID

from shared import constants
from shared.models import (
    Package,
    PackageItem,
    PackingList,
    BoxType,
    PackageStatus,
    ProductType,
    OrderStatus,
    Product,
)
from server.app.exceptions import (
    WeightExceedsLimitException,
    VolumeExceedsLimitException,
    HeavyItemMustBeAloneException,
    FragileMustBeOnBottomException,
    VolumeUtilizationExceededException,
    PackageAlreadyShippedException,
    InsufficientStockException,
    InvalidOperationException,
)
from server.app.repositories.memory_store import memory_store
from server.app.services import order_service


def validate_package(package: Package) -> None:
    total_weight = package.total_weight_kg
    total_volume = package.total_volume_m3
    effective_volume = package.box_type.effective_volume_m3

    _check_heavy_items_must_be_alone(package)

    if total_weight > constants.MAX_PACKAGE_WEIGHT_KG:
        raise WeightExceedsLimitException(total_weight, constants.MAX_PACKAGE_WEIGHT_KG)

    if total_volume > constants.MAX_PACKAGE_VOLUME_M3:
        raise VolumeExceedsLimitException(total_volume, constants.MAX_PACKAGE_VOLUME_M3)

    if total_volume > effective_volume:
        utilization = total_volume / package.box_type.volume_m3
        raise VolumeUtilizationExceededException(
            utilization, constants.VOLUME_UTILIZATION_RATE
        )

    has_fragile_or_liquid = False
    for item in package.items:
        if item.product.product_type in (ProductType.FRAGILE, ProductType.LIQUID):
            has_fragile_or_liquid = True
            if item.position_layer != 1:
                raise FragileMustBeOnBottomException(item.product.name)

    if has_fragile_or_liquid:
        package.has_fragile_label = True


def _check_heavy_items_must_be_alone(package: Package) -> None:
    heavy_items: list[PackageItem] = []
    for item in package.items:
        if item.product.weight_kg > constants.MAX_PACKAGE_WEIGHT_KG:
            heavy_items.append(item)

    if not heavy_items:
        return

    if len(heavy_items) > 1:
        raise HeavyItemMustBeAloneException(
            heavy_items[0].product.name,
            heavy_items[0].product.weight_kg,
        )

    if len(package.items) > 1:
        raise HeavyItemMustBeAloneException(
            heavy_items[0].product.name,
            heavy_items[0].product.weight_kg,
        )


def create_package(order_id: str, box_type_id: str) -> Package:
    order = order_service.get_order(order_id)

    if order.status == OrderStatus.SHIPPED:
        raise InvalidOperationException("订单已发货，无法创建新包裹")

    box_type = memory_store.get_box_type(box_type_id)
    if box_type is None:
        from server.app.exceptions import BoxTypeNotFoundException

        raise BoxTypeNotFoundException(box_type_id)

    inventory = memory_store.get_inventory(box_type_id)
    if inventory is None or inventory.stock_quantity <= 0:
        raise InsufficientStockException(box_type.name)

    packing_list = memory_store.get_packing_list(order_id)
    if packing_list is None:
        packing_list = PackingList(order_id=order_id)

    box_number = f"BOX-{order_id}-{len(packing_list.packages) + 1:03d}"

    package = Package(
        box_number=box_number,
        box_type=box_type,
        items=[],
        status=PackageStatus.PENDING,
    )

    packing_list.packages.append(package)

    inventory.stock_quantity -= 1
    inventory.used_quantity += 1

    memory_store.save_package(package)
    memory_store.save_packing_list(packing_list)
    memory_store.save_inventory(inventory)

    return package


def get_package(package_id: UUID) -> Package:
    package = memory_store.get_package(package_id)
    if package is None:
        from server.app.exceptions import PackageNotFoundException

        raise PackageNotFoundException(str(package_id))
    return package


def get_packages_by_order(order_id: str) -> List[Package]:
    return memory_store.get_packages_by_order(order_id)


def add_item_to_package(
    package_id: UUID, product_id: str, quantity: int
) -> Package:
    package = get_package(package_id)

    if package.status == PackageStatus.SHIPPED:
        raise PackageAlreadyShippedException(str(package_id))

    product = memory_store.get_product(product_id)
    if product is None:
        from server.app.exceptions import ProductNotFoundException

        raise ProductNotFoundException(product_id)

    temp_items = _create_updated_items_list(package, product, quantity)

    temp_package = Package(
        package_id=package.package_id,
        box_number=package.box_number,
        box_type=package.box_type,
        items=temp_items,
        status=package.status,
        tracking_number=package.tracking_number,
        has_fragile_label=package.has_fragile_label,
        shipped_at=package.shipped_at,
        created_at=package.created_at,
    )

    validate_package(temp_package)

    package.items = temp_items
    package.has_fragile_label = temp_package.has_fragile_label

    memory_store.save_package(package)

    return package


def _create_updated_items_list(
    package: Package, product: Product, quantity: int
) -> List[PackageItem]:
    updated_items: list[PackageItem] = [
        PackageItem(
            product=item.product,
            quantity=item.quantity,
            position_layer=item.position_layer,
        )
        for item in package.items
    ]

    found = False
    for item in updated_items:
        if item.product.product_id == product.product_id:
            item.quantity += quantity
            found = True
            break

    if not found:
        layer = 1
        if product.product_type not in (ProductType.FRAGILE, ProductType.LIQUID):
            max_layer = 0
            for item in updated_items:
                max_layer = max(max_layer, item.position_layer)
            layer = max_layer + 1 if max_layer > 0 else 1

        new_item = PackageItem(
            product=product,
            quantity=quantity,
            position_layer=layer,
        )
        updated_items.append(new_item)

    return updated_items


def mark_package_shipped(package_id: UUID, tracking_number: str) -> Package:
    package = get_package(package_id)

    if package.status == PackageStatus.SHIPPED:
        raise PackageAlreadyShippedException(str(package_id))

    package.status = PackageStatus.SHIPPED
    package.tracking_number = tracking_number
    package.shipped_at = datetime.now()

    memory_store.save_package(package)
    return package


def get_packing_list(order_id: str) -> Optional[PackingList]:
    return memory_store.get_packing_list(order_id)


def generate_packing_list_text(order_id: str) -> str:
    packing_list = get_packing_list(order_id)
    if packing_list is None:
        return f"订单 {order_id} 暂无装箱记录"

    order = order_service.get_order(order_id)

    lines: list[str] = []
    lines.append("=" * 60)
    lines.append(f"装箱单号: {packing_list.packing_id}")
    lines.append(f"订单号: {order_id}")
    lines.append(f"装箱时间: {packing_list.created_at.strftime('%Y-%m-%d %H:%M:%S')}")
    lines.append(f"总箱数: {packing_list.total_boxes}")
    lines.append(f"总重量: {packing_list.total_weight_kg:.2f} kg")
    lines.append(f"总体积: {packing_list.total_volume_m3:.4f} m³")
    lines.append("-" * 60)

    for i, package in enumerate(packing_list.packages, 1):
        lines.append(f"\n【箱 {i} - {package.box_number}】")
        lines.append(f"  状态: {package.status.value}")
        lines.append(
            f"  尺寸: {package.box_type.length_cm}x{package.box_type.width_cm}x"
            f"{package.box_type.height_cm} cm"
        )
        lines.append(f"  重量: {package.total_weight_kg:.2f} kg")
        lines.append(f"  体积: {package.total_volume_m3:.4f} m³")
        lines.append(f"  填充物体积: {package.filler_volume_m3:.4f} m³")

        if package.has_fragile_label:
            lines.append("  【标签】易碎品")

        if package.tracking_number:
            lines.append(f"  物流单号: {package.tracking_number}")

        lines.append("  商品清单:")
        sorted_items = sorted(package.items, key=lambda x: x.position_layer)
        for item in sorted_items:
            frag_tag = ""
            if item.product.product_type == ProductType.FRAGILE:
                frag_tag = " [易碎]"
            elif item.product.product_type == ProductType.LIQUID:
                frag_tag = " [液体]"
            lines.append(
                f"    - 第{item.position_layer}层: {item.product.name} "
                f"x{item.quantity} ({item.product.weight_kg}kg/件){frag_tag}"
            )

    lines.append("\n" + "=" * 60)
    return "\n".join(lines)
