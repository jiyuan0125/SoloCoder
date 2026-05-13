package com.example.healthchecker.controller;

import com.example.healthchecker.model.AddDependencyRequest;
import com.example.healthchecker.model.HealthEvent;
import com.example.healthchecker.model.HealthStatus;
import com.example.healthchecker.model.ServiceRegistry;
import com.example.healthchecker.model.StatusReportRequest;
import com.example.healthchecker.service.EventLogService;
import com.example.healthchecker.service.HealthCheckService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Collection;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
public class HealthCheckController {

    private final HealthCheckService healthCheckService;
    private final EventLogService eventLogService;

    public HealthCheckController(HealthCheckService healthCheckService, EventLogService eventLogService) {
        this.healthCheckService = healthCheckService;
        this.eventLogService = eventLogService;
    }

    @PostMapping("/services/{name}/dependencies")
    public ResponseEntity<Map<String, Object>> addDependency(
            @PathVariable String name,
            @RequestBody AddDependencyRequest request) {
        healthCheckService.addDependency(name, request.getDependency_service_name());
        ServiceRegistry service = healthCheckService.getOrCreateService(name);
        return ResponseEntity.ok(toResponse(service));
    }

    @DeleteMapping("/services/{name}/dependencies/{dep}")
    public ResponseEntity<Map<String, Object>> removeDependency(
            @PathVariable String name,
            @PathVariable String dep) {
        healthCheckService.removeDependency(name, dep);
        ServiceRegistry service = healthCheckService.getOrCreateService(name);
        return ResponseEntity.ok(toResponse(service));
    }

    @PostMapping("/reports/{service}")
    public ResponseEntity<Map<String, Object>> reportStatus(
            @PathVariable String service,
            @RequestBody StatusReportRequest request) {
        HealthStatus status = parseStatus(request.getStatus());
        if (status == null) {
            Map<String, Object> error = new HashMap<>();
            error.put("error", "Invalid status. Must be one of: healthy, degraded, down");
            return ResponseEntity.badRequest().body(error);
        }
        healthCheckService.reportStatus(service, status);
        ServiceRegistry registry = healthCheckService.getOrCreateService(service);
        return ResponseEntity.ok(toResponse(registry));
    }

    @PutMapping("/services/{name}/maintenance")
    public ResponseEntity<Map<String, Object>> setMaintenance(
            @PathVariable String name,
            @RequestParam boolean enabled) {
        healthCheckService.setMaintenanceMode(name, enabled);
        ServiceRegistry service = healthCheckService.getOrCreateService(name);
        return ResponseEntity.ok(toResponse(service));
    }

    @GetMapping("/services")
    public ResponseEntity<Collection<Map<String, Object>>> getAllServices() {
        Collection<ServiceRegistry> services = healthCheckService.getAllServices();
        Collection<Map<String, Object>> result = services.stream()
                .map(this::toResponse)
                .toList();
        return ResponseEntity.ok(result);
    }

    @GetMapping("/services/{name}")
    public ResponseEntity<Map<String, Object>> getService(@PathVariable String name) {
        ServiceRegistry service = healthCheckService.getOrCreateService(name);
        return ResponseEntity.ok(toResponse(service));
    }

    @GetMapping("/events")
    public ResponseEntity<List<HealthEvent>> getAllEvents() {
        return ResponseEntity.ok(eventLogService.getAllEvents());
    }

    @GetMapping("/services/{name}/events")
    public ResponseEntity<List<HealthEvent>> getServiceEvents(@PathVariable String name) {
        return ResponseEntity.ok(eventLogService.getEventsByService(name));
    }

    private Map<String, Object> toResponse(ServiceRegistry service) {
        Map<String, Object> response = new HashMap<>();
        response.put("name", service.getName());
        response.put("status", service.getStatus().name().toLowerCase());
        response.put("inMaintenance", service.isInMaintenance());
        response.put("dependencies", service.getDependencies());
        response.put("lastStatusChangeTime", service.getLastStatusChangeTime().toString());
        return response;
    }

    private HealthStatus parseStatus(String status) {
        if (status == null) {
            return null;
        }
        return switch (status.toLowerCase()) {
            case "healthy" -> HealthStatus.HEALTHY;
            case "degraded" -> HealthStatus.DEGRADED;
            case "down" -> HealthStatus.DOWN;
            case "maintenance" -> HealthStatus.MAINTENANCE;
            default -> null;
        };
    }
}
