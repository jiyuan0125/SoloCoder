package com.safety.inspection.enums;

import lombok.Getter;

@Getter
public enum HazardLevelEnum {
    GENERAL("GENERAL", "一般隐患", 7),
    LARGER("LARGER", "较大隐患", 3),
    MAJOR("MAJOR", "重大隐患", 0);

    private final String code;
    private final String desc;
    private final Integer deadlineDays;

    HazardLevelEnum(String code, String desc, Integer deadlineDays) {
        this.code = code;
        this.desc = desc;
        this.deadlineDays = deadlineDays;
    }

    public static HazardLevelEnum getByCode(String code) {
        for (HazardLevelEnum level : values()) {
            if (level.code.equals(code)) {
                return level;
            }
        }
        return null;
    }

    public HazardLevelEnum getNextLevel() {
        switch (this) {
            case GENERAL:
                return LARGER;
            case LARGER:
                return MAJOR;
            default:
                return null;
        }
    }
}
