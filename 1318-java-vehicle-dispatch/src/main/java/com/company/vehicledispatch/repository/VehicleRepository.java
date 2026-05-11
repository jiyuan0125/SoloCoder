package com.company.vehicledispatch.repository;

import com.company.vehicledispatch.entity.Vehicle;
import com.company.vehicledispatch.enums.VehicleStatus;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

@Repository
public interface VehicleRepository extends JpaRepository<Vehicle, Long> {
    Optional<Vehicle> findByPlateNumber(String plateNumber);
    
    List<Vehicle> findByStatus(VehicleStatus status);
    
    List<Vehicle> findByFuelWarningTrue();
    
    List<Vehicle> findByMaintenanceDueTrue();
    
    List<Vehicle> findByInsuranceDueTrueOrAnnualInspectionDueTrue();
}
