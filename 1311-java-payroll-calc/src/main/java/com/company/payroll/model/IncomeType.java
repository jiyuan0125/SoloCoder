package com.company.payroll.model;

public enum IncomeType {
    PERFORMANCE_BONUS("绩效奖金"),
    OVERTIME_PAY("加班费"),
    PROJECT_BONUS("项目奖金"),
    PART_TIME_INCOME("兼职收入"),
    TRANSPORTATION_ALLOWANCE("交通补贴"),
    COMMUNICATION_ALLOWANCE("通讯补贴"),
    MEAL_ALLOWANCE("餐补"),
    YEAR_END_BONUS("年终奖");

    private final String description;

    IncomeType(String description) {
        this.description = description;
    }

    public String getDescription() {
        return description;
    }
}
