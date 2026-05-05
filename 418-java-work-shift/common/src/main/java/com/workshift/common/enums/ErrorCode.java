package com.workshift.common.enums;

public enum ErrorCode {
    SUCCESS(0, "操作成功"),
    INVALID_PARAMETER(10001, "参数无效"),
    EMPLOYEE_NOT_FOUND(10002, "员工不存在"),
    SHIFT_NOT_FOUND(10003, "排班记录不存在"),
    SWAP_NOT_FOUND(10004, "换班申请不存在"),
    OBJECTION_NOT_FOUND(10005, "异议记录不存在"),
    
    CONSECUTIVE_NIGHT_SHIFT_VIOLATION(20001, "连续夜班超过3天，违反排班规则"),
    SHIFT_INTERVAL_VIOLATION(20002, "相邻班次间隔不足8小时，违反排班规则"),
    NIGHT_TO_MORNING_VIOLATION(20003, "夜班接早班属于违规排班"),
    WEEKLY_REST_DAY_VIOLATION(20004, "每周至少需要1天休息日"),
    CONSECUTIVE_WORK_DAY_VIOLATION(20005, "连续工作天数超过6天"),
    SHIFT_ALREADY_EXISTS(20006, "该日期已有排班记录"),
    SHIFT_SWAP_RULE_VIOLATION(20007, "换班后违反排班规则"),
    
    OBJECTION_TIME_EXPIRED(30001, "异议期限已过（超过24小时）"),
    OBJECTION_ALREADY_SUBMITTED(30002, "已提交过异议"),
    SHIFT_NOT_PUBLISHED(30003, "排班尚未发布"),
    SWAP_ALREADY_PROCESSED(30004, "换班申请已处理"),
    CANNOT_SWAP_SELF(30005, "不能与自己换班"),
    TARGET_SHIFT_NOT_FOUND(30006, "目标日期无排班记录"),
    SWAP_CONFIRM_FAILED(30007, "换班确认失败"),
    
    HOLIDAY_CONFIG_NOT_FOUND(40001, "节假日配置不存在"),
    STATISTICS_GENERATION_FAILED(40002, "统计报表生成失败"),
    INVALID_DATE_RANGE(40003, "日期范围无效"),
    
    INTERNAL_ERROR(50000, "系统内部错误");

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