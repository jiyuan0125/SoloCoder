package com.company.vehicledispatch.repository;

import com.company.vehicledispatch.entity.Maintenance;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.List;

@Repository
public interface MaintenanceRepository extends JpaRepository<Maintenance, Long> {
    List<Maintenance> findByVehicleId(Long vehicleId);
    
    @Query("SELECT m FROM Maintenance m WHERE m.vehicle.id = :vehicleId " +
           "AND m.status = 'IN_PROGRESS'")
    List<Maintenance> findActiveMaintenancesByVehicleId(@Param("vehicleId") Long vehicleId);
    
    @Query("SELECT m FROM Maintenance m WHERE m.vehicle.id = :vehicleId " +
           "AND m.startDate >= :startOfMonth AND m.startDate < :startOfNextMonth")
    List<Maintenance> findByVehicleIdAndMonth(@Param("vehicleId") Long vehicleId,
                                               @Param("startOfMonth") LocalDate startOfMonth,
                                               @Param("startOfNextMonth") LocalDate startOfNextMonth);
}
