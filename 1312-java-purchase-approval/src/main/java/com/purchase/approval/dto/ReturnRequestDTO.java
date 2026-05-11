package com.purchase.approval.dto;

import lombok.Data;
import javax.validation.constraints.NotBlank;
import javax.validation.constraints.NotNull;
import javax.validation.constraints.Positive;

@Data
public class ReturnRequestDTO {
    @NotNull(message = "采购订单ID不能为空")
    private Long purchaseOrderId;
    
    @NotNull(message = "退货数量不能为空")
    @Positive(message = "退货数量必须大于0")
    private Integer returnQuantity;
    
    @NotBlank(message = "退货原因不能为空")
    private String returnReason;
    
    @NotNull(message = "提交人ID不能为空")
    private Long submitterId;
    
    @NotBlank(message = "提交人姓名不能为空")
    private String submitterName;
}
