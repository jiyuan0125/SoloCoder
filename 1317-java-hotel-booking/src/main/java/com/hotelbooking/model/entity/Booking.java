package com.hotelbooking.model.entity;

import com.hotelbooking.model.enums.BookingStatus;
import com.hotelbooking.model.enums.RoomType;
import jakarta.persistence.*;
import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.AllArgsConstructor;
import lombok.Builder;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.time.LocalDateTime;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Builder
@Entity
@Table(name = "bookings")
public class Booking {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(name = "booking_number", unique = true, nullable = false)
    private String bookingNumber;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "customer_id", nullable = false)
    private Customer customer;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "room_id")
    private Room room;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    private RoomType roomType;

    @Column(name = "check_in_date", nullable = false)
    private LocalDate checkInDate;

    @Column(name = "check_out_date", nullable = false)
    private LocalDate checkOutDate;

    @Column(name = "number_of_guests", nullable = false)
    private Integer numberOfGuests = 1;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    private BookingStatus status = BookingStatus.PENDING_CONFIRMATION;

    @Column(name = "original_check_out_date")
    private LocalDate originalCheckOutDate;

    @Column(name = "base_total", nullable = false, precision = 12, scale = 2)
    private BigDecimal baseTotal;

    @Column(name = "continuous_stay_discount", precision = 5, scale = 2)
    private BigDecimal continuousStayDiscount = BigDecimal.ZERO;

    @Column(name = "member_discount", precision = 5, scale = 2)
    private BigDecimal memberDiscount = BigDecimal.ZERO;

    @Column(name = "holiday_surcharge_total", precision = 12, scale = 2)
    private BigDecimal holidaySurchargeTotal = BigDecimal.ZERO;

    @Column(name = "total_amount", nullable = false, precision = 12, scale = 2)
    private BigDecimal totalAmount;

    @Column(name = "paid_amount", precision = 12, scale = 2)
    private BigDecimal paidAmount = BigDecimal.ZERO;

    @Column(name = "cancellation_fee", precision = 12, scale = 2)
    private BigDecimal cancellationFee = BigDecimal.ZERO;

    @Column(name = "early_check_out_fee", precision = 12, scale = 2)
    private BigDecimal earlyCheckOutFee = BigDecimal.ZERO;

    @Column(name = "special_requests")
    private String specialRequests;

    @Column(name = "upgraded_from_room_type")
    @Enumerated(EnumType.STRING)
    private RoomType upgradedFromRoomType;

    @Column(name = "is_hotel_caused_upgrade")
    private Boolean isHotelCausedUpgrade = false;

    @Column(name = "check_in_time")
    private LocalDateTime checkInTime;

    @Column(name = "check_out_time")
    private LocalDateTime checkOutTime;

    @Column(name = "cancelled_at")
    private LocalDateTime cancelledAt;

    @Column(name = "created_at", nullable = false, updatable = false)
    private LocalDateTime createdAt;

    @Column(name = "updated_at")
    private LocalDateTime updatedAt;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "created_by")
    private User createdBy;

    @PrePersist
    protected void onCreate() {
        createdAt = LocalDateTime.now();
        updatedAt = LocalDateTime.now();
        if (bookingNumber == null) {
            bookingNumber = generateBookingNumber();
        }
    }

    @PreUpdate
    protected void onUpdate() {
        updatedAt = LocalDateTime.now();
    }

    private String generateBookingNumber() {
        return "BK" + System.currentTimeMillis();
    }
}
