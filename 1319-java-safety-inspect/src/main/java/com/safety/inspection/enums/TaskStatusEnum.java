package com.safety.inspection.enums;

import lombok.Getter;

@Getter
public enum TaskStatusEnum {
    PENDING("PENDING", "待执行"),
    IN_PROGRESS("IN_PROGRESS", "进行中"),
    COMPLETED("COMPLETED", "已完成"),
    MISSED("MISSED", "漏检");

    private final String code;
    private final String desc;

    TaskStatusEnum(String code, String desc) {
        this.code = code;
        this.desc = desc;
    }
}
