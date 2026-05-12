package com.health.controller;

import com.health.model.CheckHistory;
import com.health.model.ServiceRegistration;
import com.health.service.HealthMonitorService;
import com.health.store.HistoryStore;
import com.health.store.ServiceRegistry;
import jakarta.validation.Valid;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/health")
public class HealthMonitorController {

    private final ServiceRegistry serviceRegistry;
    private final HistoryStore historyStore;
    private final HealthMonitorService healthMonitorService;

    public HealthMonitorController(ServiceRegistry serviceRegistry,
                                   HistoryStore historyStore,
                                   HealthMonitorService healthMonitorService) {
        this.serviceRegistry = serviceRegistry;
        this.historyStore = historyStore;
        this.healthMonitorService = healthMonitorService;
    }

    @PostMapping("/register")
    public ResponseEntity<ServiceRegistration> register(@Valid @RequestBody ServiceRegistration service) {
        ServiceRegistration registered = serviceRegistry.register(service);
        return ResponseEntity.ok(registered);
    }

    @GetMapping("/services")
    public ResponseEntity<List<ServiceRegistration>> getAllServices() {
        return ResponseEntity.ok(serviceRegistry.getAll());
    }

    @GetMapping("/services/{id}")
    public ResponseEntity<ServiceRegistration> getServiceById(@PathVariable String id) {
        ServiceRegistration service = serviceRegistry.getById(id);
        if (service == null) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(service);
    }

    @DeleteMapping("/services/{id}")
    public ResponseEntity<Void> removeService(@PathVariable String id) {
        boolean removed = serviceRegistry.remove(id);
        if (removed) {
            return ResponseEntity.noContent().build();
        }
        return ResponseEntity.notFound().build();
    }

    @PostMapping("/services/{id}/check")
    public ResponseEntity<ServiceRegistration> triggerCheck(@PathVariable String id) {
        ServiceRegistration service = serviceRegistry.getById(id);
        if (service == null) {
            return ResponseEntity.notFound().build();
        }
        healthMonitorService.checkService(service);
        return ResponseEntity.ok(service);
    }

    @GetMapping("/services/{id}/history")
    public ResponseEntity<List<CheckHistory>> getServiceHistory(
            @PathVariable String id,
            @RequestParam(defaultValue = "100") int limit) {
        return ResponseEntity.ok(historyStore.getByServiceId(id, limit));
    }

    @GetMapping("/history")
    public ResponseEntity<List<CheckHistory>> getAllHistory(
            @RequestParam(defaultValue = "100") int limit) {
        return ResponseEntity.ok(historyStore.getAll(limit));
    }

    @GetMapping("/status")
    public ResponseEntity<String> getMonitorStatus() {
        long total = serviceRegistry.getAll().size();
        long healthy = serviceRegistry.getAll().stream()
                .filter(s -> s.getStatus() != null && 
                        s.getStatus().name().equals("HEALTHY"))
                .count();
        long unhealthy = serviceRegistry.getAll().stream()
                .filter(s -> s.getStatus() != null && 
                        s.getStatus().name().equals("UNHEALTHY"))
                .count();
        
        String status = String.format(
                "Health Monitor Running - Total: %d, Healthy: %d, Unhealthy: %d",
                total, healthy, unhealthy);
        return ResponseEntity.ok(status);
    }
}