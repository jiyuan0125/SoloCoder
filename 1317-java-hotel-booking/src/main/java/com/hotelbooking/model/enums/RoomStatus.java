package com.hotelbooking.model.enums;

public enum RoomStatus {
    AVAILABLE("可预订"),
    OCCUPIED("已入住"),
    CLEANING("清洁中"),
    MAINTENANCE("维修中"),
    RESERVED("已预订");

    private final String displayName;

    RoomStatus(String displayName) {
        this.displayName = displayName;
    }

    public String getDisplayName() {
        return displayName;
    }
}
