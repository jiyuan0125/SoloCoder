package com.loadbalancer.controller;

import com.loadbalancer.model.Instance;
import com.loadbalancer.registry.InstanceRegistry;
import com.loadbalancer.service.LoadBalancerService;
import jakarta.validation.Valid;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Positive;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/admin")
public class ManagementController {
    private final InstanceRegistry instanceRegistry;
    private final LoadBalancerService loadBalancerService;

    public ManagementController(InstanceRegistry instanceRegistry, LoadBalancerService loadBalancerService) {
        this.instanceRegistry = instanceRegistry;
        this.loadBalancerService = loadBalancerService;
    }

    @GetMapping("/instances")
    public ResponseEntity<List<Map<String, Object>>> getAllInstances() {
        List<Map<String, Object>> instances = instanceRegistry.getAllInstances().stream()
                .map(this::buildInstanceInfo)
                .toList();
        return ResponseEntity.ok(instances);
    }

    @GetMapping("/instances/{id}")
    public ResponseEntity<Map<String, Object>> getInstance(@PathVariable String id) {
        Instance instance = instanceRegistry.getInstance(id);
        if (instance == null) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(buildInstanceInfo(instance));
    }

    @GetMapping("/instances/{id}/stats")
    public ResponseEntity<Map<String, Object>> getInstanceStats(@PathVariable String id) {
        Instance instance = instanceRegistry.getInstance(id);
        if (instance == null) {
            return ResponseEntity.notFound().build();
        }

        Map<String, Object> stats = new HashMap<>();
        stats.put("id", instance.getId());
        stats.put("totalRequests", instance.getTotalRequests());
        stats.put("successfulRequests", instance.getSuccessfulRequests());
        stats.put("failedRequests", instance.getFailedRequests());
        stats.put("activeRequests", instance.getActiveRequests());
        stats.put("successRate", String.format("%.2f%%", instance.getSuccessRate()));

        return ResponseEntity.ok(stats);
    }

    @GetMapping("/stats")
    public ResponseEntity<Map<String, Object>> getOverallStats() {
        long totalRequests = 0;
        long successfulRequests = 0;
        long failedRequests = 0;
        int activeRequests = 0;
        int availableCount = 0;

        List<Instance> allInstances = instanceRegistry.getAllInstances();
        for (Instance instance : allInstances) {
            totalRequests += instance.getTotalRequests();
            successfulRequests += instance.getSuccessfulRequests();
            failedRequests += instance.getFailedRequests();
            activeRequests += instance.getActiveRequests();
            if (instanceRegistry.isInstanceAvailable(instance.getId())) {
                availableCount++;
            }
        }

        double successRate = totalRequests == 0 ? 100.0 : (successfulRequests * 100.0) / totalRequests;

        Map<String, Object> stats = new HashMap<>();
        stats.put("currentStrategy", loadBalancerService.getCurrentStrategy());
        stats.put("totalInstances", allInstances.size());
        stats.put("availableInstances", availableCount);
        stats.put("totalRequests", totalRequests);
        stats.put("successfulRequests", successfulRequests);
        stats.put("failedRequests", failedRequests);
        stats.put("activeRequests", activeRequests);
        stats.put("overallSuccessRate", String.format("%.2f%%", successRate));

        return ResponseEntity.ok(stats);
    }

    @PostMapping("/instances/register")
    public ResponseEntity<Map<String, Object>> registerInstance(@Valid @RequestBody RegisterRequest request) {
        String id = request.id != null ? request.id : request.host + ":" + request.port;
        instanceRegistry.registerInstance(id, request.host, request.port);
        
        Instance instance = instanceRegistry.getInstance(id);
        return ResponseEntity.ok(buildInstanceInfo(instance));
    }

    @DeleteMapping("/instances/{id}")
    public ResponseEntity<Void> deregisterInstance(@PathVariable String id) {
        if (instanceRegistry.getInstance(id) == null) {
            return ResponseEntity.notFound().build();
        }
        instanceRegistry.deregisterInstance(id);
        return ResponseEntity.noContent().build();
    }

    @PostMapping("/instances/{id}/offline")
    public ResponseEntity<Map<String, Object>> markOffline(@PathVariable String id) {
        Instance instance = instanceRegistry.getInstance(id);
        if (instance == null) {
            return ResponseEntity.notFound().build();
        }
        instanceRegistry.markInstanceOffline(id);
        return ResponseEntity.ok(buildInstanceInfo(instance));
    }

    @PostMapping("/instances/{id}/online")
    public ResponseEntity<Map<String, Object>> markOnline(@PathVariable String id) {
        Instance instance = instanceRegistry.getInstance(id);
        if (instance == null) {
            return ResponseEntity.notFound().build();
        }
        instanceRegistry.markInstanceOnline(id);
        return ResponseEntity.ok(buildInstanceInfo(instance));
    }

    @PostMapping("/strategy")
    public ResponseEntity<Map<String, String>> setStrategy(@RequestBody StrategyRequest request) {
        try {
            loadBalancerService.setStrategy(request.strategy);
            Map<String, String> response = new HashMap<>();
            response.put("currentStrategy", loadBalancerService.getCurrentStrategy());
            return ResponseEntity.ok(response);
        } catch (IllegalArgumentException e) {
            Map<String, String> response = new HashMap<>();
            response.put("error", e.getMessage());
            return ResponseEntity.badRequest().body(response);
        }
    }

    @GetMapping("/strategy")
    public ResponseEntity<Map<String, String>> getStrategy() {
        Map<String, String> response = new HashMap<>();
        response.put("currentStrategy", loadBalancerService.getCurrentStrategy());
        return ResponseEntity.ok(response);
    }

    private Map<String, Object> buildInstanceInfo(Instance instance) {
        Map<String, Object> info = new HashMap<>();
        info.put("id", instance.getId());
        info.put("host", instance.getHost());
        info.put("port", instance.getPort());
        info.put("url", instance.getUrl());
        info.put("registeredAt", instance.getRegisteredAt());
        info.put("online", instance.isOnline());
        info.put("healthy", instance.isHealthy());
        info.put("activeRequests", instance.getActiveRequests());
        return info;
    }

    public static class RegisterRequest {
        public String id;
        @NotBlank(message = "Host is required")
        public String host;
        @Positive(message = "Port must be positive")
        public int port;
    }

    public static class StrategyRequest {
        @NotBlank(message = "Strategy is required")
        public String strategy;
    }
}
