package com.payroll.common.enums;

public enum ErrorCode {

    SUCCESS(0, "操作成功"),
    INVALID_PARAM(1001, "参数错误"),
    ARCHIVE_NOT_FOUND(1002, "归档记录不存在"),
    ARCHIVE_ALREADY_EXISTS(1003, "归档记录已存在"),
    ARCHIVE_CANNOT_MODIFY(1004, "归档记录不可修改"),
    ARCHIVE_READ_ONLY(1005, "归档期间数据只读"),
    DATA_INTEGRITY_ERROR(1006, "数据完整性校验失败"),
    EMPLOYEE_NOT_FOUND(1007, "员工不存在"),
    INVALID_YEAR_MONTH(1008, "无效的年月格式"),
    INTERNAL_ERROR(9999, "系统内部错误");

    private final int code;
    private final String message;

    ErrorCode(int code, String message) {
        this.code = code;
        this.message = message;
    }

    public int getCode() {
        return code;
    }

    public String getMessage() {
        return message;
    }
}
