package com.purchase.approval.dto;

import lombok.Data;
import javax.validation.constraints.NotNull;
import javax.validation.constraints.Positive;

@Data
public class ReceiptDTO {
    @NotNull(message = "收货数量不能为空")
    @Positive(message = "收货数量必须大于0")
    private Integer receivedQuantity;
    
    private String receivedBy;
    
    private String remarks;
}
