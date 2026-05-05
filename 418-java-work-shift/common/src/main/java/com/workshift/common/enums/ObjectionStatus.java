package com.workshift.common.enums;

public enum ObjectionStatus {
    PENDING("待处理"),
    RESOLVED("已解决"),
    DISMISSED("已驳回");

    private final String description;

    ObjectionStatus(String description) {
        this.description = description;
    }

    public String getDescription() {
        return description;
    }
}