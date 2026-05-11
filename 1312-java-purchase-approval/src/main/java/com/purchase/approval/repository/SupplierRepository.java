package com.purchase.approval.repository;

import com.purchase.approval.entity.Supplier;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.LocalDate;
import java.util.List;
import java.util.Optional;

@Repository
public interface SupplierRepository extends JpaRepository<Supplier, Long> {
    Optional<Supplier> findBySupplierCode(String supplierCode);
    List<Supplier> findByIsBlacklistedTrue();
    
    @Query("SELECT s FROM Supplier s WHERE s.isBlacklisted = true AND s.blacklistEndDate <= :currentDate")
    List<Supplier> findBlacklistedSuppliersExpiringBefore(@Param("currentDate") LocalDate currentDate);
}
