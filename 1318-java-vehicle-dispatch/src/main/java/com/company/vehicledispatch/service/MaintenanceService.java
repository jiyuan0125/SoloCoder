package com.company.vehicledispatch.service;

import com.company.vehicledispatch.config.AppConfig;
import com.company.vehicledispatch.dto.MaintenanceDTO;
import com.company.vehicledispatch.entity.Maintenance;
import com.company.vehicledispatch.entity.Vehicle;
import com.company.vehicledispatch.enums.VehicleStatus;
import com.company.vehicledispatch.exception.BusinessException;
import com.company.vehicledispatch.repository.MaintenanceRepository;
import com.company.vehicledispatch.repository.VehicleRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;

@Service
public class MaintenanceService {

    @Autowired
    private MaintenanceRepository maintenanceRepository;

    @Autowired
    private VehicleRepository vehicleRepository;

    @Autowired
    private AppConfig appConfig;

    public List<Maintenance> getAllMaintenances() {
        return maintenanceRepository.findAll();
    }

    public Optional<Maintenance> getMaintenanceById(Long id) {
        return maintenanceRepository.findById(id);
    }

    public List<Maintenance> getMaintenancesByVehicle(Long vehicleId) {
        return maintenanceRepository.findByVehicleId(vehicleId);
    }

    public List<Maintenance> getActiveMaintenancesByVehicle(Long vehicleId) {
        return maintenanceRepository.findActiveMaintenancesByVehicleId(vehicleId);
    }

    @Transactional
    public Maintenance createMaintenance(MaintenanceDTO dto) {
        Vehicle vehicle = vehicleRepository.findById(dto.getVehicleId())
                .orElseThrow(() -> new BusinessException("车辆不存在: " + dto.getVehicleId()));

        Maintenance maintenance = new Maintenance();
        maintenance.setVehicle(vehicle);
        maintenance.setType(dto.getType());
        maintenance.setDescription(dto.getDescription());
        maintenance.setScheduledDate(dto.getScheduledDate());
        maintenance.setStartDate(dto.getStartDate());
        maintenance.setEndDate(dto.getEndDate());
        maintenance.setCost(dto.getCost());
        maintenance.setStatus(dto.getStatus() != null ? dto.getStatus() : "SCHEDULED");
        maintenance.setRemarks(dto.getRemarks());
        maintenance.setCreatedAt(LocalDateTime.now());
        maintenance.setUpdatedAt(LocalDateTime.now());

        if ("IN_PROGRESS".equals(maintenance.getStatus())) {
            vehicle.setStatus(VehicleStatus.IN_MAINTENANCE);
            vehicleRepository.save(vehicle);
        }

        return maintenanceRepository.save(maintenance);
    }

    @Transactional
    public Maintenance startMaintenance(Long id) {
        Maintenance maintenance = maintenanceRepository.findById(id)
                .orElseThrow(() -> new BusinessException("维修记录不存在: " + id));

        if (!"SCHEDULED".equals(maintenance.getStatus())) {
            throw new BusinessException("只有计划中的维修才能开始");
        }

        maintenance.setStartDate(LocalDate.now());
        maintenance.setStatus("IN_PROGRESS");
        maintenance.setUpdatedAt(LocalDateTime.now());

        Vehicle vehicle = maintenance.getVehicle();
        vehicle.setStatus(VehicleStatus.IN_MAINTENANCE);
        vehicleRepository.save(vehicle);

        return maintenanceRepository.save(maintenance);
    }

    @Transactional
    public Maintenance completeMaintenance(Long id, Double cost, String remarks) {
        Maintenance maintenance = maintenanceRepository.findById(id)
                .orElseThrow(() -> new BusinessException("维修记录不存在: " + id));

        if (!"IN_PROGRESS".equals(maintenance.getStatus())) {
            throw new BusinessException("只有进行中的维修才能完成");
        }

        maintenance.setEndDate(LocalDate.now());
        maintenance.setCost(cost);
        maintenance.setRemarks(remarks);
        maintenance.setStatus("COMPLETED");
        maintenance.setUpdatedAt(LocalDateTime.now());

        Vehicle vehicle = maintenance.getVehicle();
        vehicle.setStatus(VehicleStatus.AVAILABLE);

        double maintenanceInterval = appConfig.getVehicle().getMaintenanceIntervalKm();
        double mileageSinceLastMaintenance = vehicle.getCurrentMileage() - vehicle.getLastMaintenanceMileage();
        if (mileageSinceLastMaintenance >= maintenanceInterval) {
            vehicle.setLastMaintenanceMileage(vehicle.getCurrentMileage());
            vehicle.setMaintenanceDue(false);
        }

        vehicleRepository.save(vehicle);

        return maintenanceRepository.save(maintenance);
    }

    @Transactional
    public void deleteMaintenance(Long id) {
        if (!maintenanceRepository.existsById(id)) {
            throw new BusinessException("维修记录不存在: " + id);
        }
        maintenanceRepository.deleteById(id);
    }
}
