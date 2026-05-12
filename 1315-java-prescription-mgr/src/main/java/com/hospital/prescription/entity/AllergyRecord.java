package com.hospital.prescription.entity;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.hospital.prescription.enums.AllergyType;
import jakarta.persistence.*;
import lombok.Data;

import java.time.LocalDateTime;

@Entity
@Data
@Table(name = "allergy_records")
public class AllergyRecord {
    
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    
    @JsonIgnore
    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "patient_id", nullable = false)
    private Patient patient;
    
    @Column(nullable = false)
    private String allergen;
    
    @Column(nullable = false)
    @Enumerated(EnumType.STRING)
    private AllergyType allergyType;
    
    private String reaction;
    
    private LocalDateTime recordDate;
    
    private LocalDateTime createdAt;
    
    @PrePersist
    protected void onCreate() {
        if (recordDate == null) {
            recordDate = LocalDateTime.now();
        }
        createdAt = LocalDateTime.now();
    }
}
