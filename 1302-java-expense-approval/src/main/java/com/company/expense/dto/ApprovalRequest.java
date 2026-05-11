package com.company.expense.dto;

import lombok.Data;

import javax.validation.constraints.NotBlank;
import javax.validation.constraints.NotNull;

@Data
public class ApprovalRequest {

    @NotBlank(message = "审批人工号不能为空")
    private String approverEmployeeId;

    @NotBlank(message = "审批动作不能为空，只能是 APPROVE 或 REJECT")
    private String action;

    private String comment;
}
