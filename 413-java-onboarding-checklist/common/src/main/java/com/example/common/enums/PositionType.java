package com.example.common.enums;

public enum PositionType {
    DEVELOPER("开发岗"),
    SALES("销售岗"),
    HR("人事岗"),
    FINANCE("财务岗"),
    ADMIN("行政岗"),
    MARKETING("市场岗");

    private final String description;

    PositionType(String description) {
        this.description = description;
    }

    public String getDescription() {
        return description;
    }
}
