package com.hospital.prescription.repository;

import com.hospital.prescription.entity.Prescription;
import com.hospital.prescription.enums.PrescriptionStatus;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface PrescriptionRepository extends JpaRepository<Prescription, Long> {
    List<Prescription> findByPatientIdOrderByCreatedAtDesc(Long patientId);
    List<Prescription> findByStatus(PrescriptionStatus status);
    List<Prescription> findByDoctorNameOrderByCreatedAtDesc(String doctorName);
}
