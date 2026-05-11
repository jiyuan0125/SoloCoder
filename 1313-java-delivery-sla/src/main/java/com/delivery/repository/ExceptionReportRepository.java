package com.delivery.repository;

import com.delivery.entity.ExceptionReport;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;
import java.util.Optional;

@Repository
public interface ExceptionReportRepository extends JpaRepository<ExceptionReport, Long> {
    
    @Query("SELECT er FROM ExceptionReport er WHERE er.order.id = :orderId " +
           "AND er.approved = true AND er.isTimeoutExempt = true")
    Optional<ExceptionReport> findApprovedExemptionForOrder(@Param("orderId") Long orderId);
}
