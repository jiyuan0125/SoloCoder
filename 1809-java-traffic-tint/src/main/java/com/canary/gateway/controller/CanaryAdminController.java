package com.canary.gateway.controller;

import com.canary.gateway.config.CanaryConfig;
import com.canary.gateway.service.CanaryService;
import com.canary.gateway.service.HealthCheckService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.LinkedHashMap;
import java.util.Map;

@RestController
@RequestMapping("/api/canary")
public class CanaryAdminController {

    @Autowired
    private CanaryConfig canaryConfig;

    @Autowired
    private CanaryService canaryService;

    @Autowired
    private HealthCheckService healthCheckService;

    @PostMapping("/config/ratio")
    public ResponseEntity<?> setRatio(@RequestBody Map<String, Object> request) {
        try {
            int ratio = Integer.parseInt(request.get("ratio").toString());
            String operator = request.get("operator") != null ? request.get("operator").toString() : "system";

            if (ratio < 0 || ratio > 100) {
                return ResponseEntity.badRequest().body("Ratio must be between 0 and 100");
            }

            canaryService.setRatio(ratio, operator);
            return ResponseEntity.ok().build();
        } catch (NumberFormatException e) {
            return ResponseEntity.badRequest().body("Invalid ratio format");
        } catch (IllegalArgumentException e) {
            return ResponseEntity.badRequest().body(e.getMessage());
        }
    }

    @PostMapping("/config/backends")
    public ResponseEntity<?> setBackends(@RequestBody Map<String, String> request) {
        String grayBackend = request.get("grayBackend");
        String prodBackend = request.get("prodBackend");

        if (grayBackend != null && !grayBackend.isEmpty()) {
            canaryService.setGrayBackend(grayBackend);
        }
        if (prodBackend != null && !prodBackend.isEmpty()) {
            canaryService.setProdBackend(prodBackend);
        }

        return ResponseEntity.ok().build();
    }

    @GetMapping("/config")
    public ResponseEntity<Map<String, Object>> getConfig() {
        Map<String, Object> response = new LinkedHashMap<>();
        response.put("ratio", canaryConfig.getRatio());
        response.put("grayBackend", canaryConfig.getGrayBackend());
        response.put("prodBackend", canaryConfig.getProdBackend());
        response.put("lastOperator", canaryService.getLastOperator());
        response.put("lastRatio", canaryService.getLastRatio());
        response.put("lastModified", canaryService.getLastModified().format(DateTimeFormatter.ISO_LOCAL_DATE_TIME));
        return ResponseEntity.ok(response);
    }

    @GetMapping("/status")
    public ResponseEntity<Map<String, Object>> getStatus() {
        Map<String, Object> response = new LinkedHashMap<>();
        response.put("grayBackendHealthy", healthCheckService.isHealthy());
        response.put("lastHealthStatusChange", healthCheckService.getLastStatusChange().format(DateTimeFormatter.ISO_LOCAL_DATE_TIME));

        Map<String, Object> stats = new LinkedHashMap<>();
        stats.put("totalGrayRequests", canaryService.getTotalGrayRequests());
        stats.put("totalProdRequests", canaryService.getTotalProdRequests());
        stats.put("grayPerMinute", canaryService.getGrayStats());
        stats.put("prodPerMinute", canaryService.getProdStats());
        response.put("statistics", stats);

        Map<String, Object> config = new LinkedHashMap<>();
        config.put("ratio", canaryConfig.getRatio());
        config.put("grayBackend", canaryConfig.getGrayBackend());
        config.put("prodBackend", canaryConfig.getProdBackend());
        config.put("lastOperator", canaryService.getLastOperator());
        config.put("lastRatio", canaryService.getLastRatio());
        config.put("lastModified", canaryService.getLastModified().format(DateTimeFormatter.ISO_LOCAL_DATE_TIME));
        response.put("config", config);

        return ResponseEntity.ok(response);
    }
}
