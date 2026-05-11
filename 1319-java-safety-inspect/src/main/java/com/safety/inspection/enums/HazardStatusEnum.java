package com.safety.inspection.enums;

import lombok.Getter;

@Getter
public enum HazardStatusEnum {
    PENDING_RECTIFICATION("PENDING_RECTIFICATION", "待整改"),
    RECTIFYING("RECTIFYING", "整改中"),
    PENDING_RECHECK("PENDING_RECHECK", "待复检"),
    RECHECKED("RECHECKED", "已复检"),
    CLOSED("CLOSED", "已关闭");

    private final String code;
    private final String desc;

    HazardStatusEnum(String code, String desc) {
        this.code = code;
        this.desc = desc;
    }

    public static HazardStatusEnum getByCode(String code) {
        for (HazardStatusEnum status : values()) {
            if (status.code.equals(code)) {
                return status;
            }
        }
        return null;
    }
}
