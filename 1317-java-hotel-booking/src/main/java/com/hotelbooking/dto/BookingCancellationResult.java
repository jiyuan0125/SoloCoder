package com.hotelbooking.dto;

import lombok.Builder;
import lombok.Data;

import java.math.BigDecimal;

@Data
@Builder
public class BookingCancellationResult {
    private boolean success;
    private String message;
    private BigDecimal cancellationFee;
    private BigDecimal refundAmount;
}
