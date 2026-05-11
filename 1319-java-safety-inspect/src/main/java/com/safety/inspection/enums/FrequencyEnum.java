package com.safety.inspection.enums;

import lombok.Getter;

@Getter
public enum FrequencyEnum {
    DAILY("DAILY", "每天", 1),
    WEEKLY("WEEKLY", "每周", 7),
    MONTHLY("MONTHLY", "每月", 30),
    CUSTOM("CUSTOM", "自定义", null);

    private final String code;
    private final String desc;
    private final Integer days;

    FrequencyEnum(String code, String desc, Integer days) {
        this.code = code;
        this.desc = desc;
        this.days = days;
    }
}
