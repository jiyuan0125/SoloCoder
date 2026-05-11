package com.hotelbooking.model.enums;

public enum BookingStatus {
    PENDING_CONFIRMATION("待确认"),
    CONFIRMED("已确认"),
    CHECKED_IN("已入住"),
    CHECKED_OUT("已退房"),
    CANCELLED("已取消"),
    NO_SHOW("未入住");

    private final String displayName;

    BookingStatus(String displayName) {
        this.displayName = displayName;
    }

    public String getDisplayName() {
        return displayName;
    }
}
