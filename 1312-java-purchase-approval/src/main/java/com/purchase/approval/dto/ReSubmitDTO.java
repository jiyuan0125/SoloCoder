package com.purchase.approval.dto;

import lombok.Data;
import javax.validation.constraints.NotBlank;
import javax.validation.constraints.NotNull;

@Data
public class ReSubmitDTO {
    @NotBlank(message = "提交人姓名不能为空")
    private String submitterName;
    
    @NotNull(message = "提交人ID不能为空")
    private Long submitterId;
}
