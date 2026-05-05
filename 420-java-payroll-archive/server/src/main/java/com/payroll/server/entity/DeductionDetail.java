package com.payroll.server.entity;

import java.math.BigDecimal;

public class DeductionDetail {

    private String deductionType;
    private BigDecimal amount;
    private String description;

    public DeductionDetail() {
    }

    public DeductionDetail(String deductionType, BigDecimal amount, String description) {
        this.deductionType = deductionType;
        this.amount = amount;
        this.description = description;
    }

    public String getDeductionType() {
        return deductionType;
    }

    public void setDeductionType(String deductionType) {
        this.deductionType = deductionType;
    }

    public BigDecimal getAmount() {
        return amount;
    }

    public void setAmount(BigDecimal amount) {
        this.amount = amount;
    }

    public String getDescription() {
        return description;
    }

    public void setDescription(String description) {
        this.description = description;
    }
}
