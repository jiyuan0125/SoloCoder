package com.hospital.prescription.enums;

public enum DrugCategory {
    NORMAL("普通药品", 1),
    PSYCHOTROPIC("精神类药品", 2),
    NARCOTIC("麻醉类药品", 2);

    private final String description;
    private final int requiredSignatures;

    DrugCategory(String description, int requiredSignatures) {
        this.description = description;
        this.requiredSignatures = requiredSignatures;
    }

    public String getDescription() {
        return description;
    }

    public int getRequiredSignatures() {
        return requiredSignatures;
    }
}
