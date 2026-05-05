package com.example.common.enums;

public enum ErrorCode {
    SUCCESS(0, "操作成功"),
    SYSTEM_ERROR(1000, "系统错误"),
    PARAM_ERROR(1001, "参数错误"),
    EMPLOYEE_NOT_FOUND(2001, "员工不存在"),
    TEMPLATE_NOT_FOUND(2002, "模板不存在"),
    CHECKLIST_ITEM_NOT_FOUND(2003, "清单事项不存在"),
    ALERT_NOT_FOUND(2004, "告警记录不存在"),
    DUPLICATE_TEMPLATE_NAME(2005, "模板名称已存在"),
    INVALID_POSITION_TYPE(2006, "无效的岗位类型");

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
