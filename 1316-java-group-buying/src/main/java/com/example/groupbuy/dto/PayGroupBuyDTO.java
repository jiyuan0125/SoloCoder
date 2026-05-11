package com.example.groupbuy.dto;

import lombok.Data;

import javax.validation.constraints.NotNull;
import java.math.BigDecimal;

@Data
public class PayGroupBuyDTO {
    
    @NotNull(message = "拼团订单ID不能为空")
    private Long groupOrderId;
    
    @NotNull(message = "用户ID不能为空")
    private Long userId;
    
    @NotNull(message = "支付金额不能为空")
    private BigDecimal payAmount;
    
    private String payType;
    
    private String payChannel;
}
