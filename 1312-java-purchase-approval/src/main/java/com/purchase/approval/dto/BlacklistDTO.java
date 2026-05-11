package com.purchase.approval.dto;

import lombok.Data;
import javax.validation.constraints.NotBlank;
import javax.validation.constraints.NotNull;
import java.time.LocalDate;

@Data
public class BlacklistDTO {
    @NotNull(message = "供应商ID不能为空")
    private Long supplierId;
    
    @NotBlank(message = "拉黑原因不能为空")
    private String blacklistReason;
    
    @NotNull(message = "截止日期不能为空")
    private LocalDate blacklistEndDate;
    
    private String remarks;
}
