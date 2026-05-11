package com.hotelbooking.dto;

import lombok.Builder;
import lombok.Data;

import java.math.BigDecimal;
import java.util.List;

@Data
@Builder
public class PriceCalculationResult {
    private BigDecimal baseTotal;
    private BigDecimal holidaySurchargeTotal;
    private BigDecimal continuousStayDiscount;
    private BigDecimal memberDiscount;
    private BigDecimal totalAmount;
    private Integer numberOfNights;
    private List<DailyPriceDetail> dailyDetails;
}
