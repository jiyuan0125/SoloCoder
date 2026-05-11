package com.hospital.prescription.entity;

import com.hospital.prescription.enums.AdministrationRoute;
import com.hospital.prescription.enums.DosageForm;
import com.hospital.prescription.enums.DrugCategory;
import com.hospital.prescription.enums.SolventType;
import jakarta.persistence.*;
import lombok.Data;

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.time.LocalDateTime;

@Entity
@Data
@Table(name = "drugs")
public class Drug {
    
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    
    @Column(nullable = false)
    private String name;
    
    @Column(nullable = false)
    @Enumerated(EnumType.STRING)
    private DosageForm dosageForm;
    
    @Enumerated(EnumType.STRING)
    private AdministrationRoute administrationRoute;
    
    @Column(nullable = false)
    @Enumerated(EnumType.STRING)
    private DrugCategory category;
    
    @Column(nullable = false, precision = 10, scale = 4)
    private BigDecimal maxSingleDose;
    
    @Column(nullable = false)
    private String doseUnit;
    
    @Enumerated(EnumType.STRING)
    private SolventType requiredSolvent;
    
    private String genericName;
    
    @Column(precision = 10, scale = 2)
    private BigDecimal inventoryQuantity;
    
    private LocalDateTime createdAt;
    
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
    
    public boolean isDoseExceeded(BigDecimal dose) {
        return dose.setScale(4, RoundingMode.HALF_UP).compareTo(maxSingleDose) > 0;
    }
    
    public boolean isLowInventory() {
        return inventoryQuantity != null && inventoryQuantity.compareTo(BigDecimal.ZERO) <= 0;
    }
}
