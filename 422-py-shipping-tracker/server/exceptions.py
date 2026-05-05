from shared.constants import ErrorCode, ERROR_MESSAGES


class ShippingServiceException(Exception):
    def __init__(self, error_code: ErrorCode, message: str | None = None) -> None:
        self.error_code = error_code
        self.message = message or ERROR_MESSAGES.get(error_code, "未知错误")
        super().__init__(self.message)


class WaybillNotFoundException(ShippingServiceException):
    def __init__(self, waybill_number: str) -> None:
        super().__init__(ErrorCode.WAYBILL_NOT_FOUND, f"运单 {waybill_number} 不存在")


class WaybillAlreadySignedException(ShippingServiceException):
    def __init__(self, waybill_number: str) -> None:
        super().__init__(
            ErrorCode.WAYBILL_ALREADY_SIGNED, f"运单 {waybill_number} 已签收，无法添加新节点"
        )


class InvalidNodeTimeException(ShippingServiceException):
    def __init__(self) -> None:
        super().__init__(ErrorCode.INVALID_NODE_TIME)


class LocationRequiredException(ShippingServiceException):
    def __init__(self) -> None:
        super().__init__(ErrorCode.LOCATION_REQUIRED)


class SignedByRequiredException(ShippingServiceException):
    def __init__(self) -> None:
        super().__init__(ErrorCode.SIGNED_BY_REQUIRED)


class ExceedBatchLimitException(ShippingServiceException):
    def __init__(self, count: int, limit: int) -> None:
        super().__init__(
            ErrorCode.EXCEED_BATCH_LIMIT, f"批量查询运单号不能超过{limit}个，当前{count}个"
        )


class ParentWaybillNotFoundException(ShippingServiceException):
    def __init__(self, waybill_number: str) -> None:
        super().__init__(
            ErrorCode.PARENT_WAYBILL_NOT_FOUND, f"父运单 {waybill_number} 不存在"
        )


class ParentWaybillAlreadySignedException(ShippingServiceException):
    def __init__(self, waybill_number: str) -> None:
        super().__init__(
            ErrorCode.PARENT_WAYBILL_ALREADY_SIGNED, f"父运单 {waybill_number} 已签收，无法拆分"
        )


class InternationalOnlyException(ShippingServiceException):
    def __init__(self) -> None:
        super().__init__(ErrorCode.INTERNATIONAL_ONLY)
