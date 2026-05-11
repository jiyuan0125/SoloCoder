package com.company.vehicledispatch.service;

import com.company.vehicledispatch.config.AppConfig;
import com.company.vehicledispatch.dto.VehicleDTO;
import com.company.vehicledispatch.entity.Vehicle;
import com.company.vehicledispatch.enums.VehicleStatus;
import com.company.vehicledispatch.exception.BusinessException;
import com.company.vehicledispatch.repository.VehicleRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDate;
import java.util.List;
import java.util.Optional;

@Service
public class VehicleService {

    @Autowired
    private VehicleRepository vehicleRepository;

    @Autowired
    private AppConfig appConfig;

    public List<Vehicle> getAllVehicles() {
        return vehicleRepository.findAll();
    }

    public Optional<Vehicle> getVehicleById(Long id) {
        return vehicleRepository.findById(id);
    }

    public Optional<Vehicle> getVehicleByPlateNumber(String plateNumber) {
        return vehicleRepository.findByPlateNumber(plateNumber);
    }

    @Transactional
    public Vehicle createVehicle(VehicleDTO dto) {
        if (vehicleRepository.findByPlateNumber(dto.getPlateNumber()).isPresent()) {
            throw new BusinessException("车牌号已存在: " + dto.getPlateNumber());
        }

        Vehicle vehicle = new Vehicle();
        vehicle.setPlateNumber(dto.getPlateNumber());
        vehicle.setModel(dto.getModel());
        vehicle.setSeats(dto.getSeats());
        vehicle.setFuelType(dto.getFuelType());
        vehicle.setFuelLevel(dto.getFuelLevel());
        vehicle.setMaxFuelLevel(dto.getMaxFuelLevel());
        vehicle.setCurrentMileage(dto.getCurrentMileage());
        vehicle.setLastMaintenanceMileage(dto.getLastMaintenanceMileage() != null ? dto.getLastMaintenanceMileage() : dto.getCurrentMileage());
        vehicle.setInsuranceExpiryDate(dto.getInsuranceExpiryDate());
        vehicle.setAnnualInspectionExpiryDate(dto.getAnnualInspectionExpiryDate());
        vehicle.setStatus(dto.getStatus() != null ? dto.getStatus() : VehicleStatus.AVAILABLE);
        vehicle.setAverageFuelConsumption(dto.getAverageFuelConsumption() != null ? dto.getAverageFuelConsumption() : 8.0);

        updateVehicleStatuses(vehicle);

        return vehicleRepository.save(vehicle);
    }

    @Transactional
    public Vehicle updateVehicle(Long id, VehicleDTO dto) {
        Vehicle vehicle = vehicleRepository.findById(id)
                .orElseThrow(() -> new BusinessException("车辆不存在: " + id));

        if (!vehicle.getPlateNumber().equals(dto.getPlateNumber()) &&
                vehicleRepository.findByPlateNumber(dto.getPlateNumber()).isPresent()) {
            throw new BusinessException("车牌号已存在: " + dto.getPlateNumber());
        }

        vehicle.setPlateNumber(dto.getPlateNumber());
        vehicle.setModel(dto.getModel());
        vehicle.setSeats(dto.getSeats());
        vehicle.setFuelType(dto.getFuelType());
        vehicle.setFuelLevel(dto.getFuelLevel());
        vehicle.setMaxFuelLevel(dto.getMaxFuelLevel());
        vehicle.setCurrentMileage(dto.getCurrentMileage());
        if (dto.getLastMaintenanceMileage() != null) {
            vehicle.setLastMaintenanceMileage(dto.getLastMaintenanceMileage());
        }
        vehicle.setInsuranceExpiryDate(dto.getInsuranceExpiryDate());
        vehicle.setAnnualInspectionExpiryDate(dto.getAnnualInspectionExpiryDate());
        if (dto.getStatus() != null) {
            vehicle.setStatus(dto.getStatus());
        }
        if (dto.getAverageFuelConsumption() != null) {
            vehicle.setAverageFuelConsumption(dto.getAverageFuelConsumption());
        }

        updateVehicleStatuses(vehicle);

        return vehicleRepository.save(vehicle);
    }

    @Transactional
    public Vehicle updateFuelLevel(Long id, Double fuelLevel) {
        Vehicle vehicle = vehicleRepository.findById(id)
                .orElseThrow(() -> new BusinessException("车辆不存在: " + id));

        if (fuelLevel < 0 || fuelLevel > vehicle.getMaxFuelLevel()) {
            throw new BusinessException("油量/电量必须在 0 到 " + vehicle.getMaxFuelLevel() + " 之间");
        }

        vehicle.setFuelLevel(fuelLevel);
        updateVehicleStatuses(vehicle);

        return vehicleRepository.save(vehicle);
    }

    @Transactional
    public Vehicle updateMileage(Long id, Double mileage) {
        Vehicle vehicle = vehicleRepository.findById(id)
                .orElseThrow(() -> new BusinessException("车辆不存在: " + id));

        if (mileage < vehicle.getCurrentMileage()) {
            throw new BusinessException("里程数不能小于当前里程数");
        }

        vehicle.setCurrentMileage(mileage);
        updateVehicleStatuses(vehicle);

        return vehicleRepository.save(vehicle);
    }

    @Transactional
    public void deleteVehicle(Long id) {
        if (!vehicleRepository.existsById(id)) {
            throw new BusinessException("车辆不存在: " + id);
        }
        vehicleRepository.deleteById(id);
    }

    public void updateVehicleStatuses(Vehicle vehicle) {
        double fuelPercentage = (vehicle.getFuelLevel() / vehicle.getMaxFuelLevel()) * 100;
        vehicle.setFuelWarning(fuelPercentage <= appConfig.getVehicle().getFuelWarningThreshold());

        double maintenanceInterval = appConfig.getVehicle().getMaintenanceIntervalKm();
        double mileageSinceLastMaintenance = vehicle.getCurrentMileage() - vehicle.getLastMaintenanceMileage();
        vehicle.setMaintenanceDue(mileageSinceLastMaintenance >= maintenanceInterval);

        LocalDate today = LocalDate.now();
        int reminderDays = appConfig.getInsuranceReminderDays();
        vehicle.setInsuranceDue(vehicle.getInsuranceExpiryDate().isBefore(today.plusDays(reminderDays)) ||
                vehicle.getInsuranceExpiryDate().isEqual(today.plusDays(reminderDays)) ||
                vehicle.getInsuranceExpiryDate().isBefore(today));
        vehicle.setAnnualInspectionDue(vehicle.getAnnualInspectionExpiryDate().isBefore(today.plusDays(reminderDays)) ||
                vehicle.getAnnualInspectionExpiryDate().isEqual(today.plusDays(reminderDays)) ||
                vehicle.getAnnualInspectionExpiryDate().isBefore(today));
    }

    public List<Vehicle> getVehiclesWithFuelWarning() {
        return vehicleRepository.findByFuelWarningTrue();
    }

    public List<Vehicle> getVehiclesWithMaintenanceDue() {
        return vehicleRepository.findByMaintenanceDueTrue();
    }

    public List<Vehicle> getVehiclesWithInsuranceDue() {
        return vehicleRepository.findByInsuranceDueTrueOrAnnualInspectionDueTrue();
    }
}
