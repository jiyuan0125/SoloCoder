package com.performance.common.enums;

public enum ReviewStatus {
    DRAFT("草稿"),
    SELF_REVIEW("员工自评"),
    MANAGER_REVIEW("上级评分"),
    HR_CONFIRM("HR确认"),
    COMPLETED("已完成");

    private final String description;

    ReviewStatus(String description) {
        this.description = description;
    }

    public String getDescription() {
        return description;
    }
}