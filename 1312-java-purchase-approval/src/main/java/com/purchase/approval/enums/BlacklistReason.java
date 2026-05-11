package com.purchase.approval.enums;

public enum BlacklistReason {
    SERIOUS_BREACH("严重违约"),
    QUALITY_ISSUE("质量问题"),
    COMMERCIAL_BRIBERY("商业贿赂");

    private final String description;

    BlacklistReason(String description) {
        this.description = description;
    }

    public String getDescription() {
        return description;
    }
}
