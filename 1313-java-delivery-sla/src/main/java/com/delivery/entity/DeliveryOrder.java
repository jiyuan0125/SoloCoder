package com.delivery.entity;

import com.delivery.enums.DeliveryType;
import com.delivery.enums.OrderStatus;
import com.delivery.enums.TimeSlot;
import jakarta.persistence.*;
import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.AllArgsConstructor;
import lombok.Builder;
import java.math.BigDecimal;
import java.time.LocalDateTime;

@Entity
@Table(name = "orders")
@Data
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class DeliveryOrder {
    
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    
    @Column(nullable = false, unique = true)
    private String orderNo;
    
    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "customer_id", nullable = false)
    private Customer customer;
    
    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    private DeliveryType deliveryType;
    
    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    private OrderStatus status;
    
    @Enumerated(EnumType.STRING)
    private TimeSlot preferredTimeSlot;
    
    @Column(nullable = false)
    private String senderName;
    
    @Column(nullable = false)
    private String senderPhone;
    
    @Column(nullable = false, columnDefinition = "TEXT")
    private String senderAddress;
    
    private Double senderLatitude;
    
    private Double senderLongitude;
    
    @Column(nullable = false)
    private String receiverName;
    
    @Column(nullable = false)
    private String receiverPhone;
    
    @Column(nullable = false, columnDefinition = "TEXT")
    private String receiverAddress;
    
    private Double receiverLatitude;
    
    private Double receiverLongitude;
    
    @Column(nullable = false, precision = 10, scale = 2)
    private BigDecimal weight;
    
    private Double length;
    
    private Double width;
    
    private Double height;
    
    @Column(nullable = false, precision = 10, scale = 2)
    private BigDecimal distance;
    
    @Column(nullable = false, precision = 10, scale = 2)
    private BigDecimal baseFee;
    
    @Column(nullable = false, precision = 10, scale = 2)
    private BigDecimal deliveryTypeSurcharge;
    
    @Column(nullable = false, precision = 10, scale = 2)
    private BigDecimal weightSurcharge;
    
    @Column(nullable = false, precision = 10, scale = 2)
    private BigDecimal oversizedFee;
    
    @Column(nullable = false, precision = 10, scale = 2)
    private BigDecimal discount;
    
    @Column(nullable = false, precision = 10, scale = 2)
    private BigDecimal totalFee;
    
    private Integer monthlyOrderCount;
    
    @Column(nullable = false)
    private LocalDateTime orderTime;
    
    private LocalDateTime effectiveOrderTime;
    
    private LocalDateTime promisedDeliveryTime;
    
    private LocalDateTime actualDeliveryTime;
    
    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "delivery_person_id")
    private DeliveryPerson deliveryPerson;
    
    private String assignedZoneCode;
    
    private String cancellationReason;
    
    private LocalDateTime cancelledAt;
    
    @Column(nullable = false, updatable = false)
    private LocalDateTime createdAt;
    
    @Column(nullable = false)
    private LocalDateTime updatedAt;
    
    @PrePersist
    protected void onCreate() {
        createdAt = LocalDateTime.now();
        updatedAt = LocalDateTime.now();
        if (status == null) {
            status = OrderStatus.PENDING;
        }
    }
    
    @PreUpdate
    protected void onUpdate() {
        updatedAt = LocalDateTime.now();
    }
}
