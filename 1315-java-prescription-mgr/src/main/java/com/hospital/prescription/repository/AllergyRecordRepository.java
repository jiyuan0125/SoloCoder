package com.hospital.prescription.repository;

import com.hospital.prescription.entity.AllergyRecord;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface AllergyRecordRepository extends JpaRepository<AllergyRecord, Long> {
    List<AllergyRecord> findByPatientId(Long patientId);
    
    @Query("SELECT ar FROM AllergyRecord ar WHERE ar.patient.id = :patientId AND (LOWER(ar.allergen) LIKE LOWER(CONCAT('%', :drugName, '%')) OR LOWER(:drugName) LIKE LOWER(CONCAT('%', ar.allergen, '%')))")
    List<AllergyRecord> findByPatientIdAndAllergenLike(@Param("patientId") Long patientId, @Param("drugName") String drugName);
}
