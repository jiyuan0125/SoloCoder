package com.delivery.enums;

import lombok.Getter;

@Getter
public enum OrderStatus {
    PENDING("待处理"),
    CONFIRMED("已确认"),
    DISPATCHED("已出库"),
    IN_TRANSIT("运输中"),
    DELIVERED("已送达"),
    CANCELLED("已取消"),
    RETURNED("已退回");

    private final String description;

    OrderStatus(String description) {
        this.description = description;
    }
}
