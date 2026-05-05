package com.reimbursement.common.dto;

import java.math.BigDecimal;
import java.util.List;

public class UpdateReimbursementRequest {
    private String reimbursementId;
    private String description;
    private BigDecimal totalAmount;
    private List<CreateReimbursementRequest.AllocationItem> allocations;

    public UpdateReimbursementRequest() {
    }

    public String getReimbursementId() {
        return reimbursementId;
    }

    public void setReimbursementId(String reimbursementId) {
        this.reimbursementId = reimbursementId;
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

    public List<CreateReimbursementRequest.AllocationItem> getAllocations() {
        return allocations;
    }

    public void setAllocations(List<CreateReimbursementRequest.AllocationItem> allocations) {
        this.allocations = allocations;
    }
}
