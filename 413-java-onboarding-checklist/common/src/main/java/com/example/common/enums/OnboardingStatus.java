package com.example.common.enums;

public enum OnboardingStatus {
    COMPLETE("完整"),
    INCOMPLETE("不完整"),
    IN_PROGRESS("进行中");

    private final String description;

    OnboardingStatus(String description) {
        this.description = description;
    }

    public String getDescription() {
        return description;
    }
}
