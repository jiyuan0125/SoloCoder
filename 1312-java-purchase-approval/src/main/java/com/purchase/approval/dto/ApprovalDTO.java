package com.purchase.approval.dto;

import lombok.Data;
import javax.validation.constraints.NotBlank;
import javax.validation.constraints.NotNull;

@Data
public class ApprovalDTO {
    @NotNull(message = "审批人ID不能为空")
    private Long approverId;
    
    @NotBlank(message = "审批人姓名不能为空")
    private String approverName;
    
    @NotNull(message = "是否同意不能为空")
    private Boolean approved;
    
    private String remarks;
    
    private String rejectionReason;
}
