package com.hospital.prescription.entity;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.hospital.prescription.enums.SeverityLevel;
import jakarta.persistence.*;
import lombok.Data;

import java.time.LocalDateTime;

@Entity
@Data
@Table(name = "prescription_validations")
public class PrescriptionValidation {
    
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    
    @JsonIgnore
    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "prescription_id", nullable = false)
    private Prescription prescription;
    
    @Column(nullable = false)
    private String validationType;
    
    @Column(nullable = false)
    private String message;
    
    @Column(nullable = false)
    @Enumerated(EnumType.STRING)
    private SeverityLevel severity;
    
    private String relatedDrugNames;
    
    private String suggestion;
    
    private LocalDateTime createdAt;
    
    @PrePersist
    protected void onCreate() {
        createdAt = LocalDateTime.now();
    }
}
