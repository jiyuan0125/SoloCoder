package com.training.common.enums;

public enum ErrorCode {
    SUCCESS(0, "操作成功"),
    INVALID_PARAM(1001, "参数无效"),
    COURSE_NOT_FOUND(1002, "课程不存在"),
    COURSE_ALREADY_STARTED(1003, "课程已开始"),
    REGISTRATION_FULL(1004, "报名名额已满"),
    ALREADY_REGISTERED(1005, "已报名该课程"),
    REGISTRATION_NOT_FOUND(1006, "报名记录不存在"),
    INVALID_SCORE(1007, "分数无效"),
    DEADLINE_PASSED(1008, "报名截止时间已过"),
    CANCELLATION_NOT_ALLOWED(1009, "取消报名不允许"),
    INSTRUCTOR_NOT_FOUND(1010, "讲师不存在"),
    EMPLOYEE_NOT_FOUND(1011, "员工不存在"),
    DEPARTMENT_NOT_FOUND(1012, "部门不存在"),
    EVALUATION_NOT_FOUND(1013, "评价不存在"),
    CANCEL_PENDING_APPROVAL(1014, "取消申请待审批"),
    INSUFFICIENT_PERMISSION(1015, "权限不足"),
    SYSTEM_ERROR(9999, "系统错误");

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
