package com.company.vehicledispatch.controller;

import com.company.vehicledispatch.entity.FuelAbnormalRecord;
import com.company.vehicledispatch.service.FuelAbnormalService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/fuel-abnormal")
public class FuelAbnormalController {

    @Autowired
    private FuelAbnormalService fuelAbnormalService;

    @GetMapping
    public ResponseEntity<List<FuelAbnormalRecord>> getAllAbnormalRecords() {
        return ResponseEntity.ok(fuelAbnormalService.getAllAbnormalRecords());
    }

    @GetMapping("/{id}")
    public ResponseEntity<FuelAbnormalRecord> getAbnormalRecordById(@PathVariable Long id) {
        return fuelAbnormalService.getAbnormalRecordById(id)
                .map(ResponseEntity::ok)
                .orElse(ResponseEntity.notFound().build());
    }

    @GetMapping("/vehicle/{vehicleId}")
    public ResponseEntity<List<FuelAbnormalRecord>> getAbnormalRecordsByVehicle(@PathVariable Long vehicleId) {
        return ResponseEntity.ok(fuelAbnormalService.getAbnormalRecordsByVehicle(vehicleId));
    }

    @GetMapping("/dispatch-record/{dispatchRecordId}")
    public ResponseEntity<FuelAbnormalRecord> getAbnormalRecordByDispatchRecord(@PathVariable Long dispatchRecordId) {
        return fuelAbnormalService.getAbnormalRecordByDispatchRecord(dispatchRecordId)
                .map(ResponseEntity::ok)
                .orElse(ResponseEntity.notFound().build());
    }

    @PostMapping("/dispatch-record/{dispatchRecordId}")
    public ResponseEntity<FuelAbnormalRecord> createAbnormalRecord(
            @PathVariable Long dispatchRecordId,
            @RequestParam(required = false) String remarks) {
        FuelAbnormalRecord record = fuelAbnormalService.createAbnormalRecord(dispatchRecordId, remarks);
        return ResponseEntity.ok(record);
    }
}
