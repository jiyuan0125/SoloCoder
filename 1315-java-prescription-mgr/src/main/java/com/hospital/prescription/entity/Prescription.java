package com.hospital.prescription.entity;

import com.hospital.prescription.enums.PrescriptionStatus;
import jakarta.persistence.*;
import lombok.Data;

import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;

@Entity
@Data
@Table(name = "prescriptions")
public class Prescription {
    
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    
    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "patient_id", nullable = false)
    private Patient patient;
    
    @Column(nullable = false)
    private String diagnosis;
    
    @Column(nullable = false)
    @Enumerated(EnumType.STRING)
    private PrescriptionStatus status;
    
    @Column(nullable = false)
    private String doctorName;
    
    private String pharmacistName;
    
    @Column(nullable = false)
    private String department;
    
    @OneToMany(mappedBy = "prescription", cascade = CascadeType.ALL, orphanRemoval = true)
    @OrderColumn(name = "item_order")
    private List<PrescriptionItem> items = new ArrayList<>();
    
    @OneToMany(mappedBy = "prescription", cascade = CascadeType.ALL, orphanRemoval = true)
    @OrderColumn(name = "validation_order")
    private List<PrescriptionValidation> validations = new ArrayList<>();
    
    @OneToMany(mappedBy = "prescription", cascade = CascadeType.ALL, orphanRemoval = true)
    @OrderColumn(name = "signature_order")
    private List<PrescriptionSignature> signatures = new ArrayList<>();
    
    private boolean isOverDose;
    
    private String overDoseReason;
    
    private String consultationNotes;
    
    private LocalDateTime createdAt;
    
    private LocalDateTime updatedAt;
    
    private LocalDateTime submittedAt;
    
    private LocalDateTime approvedAt;
    
    @PrePersist
    protected void onCreate() {
        if (status == null) {
            status = PrescriptionStatus.DRAFT;
        }
        createdAt = LocalDateTime.now();
        updatedAt = LocalDateTime.now();
    }
    
    @PreUpdate
    protected void onUpdate() {
        updatedAt = LocalDateTime.now();
    }
    
    public void addItem(PrescriptionItem item) {
        items.add(item);
        item.setPrescription(this);
    }
    
    public void addValidation(PrescriptionValidation validation) {
        validations.add(validation);
        validation.setPrescription(this);
    }
    
    public void addSignature(PrescriptionSignature signature) {
        signatures.add(signature);
        signature.setPrescription(this);
    }
    
    public boolean hasPsychotropicOrNarcoticDrugs() {
        return items.stream()
                .anyMatch(item -> item.getDrug() != null && 
                        (item.getDrug().getCategory().getDescription().contains("精神") ||
                         item.getDrug().getCategory().getDescription().contains("麻醉")));
    }
    
    public int getRequiredSignatures() {
        return hasPsychotropicOrNarcoticDrugs() ? 2 : 1;
    }
}
