package com.company.expense.enums;

public enum ApprovalLevel {
    SUPERVISOR("直属主管", 500.00),
    MANAGER("部门经理", 2000.00),
    GENERAL_MANAGER("总经理", Double.MAX_VALUE);

    private final String description;
    private final double maxAmount;

    ApprovalLevel(String description, double maxAmount) {
        this.description = description;
        this.maxAmount = maxAmount;
    }

    public String getDescription() {
        return description;
    }

    public double getMaxAmount() {
        return maxAmount;
    }

    public static ApprovalLevel getByAmount(double amount) {
        if (amount <= SUPERVISOR.getMaxAmount()) {
            return SUPERVISOR;
        } else if (amount <= MANAGER.getMaxAmount()) {
            return MANAGER;
        } else {
            return GENERAL_MANAGER;
        }
    }
}
