package com.company.vehicledispatch.controller;

import com.company.vehicledispatch.dto.DispatchRequestDTO;
import com.company.vehicledispatch.dto.DispatchResultDTO;
import com.company.vehicledispatch.entity.DispatchRecord;
import com.company.vehicledispatch.entity.DispatchRequest;
import com.company.vehicledispatch.entity.Vehicle;
import com.company.vehicledispatch.enums.RequestStatus;
import com.company.vehicledispatch.service.DispatchService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.format.annotation.DateTimeFormat;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.time.LocalDateTime;
import java.util.List;

@RestController
@RequestMapping("/api/dispatch")
public class DispatchController {

    @Autowired
    private DispatchService dispatchService;

    @GetMapping("/requests")
    public ResponseEntity<List<DispatchRequest>> getAllRequests() {
        return ResponseEntity.ok(dispatchService.getAllRequests());
    }

    @GetMapping("/requests/{id}")
    public ResponseEntity<DispatchRequest> getRequestById(@PathVariable Long id) {
        return dispatchService.getRequestById(id)
                .map(ResponseEntity::ok)
                .orElse(ResponseEntity.notFound().build());
    }

    @GetMapping("/requests/status/{status}")
    public ResponseEntity<List<DispatchRequest>> getRequestsByStatus(@PathVariable RequestStatus status) {
        return ResponseEntity.ok(dispatchService.getRequestsByStatus(status));
    }

    @GetMapping("/requests/vehicle/{vehicleId}")
    public ResponseEntity<List<DispatchRequest>> getRequestsByVehicle(@PathVariable Long vehicleId) {
        return ResponseEntity.ok(dispatchService.getRequestsByVehicle(vehicleId));
    }

    @PostMapping("/requests")
    public ResponseEntity<DispatchResultDTO> createRequest(@Valid @RequestBody DispatchRequestDTO dto) {
        DispatchResultDTO result = dispatchService.createRequest(dto);
        return ResponseEntity.ok(result);
    }

    @GetMapping("/available-vehicles")
    public ResponseEntity<List<Vehicle>> findAvailableVehicles(
            @RequestParam Integer requiredSeats,
            @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME) LocalDateTime startDateTime,
            @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME) LocalDateTime endDateTime) {
        List<Vehicle> vehicles = dispatchService.findAvailableVehicles(requiredSeats, startDateTime, endDateTime);
        return ResponseEntity.ok(vehicles);
    }

    @PutMapping("/requests/{id}/approve")
    public ResponseEntity<DispatchRequest> approveRequest(
            @PathVariable Long id,
            @RequestParam(required = false) Long vehicleId) {
        DispatchRequest request = dispatchService.approveRequest(id, vehicleId);
        return ResponseEntity.ok(request);
    }

    @PutMapping("/requests/{id}/reject")
    public ResponseEntity<DispatchRequest> rejectRequest(
            @PathVariable Long id,
            @RequestParam(required = false) String reason) {
        DispatchRequest request = dispatchService.rejectRequest(id, reason);
        return ResponseEntity.ok(request);
    }

    @PostMapping("/requests/{id}/start")
    public ResponseEntity<DispatchRecord> startDispatch(@PathVariable Long id) {
        DispatchRecord record = dispatchService.startDispatch(id);
        return ResponseEntity.ok(record);
    }

    @PostMapping("/records/{recordId}/complete")
    public ResponseEntity<DispatchRecord> completeDispatch(
            @PathVariable Long recordId,
            @RequestParam Double endMileage,
            @RequestParam(required = false) Double actualFuelConsumption) {
        DispatchRecord record = dispatchService.completeDispatch(recordId, endMileage, actualFuelConsumption);
        return ResponseEntity.ok(record);
    }

    @PutMapping("/requests/{id}/cancel")
    public ResponseEntity<DispatchRequest> cancelRequest(@PathVariable Long id) {
        DispatchRequest request = dispatchService.cancelRequest(id);
        return ResponseEntity.ok(request);
    }
}
