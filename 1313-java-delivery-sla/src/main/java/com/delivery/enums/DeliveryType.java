package com.delivery.enums;

import lombok.Getter;

@Getter
public enum DeliveryType {
    NEXT_DAY("次日达", 0.50, 16, 1),
    TWO_DAY("隔日达", 0.20, 16, 2),
    STANDARD("普通配送", 0.00, 18, 5);

    private final String description;
    private final double surchargeRate;
    private final int defaultCutOffHour;
    private final int maxDeliveryDays;

    DeliveryType(String description, double surchargeRate, int defaultCutOffHour, int maxDeliveryDays) {
        this.description = description;
        this.surchargeRate = surchargeRate;
        this.defaultCutOffHour = defaultCutOffHour;
        this.maxDeliveryDays = maxDeliveryDays;
    }
}
