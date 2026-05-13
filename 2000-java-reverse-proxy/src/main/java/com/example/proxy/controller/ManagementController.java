package com.example.proxy.controller;

import com.example.proxy.config.ProxyProperties;
import com.example.proxy.dto.AccessLogStats;
import com.example.proxy.dto.BackendStats;
import com.example.proxy.model.AccessLog;
import com.example.proxy.model.Backend;
import com.example.proxy.model.Route;
import com.example.proxy.service.AccessLogService;
import com.example.proxy.service.LoadBalanceService;
import com.example.proxy.service.RouteService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.*;
import java.util.stream.Collectors;

@Slf4j
@RestController
@RequestMapping("${proxy.management.base-path:/admin}")
public class ManagementController {
    
    private final RouteService routeService;
    private final AccessLogService accessLogService;
    private final LoadBalanceService loadBalanceService;
    private final ProxyProperties properties;
    
    public ManagementController(RouteService routeService, AccessLogService accessLogService,
                                LoadBalanceService loadBalanceService, ProxyProperties properties) {
        this.routeService = routeService;
        this.accessLogService = accessLogService;
        this.loadBalanceService = loadBalanceService;
        this.properties = properties;
    }
    
    @GetMapping("/health")
    public ResponseEntity<Map<String, Object>> health() {
        Map<String, Object> health = new HashMap<>();
        health.put("status", "UP");
        health.put("timestamp", new Date());
        return ResponseEntity.ok(health);
    }
    
    @GetMapping("/backends")
    public ResponseEntity<List<BackendStats>> getAllBackends() {
        List<Route> routes = routeService.getAllRoutes();
        Map<String, BackendStats> backendStatsMap = new HashMap<>();
        
        for (Route route : routes) {
            for (Backend backend : route.getBackends()) {
                BackendStats stats = BackendStats.builder()
                        .backendUrl(backend.getUrl())
                        .status(backend.getStatus().name())
                        .weight(backend.getWeight())
                        .build();
                backendStatsMap.put(backend.getUrl(), stats);
            }
        }
        
        return ResponseEntity.ok(new ArrayList<>(backendStatsMap.values()));
    }
    
    @GetMapping("/backends/{routePath}")
    public ResponseEntity<List<BackendStats>> getBackendsByRoute(
            @PathVariable String routePath,
            @RequestParam(defaultValue = "prefix") String type) {
        
        Optional<Route> routeOpt = routeService.getRoute("/" + routePath, 
                Route.RouteType.valueOf(type.toUpperCase()));
        
        if (routeOpt.isEmpty()) {
            return ResponseEntity.notFound().build();
        }
        
        List<BackendStats> stats = routeOpt.get().getBackends().stream()
                .map(backend -> BackendStats.builder()
                        .backendUrl(backend.getUrl())
                        .status(backend.getStatus().name())
                        .weight(backend.getWeight())
                        .build())
                .collect(Collectors.toList());
        
        return ResponseEntity.ok(stats);
    }
    
    @GetMapping("/routes")
    public ResponseEntity<List<Map<String, Object>>> getAllRoutes() {
        List<Map<String, Object>> routesInfo = routeService.getAllRoutes().stream()
                .map(route -> {
                    Map<String, Object> info = new HashMap<>();
                    info.put("path", route.getPath());
                    info.put("type", route.getType().name());
                    info.put("backendCount", route.getBackends().size());
                    info.put("healthyCount", route.getBackends().stream()
                            .filter(Backend::isAvailable).count());
                    return info;
                })
                .collect(Collectors.toList());
        
        return ResponseEntity.ok(routesInfo);
    }
    
    @PostMapping("/routes")
    public ResponseEntity<Map<String, String>> addRoute(@RequestBody Map<String, Object> routeData) {
        try {
            String path = (String) routeData.get("path");
            String typeStr = (String) routeData.getOrDefault("type", "prefix");
            Route.RouteType type = Route.RouteType.valueOf(typeStr.toUpperCase());
            
            List<Map<String, Object>> backendsData = (List<Map<String, Object>>) routeData.get("backends");
            List<Backend> backends = new ArrayList<>();
            
            if (backendsData != null) {
                for (Map<String, Object> backendData : backendsData) {
                    Backend backend = Backend.builder()
                            .id(UUID.randomUUID().toString())
                            .url((String) backendData.get("url"))
                            .weight((Integer) backendData.getOrDefault("weight", 1))
                            .status(Backend.BackendStatus.HEALTHY)
                            .build();
                    backends.add(backend);
                }
            }
            
            Route route = Route.builder()
                    .id(UUID.randomUUID().toString())
                    .path(path)
                    .type(type)
                    .backends(backends)
                    .build();
            
            routeService.addRoute(route);
            
            Map<String, String> response = new HashMap<>();
            response.put("status", "success");
            response.put("message", "Route added successfully");
            
            return ResponseEntity.ok(response);
            
        } catch (Exception e) {
            log.error("Error adding route: {}", e.getMessage(), e);
            Map<String, String> response = new HashMap<>();
            response.put("status", "error");
            response.put("message", e.getMessage());
            return ResponseEntity.badRequest().body(response);
        }
    }
    
    @DeleteMapping("/routes")
    public ResponseEntity<Map<String, String>> removeRoute(
            @RequestParam String path,
            @RequestParam(defaultValue = "prefix") String type) {
        
        boolean removed = routeService.removeRoute(path, Route.RouteType.valueOf(type.toUpperCase()));
        
        Map<String, String> response = new HashMap<>();
        if (removed) {
            response.put("status", "success");
            response.put("message", "Route removed successfully");
            return ResponseEntity.ok(response);
        } else {
            response.put("status", "error");
            response.put("message", "Route not found");
            return ResponseEntity.notFound().build();
        }
    }
    
    @GetMapping("/logs/stats")
    public ResponseEntity<AccessLogStats> getAccessLogStats() {
        return ResponseEntity.ok(accessLogService.getStats());
    }
    
    @GetMapping("/logs")
    public ResponseEntity<List<AccessLog>> getRecentLogs(
            @RequestParam(defaultValue = "100") int limit) {
        return ResponseEntity.ok(accessLogService.getRecentLogs(limit));
    }
    
    @DeleteMapping("/logs")
    public ResponseEntity<Map<String, String>> clearLogs() {
        accessLogService.clearStats();
        Map<String, String> response = new HashMap<>();
        response.put("status", "success");
        response.put("message", "Logs cleared successfully");
        return ResponseEntity.ok(response);
    }
    
    @GetMapping("/strategies")
    public ResponseEntity<List<String>> getAvailableStrategies() {
        return ResponseEntity.ok(loadBalanceService.getAvailableStrategies());
    }
    
    @GetMapping("/config")
    public ResponseEntity<Map<String, Object>> getConfig() {
        Map<String, Object> config = new HashMap<>();
        config.put("requestTimeout", properties.getRequestTimeout());
        config.put("connectTimeout", properties.getConnectTimeout());
        config.put("loadBalanceStrategy", properties.getLoadBalanceStrategy());
        config.put("failureRecoveryTime", properties.getFailureRecoveryTime());
        config.put("consecutive5xxThreshold", properties.getConsecutive5xxThreshold());
        config.put("healthCheck", properties.getHealthCheck());
        config.put("accessLog", properties.getAccessLog());
        return ResponseEntity.ok(config);
    }
}
