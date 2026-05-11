package com.foodordering.enums;

public enum TasteOption {
    SPICY_LIGHT("微辣"),
    SPICY_MEDIUM("中辣"),
    SPICY_HEAVY("重辣"),
    SUGAR_LESS("少糖"),
    SUGAR_NORMAL("正常糖"),
    SUGAR_MORE("多糖"),
    NONE("无");

    private final String displayName;

    TasteOption(String displayName) {
        this.displayName = displayName;
    }

    public String getDisplayName() {
        return displayName;
    }
}
