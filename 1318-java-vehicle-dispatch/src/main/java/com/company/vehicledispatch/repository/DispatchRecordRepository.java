package com.company.vehicledispatch.repository;

import com.company.vehicledispatch.entity.DispatchRecord;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.List;

@Repository
public interface DispatchRecordRepository extends JpaRepository<DispatchRecord, Long> {
    List<DispatchRecord> findByVehicleId(Long vehicleId);
    
    List<DispatchRecord> findByRequestId(Long requestId);
    
    List<DispatchRecord> findByFuelAbnormalTrue();
    
    @Query("SELECT dr FROM DispatchRecord dr WHERE dr.vehicle.id = :vehicleId " +
           "AND dr.actualStartDateTime >= :startOfMonth " +
           "AND dr.actualStartDateTime < :startOfNextMonth")
    List<DispatchRecord> findByVehicleIdAndMonth(@Param("vehicleId") Long vehicleId,
                                                  @Param("startOfMonth") LocalDateTime startOfMonth,
                                                  @Param("startOfNextMonth") LocalDateTime startOfNextMonth);
}
