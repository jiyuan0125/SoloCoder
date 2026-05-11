package com.company.payroll.model;

import java.math.BigDecimal;

public enum SpecialDeductionType {
    CHILDREN_EDUCATION("子女教育", new BigDecimal("1000")),
    ELDERLY_SUPPORT("赡养老人", new BigDecimal("2000")),
    HOUSING_LOAN_INTEREST("住房贷款利息", new BigDecimal("1000")),
    HOUSING_RENT("住房租金", new BigDecimal("1500")),
    CONTINUING_EDUCATION("继续教育", new BigDecimal("400"));

    private final String description;
    private final BigDecimal monthlyAmount;

    SpecialDeductionType(String description, BigDecimal monthlyAmount) {
        this.description = description;
        this.monthlyAmount = monthlyAmount;
    }

    public String getDescription() {
        return description;
    }

    public BigDecimal getMonthlyAmount() {
        return monthlyAmount;
    }
}