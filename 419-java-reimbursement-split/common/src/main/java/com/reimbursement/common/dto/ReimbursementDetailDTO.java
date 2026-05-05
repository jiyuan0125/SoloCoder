package com.reimbursement.common.dto;

import com.reimbursement.common.enums.ReimbursementStatus;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.List;

public class ReimbursementDetailDTO {
    private String reimbursementId;
    private String employeeId;
    private String employeeName;
    private String reimbursementType;
    private String description;
    private BigDecimal totalAmount;
    private ReimbursementStatus status;
    private LocalDateTime createTime;
    private LocalDateTime updateTime;
    private List<CostCenterApprovalDTO> allocations;
    private String rejectReason;

    public ReimbursementDetailDTO() {
    }

    public String getReimbursementId() {
        return reimbursementId;
    }

    public void setReimbursementId(String reimbursementId) {
        this.reimbursementId = reimbursementId;
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

    public ReimbursementStatus getStatus() {
        return status;
    }

    public void setStatus(ReimbursementStatus status) {
        this.status = status;
    }

    public LocalDateTime getCreateTime() {
        return createTime;
    }

    public void setCreateTime(LocalDateTime createTime) {
        this.createTime = createTime;
    }

    public LocalDateTime getUpdateTime() {
        return updateTime;
    }

    public void setUpdateTime(LocalDateTime updateTime) {
        this.updateTime = updateTime;
    }

    public List<CostCenterApprovalDTO> getAllocations() {
        return allocations;
    }

    public void setAllocations(List<CostCenterApprovalDTO> allocations) {
        this.allocations = allocations;
    }

    public String getRejectReason() {
        return rejectReason;
    }

    public void setRejectReason(String rejectReason) {
        this.rejectReason = rejectReason;
    }
}
