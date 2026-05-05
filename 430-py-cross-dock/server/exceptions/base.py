from typing import Optional, Any

from shared.models import ValidationResult


class CrossDockException(Exception):
    def __init__(
        self,
        message: str,
        details: Optional[dict[str, Any]] = None,
    ) -> None:
        super().__init__(message)
        self.message = message
        self.details = details or {}


class OrderNotFoundException(CrossDockException):
    def __init__(self, order_id: str) -> None:
        super().__init__(
            message=f"越库单 {order_id} 不存在",
            details={"order_id": order_id},
        )


class InvalidStatusTransitionException(CrossDockException):
    def __init__(
        self,
        current_status: str,
        target_status: str,
        order_id: str,
    ) -> None:
        super().__init__(
            message=f"越库单 {order_id} 无法从 {current_status} 转换到 {target_status}",
            details={
                "order_id": order_id,
                "current_status": current_status,
                "target_status": target_status,
            },
        )


class SameOperatorException(CrossDockException):
    def __init__(
        self,
        operator_id: str,
        order_id: str,
    ) -> None:
        super().__init__(
            message=f"操作员 {operator_id} 不能同时处理同一越库单的入库和出库",
            details={
                "order_id": order_id,
                "operator_id": operator_id,
            },
        )


class ItemMismatchException(CrossDockException):
    def __init__(
        self,
        order_id: str,
        validation_result: Optional[ValidationResult] = None,
    ) -> None:
        super().__init__(
            message=f"越库单 {order_id} 商品不一致",
            details={
                "order_id": order_id,
            },
        )
        self.validation_result = validation_result


class TimeoutException(CrossDockException):
    def __init__(
        self,
        order_id: str,
        timeout_minutes: int,
    ) -> None:
        super().__init__(
            message=f"越库单 {order_id} 已超时（{timeout_minutes} 分钟）",
            details={
                "order_id": order_id,
                "timeout_minutes": timeout_minutes,
            },
        )


class CrossDayException(CrossDockException):
    def __init__(
        self,
        order_id: str,
        inbound_date: str,
        outbound_date: str,
    ) -> None:
        super().__init__(
            message=f"越库单 {order_id} 入库和出库跨日（{inbound_date} -> {outbound_date}）",
            details={
                "order_id": order_id,
                "inbound_date": inbound_date,
                "outbound_date": outbound_date,
            },
        )


class BatchNotReadyException(CrossDockException):
    def __init__(
        self,
        outbound_order_number: str,
        completed_count: int,
        total_count: int,
    ) -> None:
        super().__init__(
            message=f"批量越库单 {outbound_order_number} 未全部完成入库（{completed_count}/{total_count}）",
            details={
                "outbound_order_number": outbound_order_number,
                "completed_count": completed_count,
                "total_count": total_count,
            },
        )


class DuplicateOrderException(CrossDockException):
    def __init__(
        self,
        order_number: str,
    ) -> None:
        super().__init__(
            message=f"越库单号 {order_number} 已存在",
            details={"order_number": order_number},
        )


class OutboundOrderNotFoundException(CrossDockException):
    def __init__(self, outbound_order_number: str) -> None:
        super().__init__(
            message=f"出库单号 {outbound_order_number} 不存在",
            details={"outbound_order_number": outbound_order_number},
        )


class ExceptionPlanNotFoundException(CrossDockException):
    def __init__(self, exception_type: str) -> None:
        super().__init__(
            message=f"异常类型 {exception_type} 的处理预案不存在",
            details={"exception_type": exception_type},
        )
