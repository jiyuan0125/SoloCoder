package com.company.vehicledispatch.service;

import com.company.vehicledispatch.entity.DispatchRecord;
import com.company.vehicledispatch.entity.FuelAbnormalRecord;
import com.company.vehicledispatch.exception.BusinessException;
import com.company.vehicledispatch.repository.DispatchRecordRepository;
import com.company.vehicledispatch.repository.FuelAbnormalRecordRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;

@Service
public class FuelAbnormalService {

    @Autowired
    private FuelAbnormalRecordRepository fuelAbnormalRecordRepository;

    @Autowired
    private DispatchRecordRepository dispatchRecordRepository;

    public List<FuelAbnormalRecord> getAllAbnormalRecords() {
        return fuelAbnormalRecordRepository.findAll();
    }

    public Optional<FuelAbnormalRecord> getAbnormalRecordById(Long id) {
        return fuelAbnormalRecordRepository.findById(id);
    }

    public List<FuelAbnormalRecord> getAbnormalRecordsByVehicle(Long vehicleId) {
        return fuelAbnormalRecordRepository.findByVehicleId(vehicleId);
    }

    public Optional<FuelAbnormalRecord> getAbnormalRecordByDispatchRecord(Long dispatchRecordId) {
        return fuelAbnormalRecordRepository.findByDispatchRecordId(dispatchRecordId).stream().findFirst();
    }

    public FuelAbnormalRecord createAbnormalRecord(Long dispatchRecordId, String remarks) {
        DispatchRecord dispatchRecord = dispatchRecordRepository.findById(dispatchRecordId)
                .orElseThrow(() -> new BusinessException("调度记录不存在: " + dispatchRecordId));

        if (!Boolean.TRUE.equals(dispatchRecord.getFuelAbnormal())) {
            throw new BusinessException("该调度记录没有油耗异常");
        }

        Optional<FuelAbnormalRecord> existing = getAbnormalRecordByDispatchRecord(dispatchRecordId);
        if (existing.isPresent()) {
            return existing.get();
        }

        double estimated = dispatchRecord.getEstimatedFuelConsumption() != null ?
                dispatchRecord.getEstimatedFuelConsumption() : 0.0;
        double actual = dispatchRecord.getActualFuelConsumption() != null ?
                dispatchRecord.getActualFuelConsumption() : 0.0;
        double deviation = estimated > 0 ? Math.abs(actual - estimated) / estimated : 0.0;

        FuelAbnormalRecord record = new FuelAbnormalRecord();
        record.setDispatchRecord(dispatchRecord);
        record.setVehicle(dispatchRecord.getVehicle());
        record.setEstimatedFuelConsumption(estimated);
        record.setActualFuelConsumption(actual);
        record.setDeviationPercentage(deviation);
        record.setRemarks(remarks);
        record.setCreatedAt(LocalDateTime.now());

        return fuelAbnormalRecordRepository.save(record);
    }
}
