package com.company.vehicledispatch.controller;

import com.company.vehicledispatch.dto.VehicleDTO;
import com.company.vehicledispatch.entity.Vehicle;
import com.company.vehicledispatch.service.VehicleService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/vehicles")
public class VehicleController {

    @Autowired
    private VehicleService vehicleService;

    @GetMapping
    public ResponseEntity<List<Vehicle>> getAllVehicles() {
        return ResponseEntity.ok(vehicleService.getAllVehicles());
    }

    @GetMapping("/{id}")
    public ResponseEntity<Vehicle> getVehicleById(@PathVariable Long id) {
        return vehicleService.getVehicleById(id)
                .map(ResponseEntity::ok)
                .orElse(ResponseEntity.notFound().build());
    }

    @GetMapping("/plate/{plateNumber}")
    public ResponseEntity<Vehicle> getVehicleByPlateNumber(@PathVariable String plateNumber) {
        return vehicleService.getVehicleByPlateNumber(plateNumber)
                .map(ResponseEntity::ok)
                .orElse(ResponseEntity.notFound().build());
    }

    @PostMapping
    public ResponseEntity<Vehicle> createVehicle(@Valid @RequestBody VehicleDTO dto) {
        Vehicle vehicle = vehicleService.createVehicle(dto);
        return ResponseEntity.ok(vehicle);
    }

    @PutMapping("/{id}")
    public ResponseEntity<Vehicle> updateVehicle(@PathVariable Long id, @Valid @RequestBody VehicleDTO dto) {
        Vehicle vehicle = vehicleService.updateVehicle(id, dto);
        return ResponseEntity.ok(vehicle);
    }

    @PatchMapping("/{id}/fuel")
    public ResponseEntity<Vehicle> updateFuelLevel(@PathVariable Long id, @RequestParam Double fuelLevel) {
        Vehicle vehicle = vehicleService.updateFuelLevel(id, fuelLevel);
        return ResponseEntity.ok(vehicle);
    }

    @PatchMapping("/{id}/mileage")
    public ResponseEntity<Vehicle> updateMileage(@PathVariable Long id, @RequestParam Double mileage) {
        Vehicle vehicle = vehicleService.updateMileage(id, mileage);
        return ResponseEntity.ok(vehicle);
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Map<String, Object>> deleteVehicle(@PathVariable Long id) {
        vehicleService.deleteVehicle(id);
        Map<String, Object> response = new HashMap<>();
        response.put("success", true);
        response.put("message", "车辆删除成功");
        return ResponseEntity.ok(response);
    }

    @GetMapping("/warnings/fuel")
    public ResponseEntity<List<Vehicle>> getVehiclesWithFuelWarning() {
        return ResponseEntity.ok(vehicleService.getVehiclesWithFuelWarning());
    }

    @GetMapping("/warnings/maintenance")
    public ResponseEntity<List<Vehicle>> getVehiclesWithMaintenanceDue() {
        return ResponseEntity.ok(vehicleService.getVehiclesWithMaintenanceDue());
    }

    @GetMapping("/warnings/insurance")
    public ResponseEntity<List<Vehicle>> getVehiclesWithInsuranceDue() {
        return ResponseEntity.ok(vehicleService.getVehiclesWithInsuranceDue());
    }
}
