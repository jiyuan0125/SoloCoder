package com.example.groupbuy.dto;

import lombok.Data;

import javax.validation.constraints.NotNull;

@Data
public class JoinGroupBuyDTO {
    
    @NotNull(message = "拼团订单ID不能为空")
    private Long groupOrderId;
    
    @NotNull(message = "用户ID不能为空")
    private Long userId;
}
