package com.company.expense.dto;

import com.company.expense.entity.ApprovalRecord;
import lombok.Builder;
import lombok.Data;

import java.time.LocalDateTime;

@Data
@Builder
public class ApprovalRecordResponse {

    private Long id;
    private String approverEmployeeId;
    private String approverName;
    private String approvalLevel;
    private String approvalLevelDescription;
    private String action;
    private String actionDescription;
    private String comment;
    private LocalDateTime createdAt;

    public static ApprovalRecordResponse fromEntity(ApprovalRecord record) {
        return ApprovalRecordResponse.builder()
                .id(record.getId())
                .approverEmployeeId(record.getApprover().getEmployeeId())
                .approverName(record.getApprover().getName())
                .approvalLevel(record.getApprovalLevel().name())
                .approvalLevelDescription(record.getApprovalLevel().getDescription())
                .action(record.getAction().name())
                .actionDescription(record.getAction().getDescription())
                .comment(record.getComment())
                .createdAt(record.getCreatedAt())
                .build();
    }
}
