package com.payroll.common.dto;

import java.math.BigDecimal;
import java.util.List;

public class AnnualReportDTO {

    private int year;
    private BigDecimal totalExpense;
    private BigDecimal averageSalary;
    private BigDecimal maxSalary;
    private BigDecimal minSalary;
    private int totalRecords;
    private int totalEmployees;
    private BigDecimal yearOverYearChange;
    private List<MonthlySummaryDTO> monthlySummaries;

    public AnnualReportDTO() {
    }

    public int getYear() {
        return year;
    }

    public void setYear(int year) {
        this.year = year;
    }

    public BigDecimal getTotalExpense() {
        return totalExpense;
    }

    public void setTotalExpense(BigDecimal totalExpense) {
        this.totalExpense = totalExpense;
    }

    public BigDecimal getAverageSalary() {
        return averageSalary;
    }

    public void setAverageSalary(BigDecimal averageSalary) {
        this.averageSalary = averageSalary;
    }

    public BigDecimal getMaxSalary() {
        return maxSalary;
    }

    public void setMaxSalary(BigDecimal maxSalary) {
        this.maxSalary = maxSalary;
    }

    public BigDecimal getMinSalary() {
        return minSalary;
    }

    public void setMinSalary(BigDecimal minSalary) {
        this.minSalary = minSalary;
    }

    public int getTotalRecords() {
        return totalRecords;
    }

    public void setTotalRecords(int totalRecords) {
        this.totalRecords = totalRecords;
    }

    public int getTotalEmployees() {
        return totalEmployees;
    }

    public void setTotalEmployees(int totalEmployees) {
        this.totalEmployees = totalEmployees;
    }

    public BigDecimal getYearOverYearChange() {
        return yearOverYearChange;
    }

    public void setYearOverYearChange(BigDecimal yearOverYearChange) {
        this.yearOverYearChange = yearOverYearChange;
    }

    public List<MonthlySummaryDTO> getMonthlySummaries() {
        return monthlySummaries;
    }

    public void setMonthlySummaries(List<MonthlySummaryDTO> monthlySummaries) {
        this.monthlySummaries = monthlySummaries;
    }
}
