package com.employee.common.constant;

public enum ErrorCode {
    SUCCESS(0, "操作成功"),
    SYSTEM_ERROR(10001, "系统异常"),
    PARAM_INVALID(10002, "参数无效"),
    EMPLOYEE_NOT_FOUND(20001, "员工不存在"),
    EMPLOYEE_ID_EXISTS(20002, "员工工号已存在"),
    PHONE_ALREADY_USED(20003, "手机号已被其他在职员工使用"),
    EMAIL_INVALID_DOMAIN(20004, "邮箱必须是公司域名"),
    PERMISSION_DENIED(20005, "无权限执行此操作"),
    EMPLOYEE_ID_CANNOT_MODIFY(20006, "员工工号不可修改"),
    EMPLOYEE_ALREADY_RESIGNED(20007, "员工已离职");

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
