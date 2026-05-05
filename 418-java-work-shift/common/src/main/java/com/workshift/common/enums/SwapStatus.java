package com.workshift.common.enums;

public enum SwapStatus {
    PENDING("待确认"),
    CONFIRMED("已确认"),
    REJECTED("已拒绝"),
    CANCELLED("已取消");

    private final String description;

    SwapStatus(String description) {
        this.description = description;
    }

    public String getDescription() {
        return description;
    }
}