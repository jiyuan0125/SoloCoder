package com.recruitment.common.enums;

public enum ApplicationStatus {
    IN_PROGRESS("进行中"),
    REJECTED("已拒绝"),
    ABANDONED("已放弃"),
    OFFER_ACCEPTED("已接受Offer"),
    ONBOARDED("已入职");

    private final String description;

    ApplicationStatus(String description) {
        this.description = description;
    }

    public String getDescription() {
        return description;
    }
}
