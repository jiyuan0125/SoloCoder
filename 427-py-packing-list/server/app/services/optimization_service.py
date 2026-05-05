from typing import List, Tuple

from shared import constants
from shared.models import (
    Order,
    OrderItem,
    BoxType,
    PackingSuggestion,
    ProductType,
)
from server.app.repositories.memory_store import memory_store


def suggest_packing(order_id: str) -> List[PackingSuggestion]:
    from server.app.services import order_service

    order = order_service.get_order(order_id)
    box_types = memory_store.get_all_box_types()

    if not box_types:
        return []

    suggestions: list[PackingSuggestion] = []

    heavy_items: list[Tuple[OrderItem, int]] = []
    normal_items: list[Tuple[OrderItem, int]] = []

    for item in order.items:
        if item.product.weight_kg > constants.MAX_PACKAGE_WEIGHT_KG:
            heavy_items.append((item, item.quantity))
        else:
            normal_items.append((item, item.quantity))

    for item, qty in heavy_items:
        for _ in range(qty):
            suitable_box = _find_suitable_box(
                item.product.volume_m3,
                item.product.weight_kg,
                box_types,
            )
            if suitable_box:
                suggestion = PackingSuggestion(
                    suggested_box_type=suitable_box,
                    items=[OrderItem(product=item.product, quantity=1)],
                    estimated_weight_kg=item.product.weight_kg,
                    estimated_volume_m3=item.product.volume_m3,
                    confidence_score=_calculate_confidence(
                        item.product.volume_m3,
                        item.product.weight_kg,
                        suitable_box,
                    ),
                )
                suggestions.append(suggestion)

    if normal_items:
        remaining: list[Tuple[OrderItem, int]] = list(normal_items)

        while remaining:
            selected_items: list[OrderItem] = []
            total_volume = 0.0
            total_weight = 0.0

            sorted_remaining = sorted(
                remaining,
                key=lambda x: (
                    x[0].product.product_type
                    not in (ProductType.FRAGILE, ProductType.LIQUID),
                    -x[0].product.volume_m3,
                ),
            )

            for box_type in sorted(box_types, key=lambda b: b.volume_m3):
                effective_volume = box_type.effective_volume_m3
                max_weight = min(box_type.max_weight_kg, constants.MAX_PACKAGE_WEIGHT_KG)
                max_volume = min(effective_volume, constants.MAX_PACKAGE_VOLUME_M3)

                temp_volume = 0.0
                temp_weight = 0.0
                temp_items: list[OrderItem] = []
                used_indices: set[int] = set()

                for idx, (item, qty) in enumerate(sorted_remaining):
                    for i in range(qty):
                        item_vol = item.product.volume_m3
                        item_wt = item.product.weight_kg

                        if temp_volume + item_vol <= max_volume and temp_weight + item_wt <= max_weight:
                            temp_volume += item_vol
                            temp_weight += item_wt

                            existing = next(
                                (ti for ti in temp_items if ti.product.product_id == item.product.product_id),
                                None,
                            )
                            if existing:
                                existing.quantity += 1
                            else:
                                temp_items.append(OrderItem(product=item.product, quantity=1))

                            used_indices.add(idx)
                        else:
                            break

                if temp_items:
                    selected_items = temp_items
                    total_volume = temp_volume
                    total_weight = temp_weight

                    new_remaining: list[Tuple[OrderItem, int]] = []
                    for idx, (item, qty) in enumerate(sorted_remaining):
                        if idx in used_indices:
                            used_qty = 0
                            for ti in temp_items:
                                if ti.product.product_id == item.product.product_id:
                                    used_qty = ti.quantity
                                    break
                            if qty > used_qty:
                                new_remaining.append((item, qty - used_qty))
                        else:
                            new_remaining.append((item, qty))

                    remaining = new_remaining
                    break
                else:
                    if sorted_remaining:
                        item, qty = sorted_remaining[0]
                        selected_items = [OrderItem(product=item.product, quantity=1)]
                        total_volume = item.product.volume_m3
                        total_weight = item.product.weight_kg

                        if qty > 1:
                            remaining = [(item, qty - 1)] + sorted_remaining[1:]
                        else:
                            remaining = sorted_remaining[1:]
                        break

            if selected_items:
                suitable_box = _find_suitable_box(total_volume, total_weight, box_types)
                if suitable_box:
                    suggestion = PackingSuggestion(
                        suggested_box_type=suitable_box,
                        items=selected_items,
                        estimated_weight_kg=total_weight,
                        estimated_volume_m3=total_volume,
                        confidence_score=_calculate_confidence(
                            total_volume, total_weight, suitable_box
                        ),
                    )
                    suggestions.append(suggestion)
            else:
                break

    return suggestions


def _find_suitable_box(
    volume: float, weight: float, box_types: List[BoxType]
) -> BoxType:
    sorted_boxes = sorted(box_types, key=lambda b: b.volume_m3)

    for box in sorted_boxes:
        effective_volume = box.effective_volume_m3
        max_weight = min(box.max_weight_kg, constants.MAX_PACKAGE_WEIGHT_KG)
        max_volume = min(effective_volume, constants.MAX_PACKAGE_VOLUME_M3)

        if volume <= max_volume and weight <= max_weight:
            return box

    return sorted_boxes[-1] if sorted_boxes else sorted_boxes[0]


def _calculate_confidence(volume: float, weight: float, box_type: BoxType) -> float:
    volume_ratio = volume / box_type.effective_volume_m3
    weight_ratio = weight / min(box_type.max_weight_kg, constants.MAX_PACKAGE_WEIGHT_KG)

    return (volume_ratio * 0.6 + weight_ratio * 0.4)
