package com.hospital.prescription.entity;

import com.hospital.prescription.enums.DosageForm;
import com.hospital.prescription.enums.SolventType;
import com.fasterxml.jackson.annotation.JsonIgnore;
import jakarta.persistence.*;
import lombok.Data;

import java.math.BigDecimal;

@Entity
@Data
@Table(name = "prescription_items")
public class PrescriptionItem {
    
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    
    @JsonIgnore
    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "prescription_id", nullable = false)
    private Prescription prescription;
    
    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "drug_id", nullable = false)
    private Drug drug;
    
    @Column(nullable = false)
    private String drugName;
    
    @Column(nullable = false)
    @Enumerated(EnumType.STRING)
    private DosageForm dosageForm;
    
    @Column(nullable = false, precision = 10, scale = 4)
    private BigDecimal singleDose;
    
    @Column(nullable = false)
    private String doseUnit;
    
    @Column(nullable = false)
    private Integer frequencyPerDay;
    
    @Column(nullable = false)
    private Integer durationDays;
    
    @Enumerated(EnumType.STRING)
    private SolventType solventType;
    
    private String administrationNotes;
}
