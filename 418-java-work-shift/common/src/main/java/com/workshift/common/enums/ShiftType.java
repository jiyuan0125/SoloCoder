package com.workshift.common.enums;

public enum ShiftType {
    MORNING("早班", 8, 16),
    AFTERNOON("中班", 16, 24),
    NIGHT("夜班", 0, 8);

    private final String description;
    private final int startHour;
    private final int endHour;

    ShiftType(String description, int startHour, int endHour) {
        this.description = description;
        this.startHour = startHour;
        this.endHour = endHour;
    }

    public String getDescription() {
        return description;
    }

    public int getStartHour() {
        return startHour;
    }

    public int getEndHour() {
        return endHour;
    }
}