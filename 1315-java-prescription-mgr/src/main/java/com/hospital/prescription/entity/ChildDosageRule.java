package com.hospital.prescription.entity;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.hospital.prescription.enums.ChildAgeGroup;
import jakarta.persistence.*;
import lombok.Data;

import java.math.BigDecimal;
import java.time.LocalDateTime;

@Entity
@Data
@Table(name = "child_dosage_rules")
public class ChildDosageRule {
    
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    
    @JsonIgnore
    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "drug_id", nullable = false)
    private Drug drug;
    
    @Column(nullable = false)
    @Enumerated(EnumType.STRING)
    private ChildAgeGroup ageGroup;
    
    @Column(nullable = false, precision = 5, scale = 2)
    private BigDecimal doseRatio;
    
    @Column(precision = 10, scale = 4)
    private BigDecimal maxSingleDose;
    
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
}
