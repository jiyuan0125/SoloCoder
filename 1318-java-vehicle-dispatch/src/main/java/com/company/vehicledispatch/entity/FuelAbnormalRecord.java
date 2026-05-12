package com.company.vehicledispatch.entity;

import com.fasterxml.jackson.annotation.JsonIgnore;
import lombok.Data;

import javax.persistence.*;
import java.time.LocalDateTime;

@Data
@Entity
@Table(name = "fuel_abnormal_records")
public class FuelAbnormalRecord {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @JsonIgnore
    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "dispatch_record_id", nullable = false)
    private DispatchRecord dispatchRecord;

    @JsonIgnore
    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "vehicle_id", nullable = false)
    private Vehicle vehicle;

    @Column(nullable = false)
    private Double estimatedFuelConsumption;

    @Column(nullable = false)
    private Double actualFuelConsumption;

    @Column(nullable = false)
    private Double deviationPercentage;

    private String remarks;

    private LocalDateTime createdAt;
}
