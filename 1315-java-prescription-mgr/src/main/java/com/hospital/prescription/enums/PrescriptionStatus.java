package com.hospital.prescription.enums;

public enum PrescriptionStatus {
    DRAFT("草稿"),
    SUBMITTED("已提交"),
    UNDER_REVIEW("审核中"),
    APPROVED("已审核"),
    MODIFIED_BY_PHARMACIST("药师已修改"),
    PENDING_DOCTOR_CONFIRM("待医生确认"),
    CONSULTATION("会诊中"),
    DISPENSED("已发药"),
    COMPLETED("已完成"),
    REJECTED("已驳回");

    private final String description;

    PrescriptionStatus(String description) {
        this.description = description;
    }

    public String getDescription() {
        return description;
    }
}
