from enum import Enum, unique


@unique
class TransferStatus(str, Enum):
    PENDING_CONFIRM = "pending_confirm"
    CONFIRMED = "confirmed"
    IN_TRANSIT = "in_transit"
    ARRIVED = "arrived"
    STOCKED = "stocked"
    CANCELLED = "cancelled"


@unique
class ApprovalLevel(str, Enum):
    NONE = "none"
    WAREHOUSE_MANAGER = "warehouse_manager"
    LOGISTICS_DIRECTOR = "logistics_director"


@unique
class ApprovalStatus(str, Enum):
    PENDING = "pending"
    APPROVED = "approved"
    REJECTED = "rejected"


@unique
class TransferType(str, Enum):
    INTRA_COMPANY = "intra_company"
    INTER_COMPANY = "inter_company"


@unique
class TransactionType(str, Enum):
    TRANSFER_OUT = "transfer_out"
    TRANSFER_IN = "transfer_in"
    FREEZE = "freeze"
    UNFREEZE = "unfreeze"
    LOSS = "loss"


@unique
class ErrorCode(int, Enum):
    SUCCESS = 0
    INVALID_PARAM = 10001
    WAREHOUSE_NOT_FOUND = 20001
    PRODUCT_NOT_FOUND = 20002
    INSUFFICIENT_INVENTORY = 20003
    TRANSFER_NOT_FOUND = 30001
    INVALID_STATUS_TRANSITION = 30002
    TRANSFER_ALREADY_APPROVED = 30003
    TRANSFER_REJECTED = 30004
    APPROVAL_PERMISSION_DENIED = 40001
    APPROVAL_NOT_REQUIRED = 40002
    INTER_COMPANY_APPROVAL_PENDING = 50001
    INTERNAL_ERROR = 99999
