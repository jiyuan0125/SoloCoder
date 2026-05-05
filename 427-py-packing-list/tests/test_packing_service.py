from uuid import uuid4

import pytest

from shared import constants
from shared.models import (
    Product,
    ProductType,
    Package,
    BoxType,
    PackageStatus,
    Order,
    OrderItem,
    OrderStatus,
    PackagingMaterialInventory,
)
from server.app.repositories.memory_store import MemoryStore
from server.app.services.packing_service import (
    add_item_to_package,
    create_package,
    validate_package,
    _create_updated_items_list,
)
from server.app.exceptions import (
    WeightExceedsLimitException,
    HeavyItemMustBeAloneException,
)


@pytest.fixture
def test_store() -> MemoryStore:
    store = MemoryStore()
    store._initialized = False
    store.__init__()
    return store


@pytest.fixture
def heavy_product() -> Product:
    return Product(
        product_id="HEAVY001",
        name="超重商品",
        weight_kg=35.0,
        length_cm=100,
        width_cm=100,
        height_cm=100,
        product_type=ProductType.NORMAL,
    )


@pytest.fixture
def normal_product() -> Product:
    return Product(
        product_id="NORM001",
        name="普通商品",
        weight_kg=5.0,
        length_cm=20,
        width_cm=15,
        height_cm=10,
        product_type=ProductType.NORMAL,
    )


@pytest.fixture
def box_type_m() -> BoxType:
    return BoxType(
        box_type_id="BOX-M",
        name="中号箱",
        length_cm=40,
        width_cm=30,
        height_cm=20,
        max_weight_kg=30.0,
    )


@pytest.fixture
def test_order(normal_product: Product) -> Order:
    return Order(
        order_id="TEST-ORDER-001",
        items=[OrderItem(product=normal_product, quantity=1)],
        status=OrderStatus.CREATED,
    )


class TestDataPollutionBug:
    """
    测试修复的数据污染问题：
    往空包裹添加两件35kg商品 -> 返回WEIGHT_EXCEEDS_LIMIT错误
    但商品残留在包裹items中
    再次添加同一件商品 -> 错误报HEAVY_ITEM_MUST_BE_ALONE
    """

    def test_failed_add_should_not_pollute_package_items(
        self,
        test_store: MemoryStore,
        heavy_product: Product,
        normal_product: Product,
        box_type_m: BoxType,
    ) -> None:
        test_store.save_box_type(box_type_m)
        inventory = PackagingMaterialInventory(
            box_type=box_type_m,
            stock_quantity=100,
            used_quantity=0,
        )
        test_store.save_inventory(inventory)
        test_store.save_product(heavy_product)

        package = Package(
            package_id=uuid4(),
            box_number="BOX-TEST-001",
            box_type=box_type_m,
            items=[],
            status=PackageStatus.PENDING,
        )
        test_store.save_package(package)

        assert len(package.items) == 0, "初始包裹应该为空"

        with pytest.raises(WeightExceedsLimitException):
            add_item_to_package(
                package_id=package.package_id,
                product_id=heavy_product.product_id,
                quantity=2,
            )

        assert len(package.items) == 0, (
            "添加失败后包裹应该仍然为空，"
            f"但实际有 {len(package.items)} 件商品"
        )

        with pytest.raises(WeightExceedsLimitException):
            add_item_to_package(
                package_id=package.package_id,
                product_id=heavy_product.product_id,
                quantity=1,
            )

        assert len(package.items) == 0, (
            "第二次添加失败后包裹仍然应该为空"
        )

    def test_heavy_item_must_be_alone_logic(
        self,
        test_store: MemoryStore,
        heavy_product: Product,
        normal_product: Product,
        box_type_m: BoxType,
    ) -> None:
        test_store.save_box_type(box_type_m)
        inventory = PackagingMaterialInventory(
            box_type=box_type_m,
            stock_quantity=100,
            used_quantity=0,
        )
        test_store.save_inventory(inventory)
        test_store.save_product(heavy_product)
        test_store.save_product(normal_product)

        package = Package(
            package_id=uuid4(),
            box_number="BOX-TEST-002",
            box_type=box_type_m,
            items=[],
            status=PackageStatus.PENDING,
        )
        test_store.save_package(package)

        updated_package = add_item_to_package(
            package_id=package.package_id,
            product_id=normal_product.product_id,
            quantity=1,
        )
        assert len(updated_package.items) == 1

        with pytest.raises(HeavyItemMustBeAloneException):
            add_item_to_package(
                package_id=package.package_id,
                product_id=heavy_product.product_id,
                quantity=1,
            )

        assert len(updated_package.items) == 1, (
            "添加超重商品失败后，普通商品应该保留"
        )


class TestCreateUpdatedItemsList:
    """测试 _create_updated_items_list 辅助函数"""

    def test_should_not_modify_original_items(
        self,
        normal_product: Product,
        box_type_m: BoxType,
    ) -> None:
        package = Package(
            package_id=uuid4(),
            box_number="BOX-TEST",
            box_type=box_type_m,
            items=[],
            status=PackageStatus.PENDING,
        )

        original_items = package.items.copy()

        temp_items = _create_updated_items_list(package, normal_product, 2)

        assert len(original_items) == 0, "原始 items 不应该被修改"
        assert len(temp_items) == 1, "临时 items 应该包含新商品"


class TestValidatePackage:
    """测试 validate_package 函数"""

    def test_validate_package_weight_limit(
        self,
        heavy_product: Product,
        box_type_m: BoxType,
    ) -> None:
        package = Package(
            package_id=uuid4(),
            box_number="BOX-TEST",
            box_type=box_type_m,
            items=[],
            status=PackageStatus.PENDING,
        )

        temp_items = _create_updated_items_list(package, heavy_product, 2)

        temp_package = Package(
            package_id=package.package_id,
            box_number=package.box_number,
            box_type=package.box_type,
            items=temp_items,
            status=package.status,
        )

        with pytest.raises(WeightExceedsLimitException) as exc_info:
            validate_package(temp_package)

        assert "30" in str(exc_info.value)
