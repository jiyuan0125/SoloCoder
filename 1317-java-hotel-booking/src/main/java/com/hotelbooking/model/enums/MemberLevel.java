package com.hotelbooking.model.enums;

import java.math.BigDecimal;

public enum MemberLevel {
    NONE("普通会员", new BigDecimal("1.00")),
    SILVER("银卡会员", new BigDecimal("0.95")),
    GOLD("金卡会员", new BigDecimal("0.90")),
    PLATINUM("白金会员", new BigDecimal("0.85"));

    private final String displayName;
    private final BigDecimal discountRate;

    MemberLevel(String displayName, BigDecimal discountRate) {
        this.displayName = displayName;
        this.discountRate = discountRate;
    }

    public String getDisplayName() {
        return displayName;
    }

    public BigDecimal getDiscountRate() {
        return discountRate;
    }
}
