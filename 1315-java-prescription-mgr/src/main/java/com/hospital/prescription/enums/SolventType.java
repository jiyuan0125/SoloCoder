package com.hospital.prescription.enums;

public enum SolventType {
    NORMAL_SALINE("生理盐水"),
    GLUCOSE("葡萄糖"),
    NONE("无需溶媒");

    private final String description;

    SolventType(String description) {
        this.description = description;
    }

    public String getDescription() {
        return description;
    }
}
