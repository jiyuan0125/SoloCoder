package com.payroll.common.dto;

import java.math.BigDecimal;

public class MonthlyComparisonDTO {

    private int month;
    private BigDecimal baseYearAmount;
    private BigDecimal compareYearAmount;
    private BigDecimal change;
    private BigDecimal changeRate;

    public MonthlyComparisonDTO() {
    }

    public int getMonth() {
        return month;
    }

    public void setMonth(int month) {
        this.month = month;
    }

    public BigDecimal getBaseYearAmount() {
        return baseYearAmount;
    }

    public void setBaseYearAmount(BigDecimal baseYearAmount) {
        this.baseYearAmount = baseYearAmount;
    }

    public BigDecimal getCompareYearAmount() {
        return compareYearAmount;
    }

    public void setCompareYearAmount(BigDecimal compareYearAmount) {
        this.compareYearAmount = compareYearAmount;
    }

    public BigDecimal getChange() {
        return change;
    }

    public void setChange(BigDecimal change) {
        this.change = change;
    }

    public BigDecimal getChangeRate() {
        return changeRate;
    }

    public void setChangeRate(BigDecimal changeRate) {
        this.changeRate = changeRate;
    }
}
