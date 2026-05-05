package com.payroll.common.dto;

import java.math.BigDecimal;

public class MonthlySummaryDTO {

    private int month;
    private BigDecimal totalExpense;
    private int recordCount;

    public MonthlySummaryDTO() {
    }

    public MonthlySummaryDTO(int month, BigDecimal totalExpense, int recordCount) {
        this.month = month;
        this.totalExpense = totalExpense;
        this.recordCount = recordCount;
    }

    public int getMonth() {
        return month;
    }

    public void setMonth(int month) {
        this.month = month;
    }

    public BigDecimal getTotalExpense() {
        return totalExpense;
    }

    public void setTotalExpense(BigDecimal totalExpense) {
        this.totalExpense = totalExpense;
    }

    public int getRecordCount() {
        return recordCount;
    }

    public void setRecordCount(int recordCount) {
        this.recordCount = recordCount;
    }
}
