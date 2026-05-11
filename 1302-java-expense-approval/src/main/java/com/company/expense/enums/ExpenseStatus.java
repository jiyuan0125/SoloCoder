package com.company.expense.enums;

public enum ExpenseStatus {
    PENDING("待审批"),
    APPROVING("审批中"),
    APPROVED("已通过"),
    REJECTED("已驳回");

    private final String description;

    ExpenseStatus(String description) {
        this.description = description;
    }

    public String getDescription() {
        return description;
    }
}
