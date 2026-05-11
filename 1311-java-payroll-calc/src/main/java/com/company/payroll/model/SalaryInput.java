package com.company.payroll.model;

import java.math.BigDecimal;
import java.util.ArrayList;
import java.util.List;

public class SalaryInput {
    private String employeeId;
    private int year;
    private int month;
    private BigDecimal baseSalary;
    private List<AdditionalIncome> additionalIncomes;
    private Integer actualWorkDays;
    private Integer totalWorkDaysInMonth;

    public SalaryInput() {
        this.additionalIncomes = new ArrayList<>();
    }

    public SalaryInput(String employeeId, int year, int month, BigDecimal baseSalary) {
        this();
        this.employeeId = employeeId;
        this.year = year;
        this.month = month;
        this.baseSalary = baseSalary;
    }

    public BigDecimal getTotalAdditionalIncome() {
        return additionalIncomes.stream()
                .filter(income -> income.getType() != IncomeType.YEAR_END_BONUS)
                .map(AdditionalIncome::getAmount)
                .reduce(BigDecimal.ZERO, BigDecimal::add);
    }

    public BigDecimal getYearEndBonus() {
        return additionalIncomes.stream()
                .filter(income -> income.getType() == IncomeType.YEAR_END_BONUS)
                .map(AdditionalIncome::getAmount)
                .reduce(BigDecimal.ZERO, BigDecimal::add);
    }

    public String getEmployeeId() {
        return employeeId;
    }

    public void setEmployeeId(String employeeId) {
        this.employeeId = employeeId;
    }

    public int getYear() {
        return year;
    }

    public void setYear(int year) {
        this.year = year;
    }

    public int getMonth() {
        return month;
    }

    public void setMonth(int month) {
        this.month = month;
    }

    public BigDecimal getBaseSalary() {
        return baseSalary;
    }

    public void setBaseSalary(BigDecimal baseSalary) {
        this.baseSalary = baseSalary;
    }

    public List<AdditionalIncome> getAdditionalIncomes() {
        return additionalIncomes;
    }

    public void setAdditionalIncomes(List<AdditionalIncome> additionalIncomes) {
        this.additionalIncomes = additionalIncomes;
    }

    public Integer getActualWorkDays() {
        return actualWorkDays;
    }

    public void setActualWorkDays(Integer actualWorkDays) {
        this.actualWorkDays = actualWorkDays;
    }

    public Integer getTotalWorkDaysInMonth() {
        return totalWorkDaysInMonth;
    }

    public void setTotalWorkDaysInMonth(Integer totalWorkDaysInMonth) {
        this.totalWorkDaysInMonth = totalWorkDaysInMonth;
    }
}
