package com.company.vehicledispatch.entity;

import com.fasterxml.jackson.annotation.JsonIgnore;
import lombok.Data;

import javax.persistence.*;
import java.time.LocalDate;
import java.time.LocalDateTime;

@Data
@Entity
@Table(name = "monthly_reports")
public class MonthlyReport {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @JsonIgnore
    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "vehicle_id", nullable = false)
    private Vehicle vehicle;

    @Column(nullable = false)
    private Integer reportYear;

    @Column(nullable = false)
    private Integer reportMonth;

    private Integer usageDays;

    private Double totalMileage;

    private Double totalFuelConsumption;

    private Double maintenanceCost;

    private Double usageRate;

    private String recommendation;

    private LocalDateTime createdAt;
}
