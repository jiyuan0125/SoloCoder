package com.coupon.dto;

import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Positive;
import lombok.Data;

import java.math.BigDecimal;
import java.util.List;

@Data
public class UseCouponRequest {
    @NotNull
    private Long userId;

    private List<Long> userCouponIds;

    @NotNull
    @Positive
    private BigDecimal orderAmount;

    @NotNull
    @Positive
    private BigDecimal shippingFee;

    @NotNull
    private String orderId;
}
