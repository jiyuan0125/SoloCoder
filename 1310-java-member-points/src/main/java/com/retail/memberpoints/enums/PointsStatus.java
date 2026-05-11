package com.retail.memberpoints.enums;

public enum PointsStatus {
    AVAILABLE("可用"),
    USED("已使用"),
    EXPIRED("已过期");

    private final String description;

    PointsStatus(String description) {
        this.description = description;
    }

    public String getDescription() {
        return description;
    }
}
