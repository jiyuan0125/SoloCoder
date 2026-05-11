package com.company.expense.enums;

public enum ApprovalAction {
    APPROVE("通过"),
    REJECT("驳回");

    private final String description;

    ApprovalAction(String description) {
        this.description = description;
    }

    public String getDescription() {
        return description;
    }
}
