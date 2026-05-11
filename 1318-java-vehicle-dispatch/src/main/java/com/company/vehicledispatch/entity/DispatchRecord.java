package com.company.vehicledispatch.entity;

import lombok.Data;

import javax.persistence.*;
import java.time.LocalDateTime;

@Data
@Entity
@Table(name = "dispatch_records")
public class DispatchRecord {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @OneToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "request_id", nullable = false, unique = true)
    private DispatchRequest request;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "vehicle_id", nullable = false)
    private Vehicle vehicle;

    @Column(nullable = false)
    private LocalDateTime actualStartDateTime;

    private LocalDateTime actualEndDateTime;

    private Double startMileage;

    private Double endMileage;

    private Double actualDistance;

    private Double estimatedFuelConsumption;

    private Double actualFuelConsumption;

    private Boolean fuelAbnormal;

    private LocalDateTime createdAt;

    private LocalDateTime updatedAt;
}
