package com.hospital.prescription.entity;

import com.hospital.prescription.enums.SeverityLevel;
import jakarta.persistence.*;
import lombok.Data;

import java.time.LocalDateTime;

@Entity
@Data
@Table(name = "drug_interactions")
public class DrugInteraction {
    
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    
    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "drug_a_id", nullable = false)
    private Drug drugA;
    
    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "drug_b_id", nullable = false)
    private Drug drugB;
    
    @Column(nullable = false)
    @Enumerated(EnumType.STRING)
    private SeverityLevel severity;
    
    @Column(nullable = false, length = 2000)
    private String description;
    
    @Column(length = 2000)
    private String alternative;
    
    private LocalDateTime createdAt;
    
    @PrePersist
    protected void onCreate() {
        createdAt = LocalDateTime.now();
    }
}
