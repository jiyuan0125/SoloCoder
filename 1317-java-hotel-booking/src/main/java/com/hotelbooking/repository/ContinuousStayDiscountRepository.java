package com.hotelbooking.repository;

import com.hotelbooking.model.entity.ContinuousStayDiscount;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

@Repository
public interface ContinuousStayDiscountRepository extends JpaRepository<ContinuousStayDiscount, Long> {
    List<ContinuousStayDiscount> findByIsActiveTrueOrderByMinDaysAsc();
    
    @Query("SELECT d FROM ContinuousStayDiscount d WHERE d.isActive = true " +
           "AND :days >= d.minDays AND :days <= d.maxDays")
    Optional<ContinuousStayDiscount> findByDays(@Param("days") int days);
}
