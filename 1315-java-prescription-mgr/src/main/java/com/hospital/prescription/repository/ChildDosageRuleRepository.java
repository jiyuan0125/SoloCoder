package com.hospital.prescription.repository;

import com.hospital.prescription.entity.ChildDosageRule;
import com.hospital.prescription.enums.ChildAgeGroup;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.Optional;

@Repository
public interface ChildDosageRuleRepository extends JpaRepository<ChildDosageRule, Long> {
    Optional<ChildDosageRule> findByDrugIdAndAgeGroup(Long drugId, ChildAgeGroup ageGroup);
}
