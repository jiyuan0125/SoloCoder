package com.recruitment.common.enums;

public enum ErrorCode {
    SUCCESS(0, "成功"),
    CANDIDATE_NOT_FOUND(1001, "候选人不存在"),
    APPLICATION_NOT_FOUND(1002, "申请记录不存在"),
    INVALID_STAGE_TRANSITION(1003, "无效的阶段跳转"),
    INVALID_SCORE(1004, "无效的评分，应为1-5分"),
    OFFER_ALREADY_ACCEPTED(1005, "该候选人已接受其他岗位offer"),
    OFFER_EXPIRED(1006, "Offer已过期，需要重新审批"),
    CANDIDATE_ALREADY_EXISTS(1007, "候选人已存在"),
    VALIDATION_ERROR(1008, "参数校验失败"),
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
