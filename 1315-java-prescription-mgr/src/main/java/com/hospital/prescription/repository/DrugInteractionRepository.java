package com.hospital.prescription.repository;

import com.hospital.prescription.entity.Drug;
import com.hospital.prescription.entity.DrugInteraction;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface DrugInteractionRepository extends JpaRepository<DrugInteraction, Long> {
    @Query("SELECT di FROM DrugInteraction di WHERE (di.drugA = :drugA AND di.drugB = :drugB) OR (di.drugA = :drugB AND di.drugB = :drugA)")
    List<DrugInteraction> findInteractionBetween(@Param("drugA") Drug drugA, @Param("drugB") Drug drugB);
    
    @Query("SELECT di FROM DrugInteraction di WHERE di.drugA.id IN :drugIds OR di.drugB.id IN :drugIds")
    List<DrugInteraction> findByDrugIds(@Param("drugIds") List<Long> drugIds);
}
