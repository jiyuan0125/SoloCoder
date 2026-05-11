package com.hospital.enums;

public enum DoctorTitle {
    ORDINARY(20, 10, 10),
    EXPERT(10, 5, 5);

    private final int dailySlots;
    private final int morningSlots;
    private final int afternoonSlots;

    DoctorTitle(int dailySlots, int morningSlots, int afternoonSlots) {
        this.dailySlots = dailySlots;
        this.morningSlots = morningSlots;
        this.afternoonSlots = afternoonSlots;
    }

    public int getDailySlots() {
        return dailySlots;
    }

    public int getMorningSlots() {
        return morningSlots;
    }

    public int getAfternoonSlots() {
        return afternoonSlots;
    }
}
