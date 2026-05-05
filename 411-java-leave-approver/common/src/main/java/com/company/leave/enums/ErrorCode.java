package com.company.leave.enums;

public enum ErrorCode {
    SUCCESS(0, "成功"),
    INVALID_REQUEST(400, "无效请求"),
    EMPLOYEE_NOT_FOUND(404, "员工不存在"),
    LEAVE_NOT_FOUND(404, "请假记录不存在"),
    INSUFFICIENT_LEAVE_DAYS(400, "假期天数不足"),
    OVERLAPPING_LEAVE(400, "同一时间段已存在请假记录"),
    MISSING_ATTACHMENT(400, "缺少必需的附件"),
    INVALID_DATE_RANGE(400, "日期范围无效"),
    ALREADY_PROCESSED(400, "该请假已被处理，无法再次操作"),
    INTERNAL_ERROR(500, "服务器内部错误");

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
