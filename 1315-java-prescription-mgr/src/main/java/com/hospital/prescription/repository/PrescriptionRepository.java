package com.hospital.prescription.repository;

import com.hospital.prescription.entity.Prescription;
import com.hospital.prescription.enums.PrescriptionStatus;
import org.springframework.data.jpa.repository.EntityGraph;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

@Repository
public interface PrescriptionRepository extends JpaRepository<Prescription, Long> {
    
    @Query("SELECT DISTINCT p FROM Prescription p " +
           "LEFT JOIN FETCH p.patient " +
           "LEFT JOIN FETCH p.items i " +
           "LEFT JOIN FETCH i.drug " +
           "LEFT JOIN FETCH p.validations " +
           "LEFT JOIN FETCH p.signatures " +
           "WHERE p.id = :id")
    Optional<Prescription> findByIdWithDetails(@Param("id") Long id);
    
    @Query("SELECT DISTINCT p FROM Prescription p " +
           "LEFT JOIN FETCH p.patient " +
           "LEFT JOIN FETCH p.items i " +
           "LEFT JOIN FETCH i.drug " +
           "LEFT JOIN FETCH p.validations " +
           "LEFT JOIN FETCH p.signatures " +
           "WHERE p.patient.id = :patientId " +
           "ORDER BY p.createdAt DESC")
    List<Prescription> findByPatientIdWithDetails(@Param("patientId") Long patientId);
    
    @Query("SELECT DISTINCT p FROM Prescription p " +
           "LEFT JOIN FETCH p.patient " +
           "LEFT JOIN FETCH p.items i " +
           "LEFT JOIN FETCH i.drug " +
           "LEFT JOIN FETCH p.validations " +
           "LEFT JOIN FETCH p.signatures " +
           "WHERE p.status = :status")
    List<Prescription> findByStatusWithDetails(@Param("status") PrescriptionStatus status);
    
    List<Prescription> findByPatientIdOrderByCreatedAtDesc(Long patientId);
    List<Prescription> findByStatus(PrescriptionStatus status);
    List<Prescription> findByDoctorNameOrderByCreatedAtDesc(String doctorName);
}
