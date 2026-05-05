package com.payroll.common.dto;

import java.math.BigDecimal;
import java.util.List;

public class YearComparisonDTO {

    private int baseYear;
    private int compareYear;
    private BigDecimal baseYearTotal;
    private BigDecimal compareYearTotal;
    private BigDecimal totalChange;
    private BigDecimal totalChangeRate;
    private List<MonthlyComparisonDTO> monthlyComparisons;

    public YearComparisonDTO() {
    }

    public int getBaseYear() {
        return baseYear;
    }

    public void setBaseYear(int baseYear) {
        this.baseYear = baseYear;
    }

    public int getCompareYear() {
        return compareYear;
    }

    public void setCompareYear(int compareYear) {
        this.compareYear = compareYear;
    }

    public BigDecimal getBaseYearTotal() {
        return baseYearTotal;
    }

    public void setBaseYearTotal(BigDecimal baseYearTotal) {
        this.baseYearTotal = baseYearTotal;
    }

    public BigDecimal getCompareYearTotal() {
        return compareYearTotal;
    }

    public void setCompareYearTotal(BigDecimal compareYearTotal) {
        this.compareYearTotal = compareYearTotal;
    }

    public BigDecimal getTotalChange() {
        return totalChange;
    }

    public void setTotalChange(BigDecimal totalChange) {
        this.totalChange = totalChange;
    }

    public BigDecimal getTotalChangeRate() {
        return totalChangeRate;
    }

    public void setTotalChangeRate(BigDecimal totalChangeRate) {
        this.totalChangeRate = totalChangeRate;
    }

    public List<MonthlyComparisonDTO> getMonthlyComparisons() {
        return monthlyComparisons;
    }

    public void setMonthlyComparisons(List<MonthlyComparisonDTO> monthlyComparisons) {
        this.monthlyComparisons = monthlyComparisons;
    }
}
