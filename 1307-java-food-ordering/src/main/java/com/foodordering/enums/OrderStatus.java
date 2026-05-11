package com.foodordering.enums;

public enum OrderStatus {
    PENDING_PAYMENT("待支付"),
    PAID("已支付"),
    PREPARING("制作中"),
    COMPLETED("已完成"),
    CANCELLED("已取消");

    private final String displayName;

    OrderStatus(String displayName) {
        this.displayName = displayName;
    }

    public String getDisplayName() {
        return displayName;
    }
}
