package com.reimbursement.common.dto;

import java.math.BigDecimal;
import java.util.List;

public class MonthlyReportDTO {
    private int year;
    private int month;
    private String costCenterId;
    private String costCenterName;
    private BigDecimal totalAmount;
    private BigDecimal budgetAmount;
    private BigDecimal remainingBudget;
    private boolean isOverBudget;
    private List<ReimbursementSummaryDTO> reimbursements;

    public MonthlyReportDTO() {
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

    public String getCostCenterId() {
        return costCenterId;
    }

    public void setCostCenterId(String costCenterId) {
        this.costCenterId = costCenterId;
    }

    public String getCostCenterName() {
        return costCenterName;
    }

    public void setCostCenterName(String costCenterName) {
        this.costCenterName = costCenterName;
    }

    public BigDecimal getTotalAmount() {
        return totalAmount;
    }

    public void setTotalAmount(BigDecimal totalAmount) {
        this.totalAmount = totalAmount;
    }

    public BigDecimal getBudgetAmount() {
        return budgetAmount;
    }

    public void setBudgetAmount(BigDecimal budgetAmount) {
        this.budgetAmount = budgetAmount;
    }

    public BigDecimal getRemainingBudget() {
        return remainingBudget;
    }

    public void setRemainingBudget(BigDecimal remainingBudget) {
        this.remainingBudget = remainingBudget;
    }

    public boolean isOverBudget() {
        return isOverBudget;
    }

    public void setOverBudget(boolean overBudget) {
        isOverBudget = overBudget;
    }

    public List<ReimbursementSummaryDTO> getReimbursements() {
        return reimbursements;
    }

    public void setReimbursements(List<ReimbursementSummaryDTO> reimbursements) {
        this.reimbursements = reimbursements;
    }

    public static class ReimbursementSummaryDTO {
        private String reimbursementId;
        private String employeeName;
        private String reimbursementType;
        private BigDecimal amount;

        public ReimbursementSummaryDTO() {
        }

        public String getReimbursementId() {
            return reimbursementId;
        }

        public void setReimbursementId(String reimbursementId) {
            this.reimbursementId = reimbursementId;
        }

        public String getEmployeeName() {
            return employeeName;
        }

        public void setEmployeeName(String employeeName) {
            this.employeeName = employeeName;
        }

        public String getReimbursementType() {
            return reimbursementType;
        }

        public void setReimbursementType(String reimbursementType) {
            this.reimbursementType = reimbursementType;
        }

        public BigDecimal getAmount() {
            return amount;
        }

        public void setAmount(BigDecimal amount) {
            this.amount = amount;
        }
    }
}
