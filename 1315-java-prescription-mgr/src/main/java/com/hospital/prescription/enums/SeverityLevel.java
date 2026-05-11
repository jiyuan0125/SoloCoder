package com.hospital.prescription.enums;

public enum SeverityLevel {
    SEVERE("严重", true),
    MODERATE("中等", false),
    MILD("轻度", false);

    private final String description;
    private final boolean mustBlock;

    SeverityLevel(String description, boolean mustBlock) {
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
