package com.reimbursement.common.enums;

public enum ReimbursementStatus {
    DRAFT("草稿"),
    SUBMITTED("已提交"),
    PENDING_APPROVAL("待审批"),
    APPROVING("审批中"),
    REJECTED("已驳回"),
    ALL_APPROVED("全部审批通过"),
    PENDING_FINAL_REVIEW("待财务复核"),
    FINAL_REVIEWED("财务复核通过"),
    PAID("已打款");

    private final String description;

    ReimbursementStatus(String description) {
        this.description = description;
    }

    public String getDescription() {
        return description;
    }
}
