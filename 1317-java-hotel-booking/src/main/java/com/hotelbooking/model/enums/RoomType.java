package com.hotelbooking.model.enums;

public enum RoomType {
    STANDARD_SINGLE("标准单人间"),
    STANDARD_DOUBLE("标准双人间"),
    KING_BED("大床房"),
    FAMILY_ROOM("家庭房"),
    SUITE("套房");

    private final String displayName;

    RoomType(String displayName) {
        this.displayName = displayName;
    }

    public String getDisplayName() {
        return displayName;
    }
}
