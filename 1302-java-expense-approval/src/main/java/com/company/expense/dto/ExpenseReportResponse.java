package com.company.expense.dto;

import com.company.expense.entity.ExpenseReport;
import lombok.Builder;
import lombok.Data;

import java.math.BigDecimal;
import java.time.LocalDateTime;

@Data
@Builder
public class ExpenseReportResponse {

    private Long id;
    private String reportNo;
    private String applicantEmployeeId;
    private String applicantName;
    private String expenseType;
    private String expenseTypeDescription;
    private BigDecimal amount;
    private String description;
    private String status;
    private String statusDescription;
    private String currentApprovalLevel;
    private String currentApprovalLevelDescription;
    private String currentApproverEmployeeId;
    private String currentApproverName;
    private String rejectReason;
    private LocalDateTime submittedAt;
    private LocalDateTime approvedAt;

    public static ExpenseReportResponse fromEntity(ExpenseReport report) {
        ExpenseReportResponseBuilder builder = ExpenseReportResponse.builder()
                .id(report.getId())
                .reportNo(report.getReportNo())
                .applicantEmployeeId(report.getApplicant().getEmployeeId())
                .applicantName(report.getApplicant().getName())
                .expenseType(report.getExpenseType().name())
                .expenseTypeDescription(report.getExpenseType().getDescription())
                .amount(report.getAmount())
                .description(report.getDescription())
                .status(report.getStatus().name())
                .statusDescription(report.getStatus().getDescription())
                .currentApprovalLevel(report.getCurrentApprovalLevel().name())
                .currentApprovalLevelDescription(report.getCurrentApprovalLevel().getDescription())
                .rejectReason(report.getRejectReason())
                .submittedAt(report.getSubmittedAt())
                .approvedAt(report.getApprovedAt());

        if (report.getCurrentApprover() != null) {
            builder.currentApproverEmployeeId(report.getCurrentApprover().getEmployeeId())
                   .currentApproverName(report.getCurrentApprover().getName());
        }

        return builder.build();
    }
}
