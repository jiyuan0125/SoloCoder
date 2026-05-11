package com.safety.inspection.enums;

import lombok.Getter;

@Getter
public enum NotificationTypeEnum {
    HAZARD_CREATED("HAZARD_CREATED", "隐患创建通知"),
    HAZARD_UPGRADED("HAZARD_UPGRADED", "隐患升级通知"),
    HAZARD_OVERDUE("HAZARD_OVERDUE", "隐患超期通知"),
    HAZARD_ASSIGNED("HAZARD_ASSIGNED", "隐患分配通知"),
    HAZARD_RECTIFIED("HAZARD_RECTIFIED", "整改完成通知"),
    HAZARD_RECHECK_NEEDED("HAZARD_RECHECK_NEEDED", "待复检通知"),
    HAZARD_RECHECK_PASSED("HAZARD_RECHECK_PASSED", "复检通过通知"),
    HAZARD_RECHECK_FAILED("HAZARD_RECHECK_FAILED", "复检不通过通知"),
    TASK_ASSIGNED("TASK_ASSIGNED", "任务分配通知"),
    TASK_MISSED("TASK_MISSED", "漏检通知");

    private final String code;
    private final String desc;

    NotificationTypeEnum(String code, String desc) {
        this.code = code;
        this.desc = desc;
    }
}
