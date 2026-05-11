package com.inventory.entity;

import jakarta.persistence.*;
import lombok.Data;
import java.time.LocalDate;
import java.time.LocalDateTime;

@Data
@Entity
@Table(name = "batches", uniqueConstraints = {
    @UniqueConstraint(columnNames = {"product_id", "batch_number"})
})
public class Batch {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "product_id", nullable = false)
    private Product product;

    @Column(name = "batch_number", nullable = false)
    private String batchNumber;

    @Column(name = "inbound_date", nullable = false)
    private LocalDate inboundDate;

    @Column(name = "production_date", nullable = false)
    private LocalDate productionDate;

    @Column(name = "shelf_life_days", nullable = false)
    private Integer shelfLifeDays;

    @Column(name = "initial_quantity", nullable = false)
    private Integer initialQuantity;

    @Column(name = "created_at", nullable = false, updatable = false)
    private LocalDateTime createdAt;

    @Transient
    public LocalDate getExpiryDate() {
        return productionDate.plusDays(shelfLifeDays);
    }

    @PrePersist
    protected void onCreate() {
        createdAt = LocalDateTime.now();
    }
}
