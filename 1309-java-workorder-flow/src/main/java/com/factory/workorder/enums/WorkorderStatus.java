package com.factory.workorder.enums;

public enum WorkorderStatus {
    PENDING_ASSIGN("待分派"),
    IN_PROGRESS("处理中"),
    PENDING_ACCEPTANCE("待验收"),
    COMPLETED("已完成");

    private final String description;

    WorkorderStatus(String description) {
        this.description = description;
    }

    public String getDescription() {
        return description;
    }
}
