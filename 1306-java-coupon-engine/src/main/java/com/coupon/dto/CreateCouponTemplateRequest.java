package com.coupon.dto;

import com.coupon.enums.CouponType;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Positive;
import lombok.Data;

import java.math.BigDecimal;
import java.time.LocalDateTime;

@Data
public class CreateCouponTemplateRequest {
    @NotNull
    private String name;

    @NotNull
    private CouponType type;

    @NotNull
    @Positive
    private BigDecimal value;

    @NotNull
    private BigDecimal threshold;

    @NotNull
    private LocalDateTime startTime;

    @NotNull
    private LocalDateTime endTime;

    @NotNull
    @Positive
    private Integer totalQuantity;

    @NotNull
    @Positive
    private Integer limitPerUser;
}
