package com.hospital.prescription.entity;

import jakarta.persistence.*;
import lombok.Data;

import java.time.LocalDateTime;

@Entity
@Data
@Table(name = "prescription_signatures")
public class PrescriptionSignature {
    
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    
    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "prescription_id", nullable = false)
    private Prescription prescription;
    
    @Column(nullable = false)
    private String signerName;
    
    @Column(nullable = false)
    private String signerRole;
    
    private LocalDateTime signedAt;
    
    @PrePersist
    protected void onCreate() {
        signedAt = LocalDateTime.now();
    }
}
