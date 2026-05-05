package com.example.common.enums;

public enum AlertLevel {
    INFO("信息"),
    WARNING("警告"),
    ERROR("错误"),
    CRITICAL("严重");

    private final String description;

    AlertLevel(String description) {
        this.description = description;
    }

    public String getDescription() {
        return description;
    }
}
