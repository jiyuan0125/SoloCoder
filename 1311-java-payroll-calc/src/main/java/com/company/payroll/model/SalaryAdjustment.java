package com.company.payroll.model;

import java.math.BigDecimal;
import java.time.YearMonth;

public class SalaryAdjustment {
    private String id;
    private String employeeId;
    private BigDecimal newSalary;
    private YearMonth effectiveMonth;
    private String reason;

    public SalaryAdjustment() {}

    public SalaryAdjustment(String id, String employeeId, BigDecimal newSalary, 
                            YearMonth effectiveMonth, String reason) {
        this.id = id;
        this.employeeId = employeeId;
        this.newSalary = newSalary;
        this.effectiveMonth = effectiveMonth;
        this.reason = reason;
    }

    public boolean isEffectiveAt(int year, int month) {
        YearMonth target = YearMonth.of(year, month);
        return !target.isBefore(effectiveMonth);
    }

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public String getEmployeeId() {
        return employeeId;
    }

    public void setEmployeeId(String employeeId) {
        this.employeeId = employeeId;
    }

    public BigDecimal getNewSalary() {
        return newSalary;
    }

    public void setNewSalary(BigDecimal newSalary) {
        this.newSalary = newSalary;
    }

    public YearMonth getEffectiveMonth() {
        return effectiveMonth;
    }

    public void setEffectiveMonth(YearMonth effectiveMonth) {
        this.effectiveMonth = effectiveMonth;
    }

    public String getReason() {
        return reason;
    }

    public void setReason(String reason) {
        this.reason = reason;
    }
}
