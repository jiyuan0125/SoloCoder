package com.healthcheck.controller;

import com.healthcheck.model.*;
import com.healthcheck.scheduler.HealthCheckScheduler;
import com.healthcheck.service.HealthCheckService;
import com.healthcheck.service.HistoryService;
import com.healthcheck.service.ServiceConfigStore;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;
import java.util.Optional;

@RestController
@RequestMapping("/api")
public class HealthCheckController {

    private final ServiceConfigStore configStore;
    private final HealthCheckService healthCheckService;
    private final HistoryService historyService;
    private final HealthCheckScheduler scheduler;

    public HealthCheckController(ServiceConfigStore configStore,
                                 HealthCheckService healthCheckService,
                                 HistoryService historyService,
                                 HealthCheckScheduler scheduler) {
        this.configStore = configStore;
        this.healthCheckService = healthCheckService;
        this.historyService = historyService;
        this.scheduler = scheduler;
    }

    @PostMapping("/services")
    public ResponseEntity<ServiceConfig> registerService(@RequestBody ServiceConfig config) {
        configStore.registerService(config);
        scheduler.scheduleService(config);
        return ResponseEntity.ok(config);
    }

    @DeleteMapping("/services/{serviceId}")
    public ResponseEntity<Void> unregisterService(@PathVariable String serviceId) {
        if (!configStore.exists(serviceId)) {
            return ResponseEntity.notFound().build();
        }
        scheduler.unscheduleService(serviceId);
        configStore.unregisterService(serviceId);
        return ResponseEntity.noContent().build();
    }

    @GetMapping("/services")
    public ResponseEntity<List<ServiceConfig>> getAllServices() {
        return ResponseEntity.ok(configStore.getAllServices());
    }

    @GetMapping("/services/{serviceId}")
    public ResponseEntity<ServiceConfig> getService(@PathVariable String serviceId) {
        Optional<ServiceConfig> config = configStore.getServiceConfig(serviceId);
        return config.map(ResponseEntity::ok).orElseGet(() -> ResponseEntity.notFound().build());
    }

    @GetMapping("/services/{serviceId}/state")
    public ResponseEntity<ServiceState> getServiceState(@PathVariable String serviceId) {
        Optional<ServiceState> state = healthCheckService.getServiceState(serviceId);
        return state.map(ResponseEntity::ok).orElseGet(() -> ResponseEntity.notFound().build());
    }

    @GetMapping("/services/states")
    public ResponseEntity<Map<String, ServiceState>> getAllServiceStates() {
        return ResponseEntity.ok(healthCheckService.getAllServiceStates());
    }

    @GetMapping("/services/{serviceId}/check-items/{checkItemName}/history")
    public ResponseEntity<List<CheckResult>> getCheckItemHistory(
            @PathVariable String serviceId,
            @PathVariable String checkItemName,
            @RequestParam(defaultValue = "10") int n) {
        List<CheckResult> history = historyService.getHistory(serviceId, checkItemName, n);
        return ResponseEntity.ok(history);
    }

    @GetMapping("/services/{serviceId}/history")
    public ResponseEntity<List<CheckResult>> getServiceHistory(
            @PathVariable String serviceId,
            @RequestParam(defaultValue = "50") int n) {
        List<CheckResult> history = historyService.getServiceHistory(serviceId, n);
        return ResponseEntity.ok(history);
    }

    @PostMapping("/services/{serviceId}/check")
    public ResponseEntity<Void> triggerCheck(@PathVariable String serviceId) {
        if (!configStore.exists(serviceId)) {
            return ResponseEntity.notFound().build();
        }
        healthCheckService.checkService(serviceId);
        return ResponseEntity.noContent().build();
    }

    @PutMapping("/services/{serviceId}")
    public ResponseEntity<ServiceConfig> updateService(@PathVariable String serviceId,
                                                       @RequestBody ServiceConfig config) {
        if (!configStore.exists(serviceId)) {
            return ResponseEntity.notFound().build();
        }
        if (!serviceId.equals(config.getServiceId())) {
            return ResponseEntity.badRequest().build();
        }
        configStore.registerService(config);
        scheduler.scheduleService(config);
        return ResponseEntity.ok(config);
    }
}
