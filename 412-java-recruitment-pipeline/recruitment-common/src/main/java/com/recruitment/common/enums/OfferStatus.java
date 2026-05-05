package com.recruitment.common.enums;

public enum OfferStatus {
    PENDING("待回复"),
    ACCEPTED("已接受"),
    REJECTED("已拒绝"),
    EXPIRED("已过期");

    private final String description;

    OfferStatus(String description) {
        this.description = description;
    }

    public String getDescription() {
        return description;
    }
}
