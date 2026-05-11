package com.company.payroll.model;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.util.ArrayList;
import java.util.List;

public class Employee {
    private String id;
    private String name;
    private String departmentId;
    private BigDecimal baseSalary;
    private boolean isProbation;
    private BigDecimal probationRatio;
    private LocalDate hireDate;
    private LocalDate terminationDate;
    private List<SpecialDeductionType> specialDeductions;
    private List<SalaryAdjustment> salaryAdjustments;

    public Employee() {
        this.probationRatio = new BigDecimal("0.80");
        this.specialDeductions = new ArrayList<>();
        this.salaryAdjustments = new ArrayList<>();
    }

    public Employee(String id, String name, String departmentId, BigDecimal baseSalary, 
                    boolean isProbation, LocalDate hireDate) {
        this();
        this.id = id;
        this.name = name;
        this.departmentId = departmentId;
        this.baseSalary = baseSalary;
        this.isProbation = isProbation;
        this.hireDate = hireDate;
    }

    public BigDecimal getEffectiveSalary(int year, int month) {
        BigDecimal salary = getAdjustedSalary(year, month);
        if (isProbation && probationRatio != null) {
            return salary.multiply(probationRatio);
        }
        return salary;
    }

    public BigDecimal getAdjustedSalary(int year, int month) {
        BigDecimal currentSalary = baseSalary;
        for (SalaryAdjustment adjustment : salaryAdjustments) {
            if (adjustment.isEffectiveAt(year, month)) {
                currentSalary = adjustment.getNewSalary();
            }
        }
        return currentSalary;
    }

    public BigDecimal getSocialSecurityBaseSalary(int year, int month) {
        return getAdjustedSalary(year, month);
    }

    public boolean isActiveAt(int year, int month) {
        LocalDate firstDayOfMonth = LocalDate.of(year, month, 1);
        LocalDate lastDayOfMonth = firstDayOfMonth.withDayOfMonth(firstDayOfMonth.lengthOfMonth());
        
        if (hireDate.isAfter(lastDayOfMonth)) {
            return false;
        }
        if (terminationDate != null && terminationDate.isBefore(firstDayOfMonth)) {
            return false;
        }
        return true;
    }

    public boolean isTerminatedThisMonth(int year, int month) {
        if (terminationDate == null) {
            return false;
        }
        return terminationDate.getYear() == year && terminationDate.getMonthValue() == month;
    }

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getDepartmentId() {
        return departmentId;
    }

    public void setDepartmentId(String departmentId) {
        this.departmentId = departmentId;
    }

    public BigDecimal getBaseSalary() {
        return baseSalary;
    }

    public void setBaseSalary(BigDecimal baseSalary) {
        this.baseSalary = baseSalary;
    }

    public boolean isProbation() {
        return isProbation;
    }

    public void setProbation(boolean probation) {
        isProbation = probation;
    }

    public BigDecimal getProbationRatio() {
        return probationRatio;
    }

    public void setProbationRatio(BigDecimal probationRatio) {
        this.probationRatio = probationRatio;
    }

    public LocalDate getHireDate() {
        return hireDate;
    }

    public void setHireDate(LocalDate hireDate) {
        this.hireDate = hireDate;
    }

    public LocalDate getTerminationDate() {
        return terminationDate;
    }

    public void setTerminationDate(LocalDate terminationDate) {
        this.terminationDate = terminationDate;
    }

    public List<SpecialDeductionType> getSpecialDeductions() {
        return specialDeductions;
    }

    public void setSpecialDeductions(List<SpecialDeductionType> specialDeductions) {
        this.specialDeductions = specialDeductions;
    }

    public void addSpecialDeduction(SpecialDeductionType type) {
        if (!canAddSpecialDeduction(type)) {
            throw new IllegalArgumentException("无法添加该专项附加扣除，存在互斥项");
        }
        this.specialDeductions.add(type);
    }

    public boolean canAddSpecialDeduction(SpecialDeductionType type) {
        if (type == SpecialDeductionType.HOUSING_LOAN_INTEREST) {
            return !specialDeductions.contains(SpecialDeductionType.HOUSING_RENT);
        }
        if (type == SpecialDeductionType.HOUSING_RENT) {
            return !specialDeductions.contains(SpecialDeductionType.HOUSING_LOAN_INTEREST);
        }
        return true;
    }

    public List<SalaryAdjustment> getSalaryAdjustments() {
        return salaryAdjustments;
    }

    public void setSalaryAdjustments(List<SalaryAdjustment> salaryAdjustments) {
        this.salaryAdjustments = salaryAdjustments;
    }

    public void addSalaryAdjustment(SalaryAdjustment adjustment) {
        this.salaryAdjustments.add(adjustment);
    }
}
