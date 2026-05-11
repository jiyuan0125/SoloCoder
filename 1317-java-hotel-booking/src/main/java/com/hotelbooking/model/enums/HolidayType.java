package com.hotelbooking.model.enums;

public enum HolidayType {
    NORMAL("平日"),
    HOLIDAY_EVE("节假日前一天"),
    HOLIDAY("节假日"),
    HOLIDAY_AFTER("节假日后一天");

    private final String displayName;

    HolidayType(String displayName) {
        this.displayName = displayName;
    }

    public String getDisplayName() {
        return displayName;
    }
}
