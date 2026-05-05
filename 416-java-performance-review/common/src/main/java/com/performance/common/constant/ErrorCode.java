package com.performance.common.constant;

public enum ErrorCode {
    SUCCESS(0, "成功"),
    PARAM_ERROR(400, "参数错误"),
    NOT_FOUND(404, "资源不存在"),
    FORBIDDEN(403, "无权限"),
    SERVER_ERROR(500, "服务器错误"),
    
    CYCLE_ALREADY_EXISTS(1001, "该季度评估周期已存在"),
    CYCLE_NOT_FOUND(1002, "评估周期不存在"),
    CYCLE_NOT_STARTED(1003, "评估周期未开始"),
    CYCLE_NOT_IN_SELF_REVIEW(1004, "不在自评阶段"),
    CYCLE_NOT_IN_MANAGER_REVIEW(1005, "不在上级评分阶段"),
    CYCLE_NOT_IN_HR_CONFIRM(1006, "不在HR确认阶段"),
    CYCLE_ALREADY_COMPLETED(1007, "评估周期已完成"),
    
    EMPLOYEE_NOT_FOUND(2001, "员工不存在"),
    DEPARTMENT_NOT_FOUND(2002, "部门不存在"),
    MANAGER_NOT_FOUND(2003, "上级不存在"),
    
    SCORE_INVALID(3001, "分数必须在1到5之间"),
    SCORE_DIFFERENCE_ZERO(3002, "上级评分与自评分差异必须至少为0.5分，不能恰好相等"),
    REVIEW_ALREADY_SUBMITTED(3003, "评估已提交，不可重复提交"),
    REVIEW_NOT_SUBMITTED(3004, "评估未提交");

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