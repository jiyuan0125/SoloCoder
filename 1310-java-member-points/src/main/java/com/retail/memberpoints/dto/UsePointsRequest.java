package com.retail.memberpoints.dto;

import jakarta.validation.constraints.DecimalMin;
import jakarta.validation.constraints.Min;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class UsePointsRequest {

    @NotNull(message = "会员ID不能为空")
    private Long memberId;

    @NotBlank(message = "订单号不能为空")
    private String orderNo;

    @NotNull(message = "使用积分不能为空")
    @Min(value = 1, message = "使用积分必须大于0")
    private Integer points;

    @NotNull(message = "订单金额不能为空")
    @DecimalMin(value = "0.01", message = "订单金额必须大于0")
    private BigDecimal orderAmount;
}
