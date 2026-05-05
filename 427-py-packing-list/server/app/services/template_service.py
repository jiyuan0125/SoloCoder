from typing import List
from uuid import UUID, uuid4

from shared.models import PackingList, Package
from server.app.exceptions import (
    TemplateNotFoundException,
    InvalidOperationException,
)
from server.app.repositories.memory_store import memory_store


def save_template(order_id: str, template_name: str) -> PackingList:
    packing_list = memory_store.get_packing_list(order_id)
    if packing_list is None:
        raise InvalidOperationException(f"订单 {order_id} 暂无装箱记录，无法保存模板")

    if len(packing_list.packages) == 0:
        raise InvalidOperationException(f"订单 {order_id} 暂无包裹，无法保存模板")

    template = PackingList(
        packing_id=uuid4(),
        order_id="TEMPLATE",
        packages=[
            Package(
                box_number=pkg.box_number,
                box_type=pkg.box_type,
                items=[
                    item.model_copy()
                    for item in pkg.items
                ],
                status=pkg.status,
            )
            for pkg in packing_list.packages
        ],
        is_template=True,
        template_name=template_name,
    )

    memory_store.save_packing_template(template)
    return template


def get_template(template_id: UUID) -> PackingList:
    template = memory_store.get_packing_template(template_id)
    if template is None:
        raise TemplateNotFoundException(str(template_id))
    return template


def get_all_templates() -> List[PackingList]:
    return memory_store.get_all_templates()


def delete_template(template_id: UUID) -> None:
    template = get_template(template_id)
    memory_store.packing_templates.pop(template.packing_id, None)


def apply_template_to_order(order_id: str, template_id: UUID) -> PackingList:
    from server.app.services import order_service
    from server.app.services.packing_service import validate_package

    order = order_service.get_order(order_id)
    template = get_template(template_id)

    new_packages: list[Package] = []
    for i, template_pkg in enumerate(template.packages):
        box_number = f"BOX-{order_id}-{i + 1:03d}"
        new_pkg = Package(
            box_number=box_number,
            box_type=template_pkg.box_type,
            items=[
                item.model_copy()
                for item in template_pkg.items
            ],
            status=template_pkg.status,
        )
        validate_package(new_pkg)
        new_packages.append(new_pkg)

        memory_store.save_package(new_pkg)

    packing_list = PackingList(
        order_id=order_id,
        packages=new_packages,
    )

    memory_store.save_packing_list(packing_list)
    return packing_list
