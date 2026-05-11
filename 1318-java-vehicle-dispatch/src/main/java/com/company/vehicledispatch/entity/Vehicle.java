package com.company.vehicledispatch.entity;

import com.company.vehicledispatch.enums.FuelType;
import com.company.vehicledispatch.enums.VehicleStatus;
import lombok.Data;

import javax.persistence.*;
import java.time.LocalDate;

@Data
@Entity
@Table(name = "vehicles")
public class Vehicle {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(nullable = false, unique = true)
    private String plateNumber;

    @Column(nullable = false)
    private String model;

    @Column(nullable = false)
    private Integer seats;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    private FuelType fuelType;

    @Column(nullable = false)
    private Double fuelLevel;

    @Column(nullable = false)
    private Double maxFuelLevel;

    @Column(nullable = false)
    private Double currentMileage;

    @Column(nullable = false)
    private Double lastMaintenanceMileage;

    @Column(nullable = false)
    private LocalDate insuranceExpiryDate;

    @Column(nullable = false)
    private LocalDate annualInspectionExpiryDate;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    private VehicleStatus status;

    private Double averageFuelConsumption;

    private Boolean fuelWarning;

    private Boolean maintenanceDue;

    private Boolean insuranceDue;

    private Boolean annualInspectionDue;
}
