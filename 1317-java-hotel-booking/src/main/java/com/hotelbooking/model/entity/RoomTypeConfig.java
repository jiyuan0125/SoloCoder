package com.hotelbooking.model.entity;

import com.hotelbooking.model.enums.RoomType;
import jakarta.persistence.*;
import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.AllArgsConstructor;
import lombok.Builder;

import java.math.BigDecimal;
import java.time.LocalDateTime;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Builder
@Entity
@Table(name = "room_type_configs")
public class RoomTypeConfig {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Enumerated(EnumType.STRING)
    @Column(unique = true, nullable = false)
    private RoomType roomType;

    @Column(name = "room_count", nullable = false)
    private Integer roomCount;

    @Column(name = "base_price", nullable = false, precision = 12, scale = 2)
    private BigDecimal basePrice;

    @Column(name = "holiday_eve_surcharge", precision = 5, scale = 2)
    private BigDecimal holidayEveSurcharge = new BigDecimal("0.10");

    @Column(name = "holiday_surcharge", precision = 5, scale = 2)
    private BigDecimal holidaySurcharge = new BigDecimal("0.30");

    @Column(name = "holiday_after_surcharge", precision = 5, scale = 2)
    private BigDecimal holidayAfterSurcharge = new BigDecimal("0.05");

    @Column(name = "max_guests")
    private Integer maxGuests;

    private String description;

    @Column(name = "created_at", nullable = false, updatable = false)
    private LocalDateTime createdAt;

    @Column(name = "updated_at")
    private LocalDateTime updatedAt;

    @PrePersist
    protected void onCreate() {
        createdAt = LocalDateTime.now();
        updatedAt = LocalDateTime.now();
    }

    @PreUpdate
    protected void onUpdate() {
        updatedAt = LocalDateTime.now();
    }
}
