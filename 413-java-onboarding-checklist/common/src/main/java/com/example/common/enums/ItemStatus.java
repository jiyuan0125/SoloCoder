package com.example.common.enums;

public enum ItemStatus {
    PENDING("待处理"),
    IN_PROGRESS("进行中"),
    COMPLETED("已完成"),
    OVERDUE("已逾期"),
    ESCALATED("已升级");

    private final String description;

    ItemStatus(String description) {
        this.description = description;
    }

    public String getDescription() {
        return description;
    }
}
