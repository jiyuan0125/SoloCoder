package com.delivery.enums;

import lombok.Getter;

@Getter
public enum TimeSlot {
    MORNING("上午(9-12点)", 9, 12),
    AFTERNOON("下午(13-18点)", 13, 18),
    EVENING("晚上(18-21点)", 18, 21);

    private final String description;
    private final int startHour;
    private final int endHour;

    TimeSlot(String description, int startHour, int endHour) {
        this.description = description;
        this.startHour = startHour;
        this.endHour = endHour;
    }
}
