package com.training.common.dto.response;

import java.util.List;

public class CreditSummaryDTO {
    private String employeeId;
    private String employeeName;
    private String departmentId;
    private String departmentName;
    private int year;
    private int totalRequiredCredits;
    private int earnedRequiredCredits;
    private int totalElectiveCredits;
    private int earnedElectiveCredits;
    private int totalCredits;
    private boolean eligibleForExcellentCertificate;
    private List<CreditDetailDTO> creditDetails;

    public CreditSummaryDTO() {
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

    public int getTotalRequiredCredits() {
        return totalRequiredCredits;
    }

    public void setTotalRequiredCredits(int totalRequiredCredits) {
        this.totalRequiredCredits = totalRequiredCredits;
    }

    public int getEarnedRequiredCredits() {
        return earnedRequiredCredits;
    }

    public void setEarnedRequiredCredits(int earnedRequiredCredits) {
        this.earnedRequiredCredits = earnedRequiredCredits;
    }

    public int getTotalElectiveCredits() {
        return totalElectiveCredits;
    }

    public void setTotalElectiveCredits(int totalElectiveCredits) {
        this.totalElectiveCredits = totalElectiveCredits;
    }

    public int getEarnedElectiveCredits() {
        return earnedElectiveCredits;
    }

    public void setEarnedElectiveCredits(int earnedElectiveCredits) {
        this.earnedElectiveCredits = earnedElectiveCredits;
    }

    public int getTotalCredits() {
        return totalCredits;
    }

    public void setTotalCredits(int totalCredits) {
        this.totalCredits = totalCredits;
    }

    public boolean isEligibleForExcellentCertificate() {
        return eligibleForExcellentCertificate;
    }

    public void setEligibleForExcellentCertificate(boolean eligibleForExcellentCertificate) {
        this.eligibleForExcellentCertificate = eligibleForExcellentCertificate;
    }

    public List<CreditDetailDTO> getCreditDetails() {
        return creditDetails;
    }

    public void setCreditDetails(List<CreditDetailDTO> creditDetails) {
        this.creditDetails = creditDetails;
    }
}
