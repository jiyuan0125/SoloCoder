from enum import IntEnum, auto


class ErrorCode(IntEnum):
    SUCCESS = 0
    WAYBILL_NOT_FOUND = auto()
    WAYBILL_ALREADY_EXISTS = auto()
    WAYBILL_ALREADY_SIGNED = auto()
    INVALID_NODE_TIME = auto()
    LOCATION_REQUIRED = auto()
    SIGNED_BY_REQUIRED = auto()
    INVALID_WAYBILL_NUMBER = auto()
    EXCEED_BATCH_LIMIT = auto()
    PARENT_WAYBILL_NOT_FOUND = auto()
    PARENT_WAYBILL_ALREADY_SIGNED = auto()
    INTERNATIONAL_ONLY = auto()
    INVALID_OPERATOR = auto()


ERROR_MESSAGES: dict[ErrorCode, str] = {
    ErrorCode.SUCCESS: "成功",
    ErrorCode.WAYBILL_NOT_FOUND: "运单不存在",
    ErrorCode.WAYBILL_ALREADY_EXISTS: "运单已存在",
    ErrorCode.WAYBILL_ALREADY_SIGNED: "运单已签收，无法添加新节点",
    ErrorCode.INVALID_NODE_TIME: "节点时间不能早于上一个节点",
    ErrorCode.LOCATION_REQUIRED: "地点信息不能为空",
    ErrorCode.SIGNED_BY_REQUIRED: "签收时必须填写签收人",
    ErrorCode.INVALID_WAYBILL_NUMBER: "运单号格式无效",
    ErrorCode.EXCEED_BATCH_LIMIT: "批量查询运单号不能超过50个",
    ErrorCode.PARENT_WAYBILL_NOT_FOUND: "父运单不存在",
    ErrorCode.PARENT_WAYBILL_ALREADY_SIGNED: "父运单已签收，无法拆分",
    ErrorCode.INTERNATIONAL_ONLY: "仅国际运单可设置海关状态",
    ErrorCode.INVALID_OPERATOR: "操作人信息无效",
}


BATCH_QUERY_LIMIT: int = 50
ABNORMAL_HOURS_THRESHOLD: int = 72
WAYBILL_SEQUENCE_DIGITS: int = 6
