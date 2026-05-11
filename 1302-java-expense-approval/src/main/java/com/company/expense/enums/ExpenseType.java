package com.company.expense.enums;

public enum ExpenseType {
    TRAVEL("差旅费"),
    OFFICE_SUPPLIES("办公用品"),
    TRANSPORTATION("交通费"),
    MEAL("餐饮费");

    private final String description;

    ExpenseType(String description) {
        this.description = description;
    }

    public String getDescription() {
        return description;
    }
}
