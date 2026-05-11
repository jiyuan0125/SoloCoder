package com.retail.memberpoints.enums;

import java.math.BigDecimal;

public enum MemberLevel {
    NORMAL(new BigDecimal("0"), 1.0, "普通会员"),
    SILVER(new BigDecimal("1000"), 1.2, "银卡"),
    GOLD(new BigDecimal("5000"), 1.5, "金卡"),
    DIAMOND(new BigDecimal("20000"), 2.0, "钻石卡");

    private final BigDecimal threshold;
    private final double multiplier;
    private final String description;

    MemberLevel(BigDecimal threshold, double multiplier, String description) {
        this.threshold = threshold;
        this.multiplier = multiplier;
        this.description = description;
    }

    public BigDecimal getThreshold() {
        return threshold;
    }

    public double getMultiplier() {
        return multiplier;
    }

    public String getDescription() {
        return description;
    }

    public static MemberLevel calculateLevel(BigDecimal totalSpending) {
        if (totalSpending.compareTo(DIAMOND.threshold) >= 0) {
            return DIAMOND;
        } else if (totalSpending.compareTo(GOLD.threshold) >= 0) {
            return GOLD;
        } else if (totalSpending.compareTo(SILVER.threshold) >= 0) {
            return SILVER;
        }
        return NORMAL;
    }
}
