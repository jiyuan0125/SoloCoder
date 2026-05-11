package com.retail.memberpoints.enums;

public enum TransactionType {
    EARN("积分获取"),
    USE("积分使用"),
    DEDUCT("积分扣回"),
    EXPIRED("积分过期");

    private final String description;

    TransactionType(String description) {
        this.description = description;
    }

    public String getDescription() {
        return description;
    }
}
