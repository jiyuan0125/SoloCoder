package com.payroll.common.dto;

import java.math.BigDecimal;
import java.util.List;

public class ArchiveRequestDTO {

    private String employeeId;
    private String employeeName;
    private String idCard;
    private int year;
    private int month;
    private BigDecimal baseSalary;
    private List<DeductionDetailDTO> deductionDetails;

    public ArchiveRequestDTO() {
    }

    public String getEmployeeId() {
        return employeeId;
    }

    public void setEmployeeId(String employeeId) {
        this.employeeId = employeeId;
    }

    public String getEmployeeName() {
        return employeeName;
    }

    public void setEmployeeName(String employeeName) {
        this.employeeName = employeeName;
    }

    public String getIdCard() {
        return idCard;
    }

    public void setIdCard(String idCard) {
        this.idCard = idCard;
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

    public List<DeductionDetailDTO> getDeductionDetails() {
        return deductionDetails;
    }

    public void setDeductionDetails(List<DeductionDetailDTO> deductionDetails) {
        this.deductionDetails = deductionDetails;
    }
}
