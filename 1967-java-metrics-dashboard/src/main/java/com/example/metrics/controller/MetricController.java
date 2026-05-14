package com.example.metrics.controller;

import com.example.metrics.model.AggregatedData;
import com.example.metrics.model.MetricPoint;
import com.example.metrics.model.MetricRegistration;
import com.example.metrics.model.MetricSummary;
import com.example.metrics.service.MetricService;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/metrics")
public class MetricController {

    private final MetricService metricService;

    public MetricController(MetricService metricService) {
        this.metricService = metricService;
    }

    @GetMapping
    public ResponseEntity<List<MetricSummary>> listMetrics() {
        return ResponseEntity.ok(metricService.listAllMetrics());
    }

    @PostMapping("/register")
    public ResponseEntity<Map<String, Object>> registerMetric(@Valid @RequestBody MetricRegistration registration) {
        metricService.registerMetric(registration);
        Map<String, Object> response = new HashMap<>();
        response.put("success", true);
        response.put("name", registration.getName());
        response.put("message", "Metric registered successfully");
        return ResponseEntity.status(HttpStatus.CREATED).body(response);
    }

    @PostMapping
    public ResponseEntity<Map<String, Object>> recordMetric(@Valid @RequestBody MetricPoint point) {
        metricService.recordMetric(point);
        Map<String, Object> response = new HashMap<>();
        response.put("success", true);
        response.put("metricName", point.getMetricName());
        response.put("timestamp", point.getTimestamp());
        return ResponseEntity.status(HttpStatus.CREATED).body(response);
    }

    @GetMapping("/{name}")
    public ResponseEntity<?> queryMetric(
            @PathVariable String name,
            @RequestParam(required = false) Long startTime,
            @RequestParam(required = false) Long endTime,
            @RequestParam(required = false) String granularity) {

        if (metricService.exists(name)) {
            List<AggregatedData> data = metricService.queryMetrics(name, startTime, endTime, granularity);
            return ResponseEntity.ok(data);
        }

        Map<String, Object> error = new HashMap<>();
        error.put("error", "Metric not found");
        error.put("name", name);
        return ResponseEntity.status(HttpStatus.NOT_FOUND).body(error);
    }

    @GetMapping("/{name}/anomalies")
    public ResponseEntity<?> getAnomalies(@PathVariable String name) {
        if (!metricService.exists(name)) {
            Map<String, Object> error = new HashMap<>();
            error.put("error", "Metric not found");
            error.put("name", name);
            return ResponseEntity.status(HttpStatus.NOT_FOUND).body(error);
        }

        List<AggregatedData> anomalies = metricService.getAnomalies(name);
        return ResponseEntity.ok(anomalies);
    }
}
