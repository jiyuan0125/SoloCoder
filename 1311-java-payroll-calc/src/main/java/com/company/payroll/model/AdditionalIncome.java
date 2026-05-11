package com.company.payroll.model;

import java.math.BigDecimal;

public class AdditionalIncome {
    private String id;
    private String description;
    private BigDecimal amount;
    private IncomeType type;

    public AdditionalIncome() {}

    public AdditionalIncome(String id, String description, BigDecimal amount, IncomeType type) {
        this.id = id;
        this.description = description;
        this.amount = amount;
        this.type = type;
    }

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public String getDescription() {
        return description;
    }

    public void setDescription(String description) {
        this.description = description;
    }

    public BigDecimal getAmount() {
        return amount;
    }

    public void setAmount(BigDecimal amount) {
        this.amount = amount;
    }

    public IncomeType getType() {
        return type;
    }

    public void setType(IncomeType type) {
        this.type = type;
    }
}
