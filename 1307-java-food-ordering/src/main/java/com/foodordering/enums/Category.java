package com.foodordering.enums;

public enum Category {
    STAPLE("主食"),
    SNACK("小食"),
    DRINK("饮品"),
    DESSERT("甜品");

    private final String displayName;

    Category(String displayName) {
        this.displayName = displayName;
    }

    public String getDisplayName() {
        return displayName;
    }
}
