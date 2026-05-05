from enum import Enum


class ErrorCode(str, Enum):
    PRODUCT_NOT_FOUND = "product_not_found"
    PRODUCT_ALREADY_EXISTS = "product_already_exists"
    INSUFFICIENT_STOCK = "insufficient_stock"
    ALERT_NOT_FOUND = "alert_not_found"
    REPLENISHMENT_NOT_FOUND = "replenishment_not_found"
    REPORT_NOT_FOUND = "report_not_found"
    INVALID_SAFETY_STOCK = "invalid_safety_stock"
    INVALID_CATEGORY = "invalid_category"
    INTERNAL_ERROR = "internal_error"
