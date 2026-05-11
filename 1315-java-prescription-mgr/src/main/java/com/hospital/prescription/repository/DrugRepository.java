package com.hospital.prescription.repository;

import com.hospital.prescription.entity.Drug;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.Optional;

@Repository
public interface DrugRepository extends JpaRepository<Drug, Long> {
    Optional<Drug> findByNameAndDosageForm(String name, String dosageForm);
    Optional<Drug> findByName(String name);
}
