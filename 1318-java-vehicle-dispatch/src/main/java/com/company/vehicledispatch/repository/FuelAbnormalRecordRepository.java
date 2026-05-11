package com.company.vehicledispatch.repository;

import com.company.vehicledispatch.entity.FuelAbnormalRecord;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface FuelAbnormalRecordRepository extends JpaRepository<FuelAbnormalRecord, Long> {
    List<FuelAbnormalRecord> findByVehicleId(Long vehicleId);
    
    List<FuelAbnormalRecord> findByDispatchRecordId(Long dispatchRecordId);
}
