package com.company.payroll.model;

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.util.ArrayList;
import java.util.List;

public class DepartmentSalarySummary {
    private String departmentId;
    private String departmentName;
    private int year;
    private int month;
    private int employeeCount;
    private BigDecimal totalGrossSalary;
    private BigDecimal totalNetSalary;
    private BigDecimal totalTax;
    private BigDecimal totalSocialSecurity;
    private List<PaySlip> paySlips;

    public DepartmentSalarySummary() {
        this.totalGrossSalary = BigDecimal.ZERO;
        this.totalNetSalary = BigDecimal.ZERO;
        this.totalTax = BigDecimal.ZERO;
        this.totalSocialSecurity = BigDecimal.ZERO;
        this.paySlips = new ArrayList<>();
    }

    public void addPaySlip(PaySlip paySlip) {
        this.paySlips.add(paySlip);
        this.employeeCount++;
        this.totalGrossSalary = this.totalGrossSalary.add(paySlip.getGrossSalary());
        this.totalNetSalary = this.totalNetSalary.add(paySlip.getNetSalary());
        this.totalTax = this.totalTax.add(paySlip.getCurrentMonthTax());
        this.totalSocialSecurity = this.totalSocialSecurity.add(paySlip.getTotalSocialSecurityAndHousingFund());
    }

    public BigDecimal getAverageGrossSalary() {
        if (employeeCount == 0) {
            return BigDecimal.ZERO;
        }
        return totalGrossSalary.divide(new BigDecimal(employeeCount), 2, RoundingMode.HALF_UP);
    }

    public String getDepartmentId() {
        return departmentId;
    }

    public void setDepartmentId(String departmentId) {
        this.departmentId = departmentId;
    }

    public String getDepartmentName() {
        return departmentName;
    }

    public void setDepartmentName(String departmentName) {
        this.departmentName = departmentName;
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

    public int getEmployeeCount() {
        return employeeCount;
    }

    public void setEmployeeCount(int employeeCount) {
        this.employeeCount = employeeCount;
    }

    public BigDecimal getTotalGrossSalary() {
        return totalGrossSalary;
    }

    public void setTotalGrossSalary(BigDecimal totalGrossSalary) {
        this.totalGrossSalary = totalGrossSalary;
    }

    public BigDecimal getTotalNetSalary() {
        return totalNetSalary;
    }

    public void setTotalNetSalary(BigDecimal totalNetSalary) {
        this.totalNetSalary = totalNetSalary;
    }

    public BigDecimal getTotalTax() {
        return totalTax;
    }

    public void setTotalTax(BigDecimal totalTax) {
        this.totalTax = totalTax;
    }

    public BigDecimal getTotalSocialSecurity() {
        return totalSocialSecurity;
    }

    public void setTotalSocialSecurity(BigDecimal totalSocialSecurity) {
        this.totalSocialSecurity = totalSocialSecurity;
    }

    public List<PaySlip> getPaySlips() {
        return paySlips;
    }

    public void setPaySlips(List<PaySlip> paySlips) {
        this.paySlips = paySlips;
    }
}
