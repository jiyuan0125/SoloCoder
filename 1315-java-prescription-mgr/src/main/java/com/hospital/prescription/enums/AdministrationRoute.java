package com.hospital.prescription.enums;

public enum AdministrationRoute {
    ORAL("口服"),
    INTRAVENOUS("静脉注射"),
    INTRAMUSCULAR("肌肉注射"),
    SUBCUTANEOUS("皮下注射"),
    TOPICAL("外用");

    private final String description;

    AdministrationRoute(String description) {
        this.description = description;
    }

    public String getDescription() {
        return description;
    }
}
