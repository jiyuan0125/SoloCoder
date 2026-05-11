package com.delivery.entity;

import com.delivery.enums.CompensationType;
import jakarta.persistence.*;
import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.AllArgsConstructor;
import lombok.Builder;
import java.math.BigDecimal;
import java.time.LocalDateTime;

@Entity
@Table(name = "compensations")
@Data
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class Compensation {
    
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    
    @OneToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "order_id", nullable = false, unique = true)
    private DeliveryOrder order;
    
    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    private CompensationType compensationType;
    
    @Column(nullable = false, precision = 10, scale = 2)
    private BigDecimal compensationAmount;
    
    @Column(nullable = false, precision = 10, scale = 2)
    private BigDecimal originalFee;
    
    @Column(nullable = false)
    private Long delayMinutes;
    
    @Column(columnDefinition = "TEXT")
    private String description;
    
    private boolean approved;
    
    private String approvedBy;
    
    private LocalDateTime approvedAt;
    
    @Column(nullable = false, updatable = false)
    private LocalDateTime createdAt;
    
    @Column(nullable = false)
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
