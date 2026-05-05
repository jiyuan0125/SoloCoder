from typing import List

from shared.models import BoxType, PackagingMaterialInventory
from server.app.exceptions import (
    DuplicateBoxTypeException,
    BoxTypeNotFoundException,
)
from server.app.repositories.memory_store import memory_store


def create_box_type(
    box_type_id: str,
    name: str,
    length_cm: float,
    width_cm: float,
    height_cm: float,
    max_weight_kg: float,
    initial_stock: int,
) -> BoxType:
    existing = memory_store.get_box_type(box_type_id)
    if existing is not None:
        raise DuplicateBoxTypeException(box_type_id)

    box_type = BoxType(
        box_type_id=box_type_id,
        name=name,
        length_cm=length_cm,
        width_cm=width_cm,
        height_cm=height_cm,
        max_weight_kg=max_weight_kg,
    )

    inventory = PackagingMaterialInventory(
        box_type=box_type,
        stock_quantity=initial_stock,
        used_quantity=0,
    )

    memory_store.save_box_type(box_type)
    memory_store.save_inventory(inventory)

    return box_type


def get_box_type(box_type_id: str) -> BoxType:
    box_type = memory_store.get_box_type(box_type_id)
    if box_type is None:
        raise BoxTypeNotFoundException(box_type_id)
    return box_type


def get_all_box_types() -> List[BoxType]:
    return memory_store.get_all_box_types()


def get_all_inventories() -> List[PackagingMaterialInventory]:
    return memory_store.get_all_inventories()


def add_stock(box_type_id: str, quantity: int) -> PackagingMaterialInventory:
    box_type = get_box_type(box_type_id)
    inventory = memory_store.get_inventory(box_type_id)
    if inventory is None:
        inventory = PackagingMaterialInventory(
            box_type=box_type,
            stock_quantity=quantity,
            used_quantity=0,
        )
    else:
        inventory.stock_quantity += quantity

    memory_store.save_inventory(inventory)
    return inventory


def initialize_default_box_types() -> None:
    default_boxes: list[dict[str, str | float | int]] = [
        {"box_type_id": "BOX-S", "name": "小号箱", "length": 20.0, "width": 15.0, "height": 10.0},
        {"box_type_id": "BOX-M", "name": "中号箱", "length": 40.0, "width": 30.0, "height": 20.0},
        {"box_type_id": "BOX-L", "name": "大号箱", "length": 60.0, "width": 40.0, "height": 30.0},
        {"box_type_id": "BOX-XL", "name": "特大号箱", "length": 80.0, "width": 60.0, "height": 40.0},
    ]

    for box in default_boxes:
        box_type_id = str(box["box_type_id"])
        existing = memory_store.get_box_type(box_type_id)
        if existing is None:
            box_type = BoxType(
                box_type_id=box_type_id,
                name=str(box["name"]),
                length_cm=float(box["length"]),
                width_cm=float(box["width"]),
                height_cm=float(box["height"]),
                max_weight_kg=30.0,
            )
            inventory = PackagingMaterialInventory(
                box_type=box_type,
                stock_quantity=100,
                used_quantity=0,
            )
            memory_store.save_box_type(box_type)
            memory_store.save_inventory(inventory)
