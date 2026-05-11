package com.delivery.enums;

import lombok.Getter;

@Getter
public enum CompensationType {
    NONE("无赔偿", 0.0),
    WITHIN_2_HOURS("超时2小时内", 0.3),
    WITHIN_12_HOURS("超时2-12小时", 0.6),
    OVER_12_HOURS("超时12小时以上", 1.0);

    private final String description;
    private final double refundRate;

    CompensationType(String description, double refundRate) {
        this.description = description;
        this.refundRate = refundRate;
    }
}
