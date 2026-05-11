package com.purchase.approval.enums;

public enum ApprovalLevel {
    DEPARTMENT_MANAGER("部门经理", 1),
    FINANCE_DIRECTOR("财务总监", 2),
    GENERAL_MANAGER("总经理", 3);

    private final String description;
    private final int level;

    ApprovalLevel(String description, int level) {
        this.description = description;
        this.level = level;
    }

    public String getDescription() {
        return description;
    }

    public int getLevel() {
        return level;
    }
}
