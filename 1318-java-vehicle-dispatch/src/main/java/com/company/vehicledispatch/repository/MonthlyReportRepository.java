package com.company.vehicledispatch.repository;

import com.company.vehicledispatch.entity.MonthlyReport;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

@Repository
public interface MonthlyReportRepository extends JpaRepository<MonthlyReport, Long> {
    List<MonthlyReport> findByVehicleId(Long vehicleId);
    
    Optional<MonthlyReport> findByVehicleIdAndReportYearAndReportMonth(Long vehicleId, Integer reportYear, Integer reportMonth);
    
    List<MonthlyReport> findByReportYearAndReportMonth(Integer reportYear, Integer reportMonth);
}
