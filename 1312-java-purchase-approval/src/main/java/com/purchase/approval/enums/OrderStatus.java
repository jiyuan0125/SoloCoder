package com.purchase.approval.enums;

public enum OrderStatus {
    PENDING_CONFIRMATION("待确认"),
    CONFIRMED("已确认"),
    SHIPPED("已发货"),
    RECEIVED("已收货"),
    INSPECTED("已验收"),
    COMPLETED("已完成"),
    RETURNED("已退货"),
    CANCELLED("已取消");

    private final String description;

    OrderStatus(String description) {
        this.description = description;
    }

    public String getDescription() {
        return description;
    }
}
