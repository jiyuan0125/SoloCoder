from typing import Optional

from shared import constants


class PackingSystemException(Exception):
    error_code: str
    message: str
    status_code: int

    def __init__(
        self, error_code: str, message: str, status_code: int = 400
    ) -> None:
        self.error_code = error_code
        self.message = message
        self.status_code = status_code
        super().__init__(message)


class OrderNotFoundException(PackingSystemException):
    def __init__(self, order_id: str) -> None:
        super().__init__(
            error_code=constants.ORDER_NOT_FOUND,
            message=f"订单 {order_id} 不存在",
            status_code=404,
        )


class PackageNotFoundException(PackingSystemException):
    def __init__(self, package_id: str) -> None:
        super().__init__(
            error_code=constants.PACKAGE_NOT_FOUND,
            message=f"包裹 {package_id} 不存在",
            status_code=404,
        )


class ProductNotFoundException(PackingSystemException):
    def __init__(self, product_id: str) -> None:
        super().__init__(
            error_code=constants.PRODUCT_NOT_FOUND,
            message=f"商品 {product_id} 不存在",
            status_code=404,
        )


class BoxTypeNotFoundException(PackingSystemException):
    def __init__(self, box_type_id: str) -> None:
        super().__init__(
            error_code=constants.BOX_TYPE_NOT_FOUND,
            message=f"箱型 {box_type_id} 不存在",
            status_code=404,
        )


class TemplateNotFoundException(PackingSystemException):
    def __init__(self, template_id: str) -> None:
        super().__init__(
            error_code=constants.TEMPLATE_NOT_FOUND,
            message=f"装箱模板 {template_id} 不存在",
            status_code=404,
        )


class WeightExceedsLimitException(PackingSystemException):
    def __init__(self, weight: float, limit: float) -> None:
        super().__init__(
            error_code=constants.WEIGHT_EXCEEDS_LIMIT,
            message=f"重量 {weight:.2f}kg 超过限制 {limit}kg",
            status_code=400,
        )


class VolumeExceedsLimitException(PackingSystemException):
    def __init__(self, volume: float, limit: float) -> None:
        super().__init__(
            error_code=constants.VOLUME_EXCEEDS_LIMIT,
            message=f"体积 {volume:.6f}m³ 超过限制 {limit}m³",
            status_code=400,
        )


class HeavyItemMustBeAloneException(PackingSystemException):
    def __init__(self, product_name: str, weight: float) -> None:
        super().__init__(
            error_code=constants.HEAVY_ITEM_MUST_BE_ALONE,
            message=f"商品 {product_name} 重量 {weight:.2f}kg 超过30kg，必须单独装箱",
            status_code=400,
        )


class FragileMustBeOnBottomException(PackingSystemException):
    def __init__(self, product_name: str) -> None:
        super().__init__(
            error_code=constants.FRAGILE_MUST_BE_ON_BOTTOM,
            message=f"易碎品/液体 {product_name} 必须放在包裹最底层",
            status_code=400,
        )


class VolumeUtilizationExceededException(PackingSystemException):
    def __init__(self, utilization: float, limit: float) -> None:
        super().__init__(
            error_code=constants.VOLUME_UTILIZATION_EXCEEDED,
            message=f"容积利用率 {utilization:.1%} 超过限制 {limit:.0%}",
            status_code=400,
        )


class PackageAlreadyShippedException(PackingSystemException):
    def __init__(self, package_id: str) -> None:
        super().__init__(
            error_code=constants.PACKAGE_ALREADY_SHIPPED,
            message=f"包裹 {package_id} 已发出，不可修改或删除",
            status_code=400,
        )


class OrderAlreadyShippedException(PackingSystemException):
    def __init__(self, order_id: str) -> None:
        super().__init__(
            error_code=constants.ORDER_ALREADY_SHIPPED,
            message=f"订单 {order_id} 已发出，不可修改",
            status_code=400,
        )


class InsufficientStockException(PackingSystemException):
    def __init__(self, box_type_name: str) -> None:
        super().__init__(
            error_code=constants.INSUFFICIENT_STOCK,
            message=f"箱型 {box_type_name} 库存不足",
            status_code=400,
        )


class DuplicateOrderIdException(PackingSystemException):
    def __init__(self, order_id: str) -> None:
        super().__init__(
            error_code=constants.DUPLICATE_ORDER_ID,
            message=f"订单ID {order_id} 已存在",
            status_code=400,
        )


class DuplicateBoxTypeException(PackingSystemException):
    def __init__(self, box_type_id: str) -> None:
        super().__init__(
            error_code=constants.DUPLICATE_BOX_TYPE,
            message=f"箱型ID {box_type_id} 已存在",
            status_code=400,
        )


class InvalidOperationException(PackingSystemException):
    def __init__(self, message: str) -> None:
        super().__init__(
            error_code=constants.INVALID_OPERATION,
            message=message,
            status_code=400,
        )


class ValidationException(PackingSystemException):
    def __init__(self, message: str) -> None:
        super().__init__(
            error_code=constants.VALIDATION_ERROR,
            message=message,
            status_code=422,
        )


class InternalException(PackingSystemException):
    def __init__(self, message: str) -> None:
        super().__init__(
            error_code=constants.INTERNAL_ERROR,
            message=message,
            status_code=500,
        )
