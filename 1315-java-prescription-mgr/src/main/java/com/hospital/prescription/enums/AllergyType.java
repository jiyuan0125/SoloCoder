package com.hospital.prescription.enums;

public enum AllergyType {
    CONFIRMED("确认过敏", true),
    SUSPECTED("疑似过敏", false);

    private final String description;
    private final boolean mustBlock;

    AllergyType(String description, boolean mustBlock) {
        this.description = description;
        this.mustBlock = mustBlock;
    }

    public String getDescription() {
        return description;
    }

    public boolean isMustBlock() {
        return mustBlock;
    }
}
