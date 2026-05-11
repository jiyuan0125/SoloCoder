package com.hotelbooking.model.enums;

public enum Role {
    ADMIN("管理员"),
    RECEPTION("前台"),
    HOUSEKEEPING("清洁人员"),
    CUSTOMER("客户");

    private final String displayName;

    Role(String displayName) {
        this.displayName = displayName;
    }

    public String getDisplayName() {
        return displayName;
    }
}
