package com.hotelbooking.dto;

import com.hotelbooking.model.enums.HolidayType;
import lombok.Builder;
import lombok.Data;

import java.math.BigDecimal;
import java.time.LocalDate;

@Data
@Builder
public class DailyPriceDetail {
    private LocalDate stayDate;
    private HolidayType holidayType;
    private BigDecimal basePrice;
    private BigDecimal surchargeRate;
    private BigDecimal surchargeAmount;
    private BigDecimal dailyPrice;
}
