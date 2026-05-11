package com.company.vehicledispatch.controller;

import com.company.vehicledispatch.dto.MaintenanceDTO;
import com.company.vehicledispatch.entity.Maintenance;
import com.company.vehicledispatch.service.MaintenanceService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/maintenances")
public class MaintenanceController {

    @Autowired
    private MaintenanceService maintenanceService;

    @GetMapping
    public ResponseEntity<List<Maintenance>> getAllMaintenances() {
        return ResponseEntity.ok(maintenanceService.getAllMaintenances());
    }

    @GetMapping("/{id}")
    public ResponseEntity<Maintenance> getMaintenanceById(@PathVariable Long id) {
        return maintenanceService.getMaintenanceById(id)
                .map(ResponseEntity::ok)
                .orElse(ResponseEntity.notFound().build());
    }

    @GetMapping("/vehicle/{vehicleId}")
    public ResponseEntity<List<Maintenance>> getMaintenancesByVehicle(@PathVariable Long vehicleId) {
        return ResponseEntity.ok(maintenanceService.getMaintenancesByVehicle(vehicleId));
    }

    @GetMapping("/vehicle/{vehicleId}/active")
    public ResponseEntity<List<Maintenance>> getActiveMaintenancesByVehicle(@PathVariable Long vehicleId) {
        return ResponseEntity.ok(maintenanceService.getActiveMaintenancesByVehicle(vehicleId));
    }

    @PostMapping
    public ResponseEntity<Maintenance> createMaintenance(@Valid @RequestBody MaintenanceDTO dto) {
        Maintenance maintenance = maintenanceService.createMaintenance(dto);
        return ResponseEntity.ok(maintenance);
    }

    @PutMapping("/{id}/start")
    public ResponseEntity<Maintenance> startMaintenance(@PathVariable Long id) {
        Maintenance maintenance = maintenanceService.startMaintenance(id);
        return ResponseEntity.ok(maintenance);
    }

    @PutMapping("/{id}/complete")
    public ResponseEntity<Maintenance> completeMaintenance(
            @PathVariable Long id,
            @RequestParam(required = false) Double cost,
            @RequestParam(required = false) String remarks) {
        Maintenance maintenance = maintenanceService.completeMaintenance(id, cost, remarks);
        return ResponseEntity.ok(maintenance);
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Map<String, Object>> deleteMaintenance(@PathVariable Long id) {
        maintenanceService.deleteMaintenance(id);
        Map<String, Object> response = new HashMap<>();
        response.put("success", true);
        response.put("message", "维修记录删除成功");
        return ResponseEntity.ok(response);
    }
}
