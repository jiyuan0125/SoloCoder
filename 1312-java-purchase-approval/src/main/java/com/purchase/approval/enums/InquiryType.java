package com.purchase.approval.enums;

public enum InquiryType {
    DIRECT("直接指定"),
    LIMITED("有限询价"),
    OPEN("公开询价");

    private final String description;

    InquiryType(String description) {
        this.description = description;
    }

    public String getDescription() {
        return description;
    }
}
