package com.hospital.prescription.enums;

public enum DosageForm {
    TABLET("片剂"),
    CAPSULE("胶囊"),
    INJECTION("注射液"),
    SYRUP("糖浆"),
    POWDER("粉剂"),
    SUPPOSITORY("栓剂"),
    OINTMENT("软膏"),
    DROPS("滴剂"),
    GRANULE("颗粒剂");

    private final String description;

    DosageForm(String description) {
        this.description = description;
    }

    public String getDescription() {
        return description;
    }
}
