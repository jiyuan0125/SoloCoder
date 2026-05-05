from typing import Any

from shared.enums import ErrorCode


class BusinessException(Exception):
    def __init__(
        self,
        error_code: ErrorCode,
        message: str,
        details: dict[str, Any] | None = None,
    ) -> None:
        self.error_code = error_code
        self.message = message
        self.details = details or {}
        super().__init__(message)


class WarehouseNotFoundException(BusinessException):
    def __init__(self, warehouse_id: str) -> None:
        super().__init__(
            error_code=ErrorCode.WAREHOUSE_NOT_FOUND,
            message=f"Warehouse not found: {warehouse_id}",
            details={"warehouse_id": warehouse_id},
        )


class ProductNotFoundException(BusinessException):
    def __init__(self, product_id: str) -> None:
        super().__init__(
            error_code=ErrorCode.PRODUCT_NOT_FOUND,
            message=f"Product not found: {product_id}",
            details={"product_id": product_id},
        )


class InsufficientInventoryException(BusinessException):
    def __init__(
        self,
        warehouse_id: str,
        product_id: str,
        available: int,
        required: int,
    ) -> None:
        super().__init__(
            error_code=ErrorCode.INSUFFICIENT_INVENTORY,
            message=(
                f"Insufficient inventory. Warehouse: {warehouse_id}, "
                f"Product: {product_id}, Available: {available}, Required: {required}"
            ),
            details={
                "warehouse_id": warehouse_id,
                "product_id": product_id,
                "available_quantity": available,
                "required_quantity": required,
            },
        )


class TransferNotFoundException(BusinessException):
    def __init__(self, transfer_id: str) -> None:
        super().__init__(
            error_code=ErrorCode.TRANSFER_NOT_FOUND,
            message=f"Transfer order not found: {transfer_id}",
            details={"transfer_id": transfer_id},
        )


class InvalidStatusTransitionException(BusinessException):
    def __init__(
        self,
        current_status: str,
        target_status: str,
    ) -> None:
        super().__init__(
            error_code=ErrorCode.INVALID_STATUS_TRANSITION,
            message=f"Invalid status transition: {current_status} -> {target_status}",
            details={"current_status": current_status, "target_status": target_status},
        )


class TransferAlreadyApprovedException(BusinessException):
    def __init__(self, transfer_id: str) -> None:
        super().__init__(
            error_code=ErrorCode.TRANSFER_ALREADY_APPROVED,
            message=f"Transfer order already approved: {transfer_id}",
            details={"transfer_id": transfer_id},
        )


class TransferRejectedException(BusinessException):
    def __init__(self, transfer_id: str) -> None:
        super().__init__(
            error_code=ErrorCode.TRANSFER_REJECTED,
            message=f"Transfer order has been rejected: {transfer_id}",
            details={"transfer_id": transfer_id},
        )


class ApprovalPermissionDeniedException(BusinessException):
    def __init__(self, transfer_id: str, required_level: str) -> None:
        super().__init__(
            error_code=ErrorCode.APPROVAL_PERMISSION_DENIED,
            message=(
                f"Insufficient approval permission for transfer: {transfer_id}. "
                f"Required level: {required_level}"
            ),
            details={"transfer_id": transfer_id, "required_level": required_level},
        )


class ApprovalNotRequiredException(BusinessException):
    def __init__(self, transfer_id: str) -> None:
        super().__init__(
            error_code=ErrorCode.APPROVAL_NOT_REQUIRED,
            message=f"Approval not required for transfer: {transfer_id}",
            details={"transfer_id": transfer_id},
        )


class InterCompanyApprovalPendingException(BusinessException):
    def __init__(self, transfer_id: str) -> None:
        super().__init__(
            error_code=ErrorCode.INTER_COMPANY_APPROVAL_PENDING,
            message=f"Inter-company approval pending for transfer: {transfer_id}",
            details={"transfer_id": transfer_id},
        )


class InvalidParamException(BusinessException):
    def __init__(self, message: str, param_name: str | None = None) -> None:
        details: dict[str, Any] = {}
        if param_name:
            details["param_name"] = param_name
        super().__init__(
            error_code=ErrorCode.INVALID_PARAM,
            message=message,
            details=details,
        )
