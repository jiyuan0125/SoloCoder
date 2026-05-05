package com.payroll.server.entity;

public enum ArchiveStatus {
    PENDING("待确认"),
    CONFIRMED("已确认");

    private final String description;

    ArchiveStatus(String description) {
        this.description = description;
    }

    public String getDescription() {
        return description;
    }
}
