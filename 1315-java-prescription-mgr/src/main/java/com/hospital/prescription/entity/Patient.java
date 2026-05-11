package com.hospital.prescription.entity;

import jakarta.persistence.*;
import lombok.Data;

import java.time.LocalDate;
import java.time.LocalDateTime;
import java.time.Period;

@Entity
@Data
@Table(name = "patients")
public class Patient {
    
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    
    @Column(nullable = false)
    private String name;
    
    @Column(nullable = false)
    private String idCard;
    
    @Column(nullable = false)
    private LocalDate dateOfBirth;
    
    @Column(nullable = false)
    private String gender;
    
    private String phone;
    
    private String address;
    
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
    
    public int getAgeInYears() {
        return Period.between(dateOfBirth, LocalDate.now()).getYears();
    }
    
    public int getAgeInDays() {
        return (int) (LocalDate.now().toEpochDay() - dateOfBirth.toEpochDay());
    }
    
    public boolean isChild() {
        return getAgeInYears() < 18;
    }
    
    public boolean isElderly() {
        return getAgeInYears() >= 65;
    }
}
