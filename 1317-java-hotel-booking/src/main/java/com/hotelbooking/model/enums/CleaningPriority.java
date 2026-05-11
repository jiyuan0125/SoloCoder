package com.hotelbooking.model.enums;

public enum CleaningPriority {
    NORMAL("普通", 2, 4),
    RUSH("加急", 1, 1);

    private final String displayName;
    private final int minHours;
    private final int maxHours;

    CleaningPriority(String displayName, int minHours, int maxHours) {
        this.displayName = displayName;
        this.minHours = minHours;
        this.maxHours = maxHours;
    }

    public String getDisplayName() {
        return displayName;
    }

    public int getMinHours() {
        return minHours;
    }

    public int getMaxHours() {
        return maxHours;
    }
}
