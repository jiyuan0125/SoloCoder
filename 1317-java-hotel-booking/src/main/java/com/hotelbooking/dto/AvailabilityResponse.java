package com.hotelbooking.dto;

import com.hotelbooking.model.enums.RoomType;
import lombok.Builder;
import lombok.Data;

import java.math.BigDecimal;

@Data
@Builder
public class AvailabilityResponse {
    private RoomType roomType;
    private String roomTypeName;
    private Integer totalRooms;
    private Integer availableRooms;
    private BigDecimal basePrice;
    private BigDecimal estimatedPricePerNight;
    private boolean available;
}
