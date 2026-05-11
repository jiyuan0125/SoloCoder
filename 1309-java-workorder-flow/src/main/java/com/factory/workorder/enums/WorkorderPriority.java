package com.factory.workorder.enums;

import java.time.Duration;

public enum WorkorderPriority {
    NORMAL("普通", Duration.ofHours(48)),
    URGENT("紧急", Duration.ofHours(24)),
    CRITICAL("特急", Duration.ofHours(4));

    private final String description;
    private final Duration timeLimit;

    WorkorderPriority(String description, Duration timeLimit) {
        this.description = description;
        this.timeLimit = timeLimit;
    }

    public String getDescription() {
        return description;
    }

    public Duration getTimeLimit() {
        return timeLimit;
    }

    public WorkorderPriority upgrade() {
        switch (this) {
            case NORMAL:
                return URGENT;
            case URGENT:
                return CRITICAL;
            case CRITICAL:
            default:
                return CRITICAL;
        }
    }
}
