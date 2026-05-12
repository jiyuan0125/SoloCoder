package com.hotelbooking.dto;

import com.hotelbooking.model.enums.BookingStatus;
import com.hotelbooking.model.enums.RoomType;
import lombok.Builder;
import lombok.Data;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.time.LocalDateTime;

@Data
@Builder
public class BookingResponse {
    private Long id;
    private String bookingNumber;
    private Long customerId;
    private String customerName;
    private String customerPhone;
    private Long roomId;
    private String roomNumber;
    private RoomType roomType;
    private String roomTypeName;
    private LocalDate checkInDate;
    private LocalDate checkOutDate;
    private Integer numberOfGuests;
    private BookingStatus status;
    private String statusName;
    private LocalDate originalCheckOutDate;
    private BigDecimal baseTotal;
    private BigDecimal continuousStayDiscount;
    private BigDecimal memberDiscount;
    private BigDecimal holidaySurchargeTotal;
    private BigDecimal totalAmount;
    private BigDecimal paidAmount;
    private BigDecimal cancellationFee;
    private BigDecimal earlyCheckOutFee;
    private String specialRequests;
    private RoomType upgradedFromRoomType;
    private Boolean isHotelCausedUpgrade;
    private LocalDateTime checkInTime;
    private LocalDateTime checkOutTime;
    private LocalDateTime cancelledAt;
    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;
}
