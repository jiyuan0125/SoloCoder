package com.reimbursement.common.dto;

import java.math.BigDecimal;
import java.util.List;

public class CreateReimbursementRequest {
    private String employeeId;
    private String employeeName;
    private String reimbursementType;
    private String description;
    private BigDecimal totalAmount;
    private List<AllocationItem> allocations;
    private boolean specialApprovalRequested;

    public CreateReimbursementRequest() {
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

    public String getReimbursementType() {
        return reimbursementType;
    }

    public void setReimbursementType(String reimbursementType) {
        this.reimbursementType = reimbursementType;
    }

    public String getDescription() {
        return description;
    }

    public void setDescription(String description) {
        this.description = description;
    }

    public BigDecimal getTotalAmount() {
        return totalAmount;
    }

    public void setTotalAmount(BigDecimal totalAmount) {
        this.totalAmount = totalAmount;
    }

    public List<AllocationItem> getAllocations() {
        return allocations;
    }

    public void setAllocations(List<AllocationItem> allocations) {
        this.allocations = allocations;
    }

    public boolean isSpecialApprovalRequested() {
        return specialApprovalRequested;
    }

    public void setSpecialApprovalRequested(boolean specialApprovalRequested) {
        this.specialApprovalRequested = specialApprovalRequested;
    }

    public static class AllocationItem {
        private String costCenterId;
        private int percentage;

        public AllocationItem() {
        }

        public String getCostCenterId() {
            return costCenterId;
        }

        public void setCostCenterId(String costCenterId) {
            this.costCenterId = costCenterId;
        }

        public int getPercentage() {
            return percentage;
        }

        public void setPercentage(int percentage) {
            this.percentage = percentage;
        }
    }
}
