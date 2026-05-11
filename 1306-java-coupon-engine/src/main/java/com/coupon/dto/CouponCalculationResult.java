package com.coupon.dto;

import lombok.Data;

import java.math.BigDecimal;
import java.util.List;

@Data
public class CouponCalculationResult {
    private BigDecimal finalAmount;
    private BigDecimal productDiscount;
    private BigDecimal shippingDiscount;
    private List<Long> appliedCouponIds;
}
