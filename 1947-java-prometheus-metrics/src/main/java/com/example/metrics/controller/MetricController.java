package com.example.metrics.controller;

import com.example.metrics.dto.MetricRequest;
import com.example.metrics.dto.QueryResponse;
import com.example.metrics.service.MetricService;
import jakarta.validation.Valid;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.time.Instant;
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

    @GetMapping(produces = MediaType.TEXT_PLAIN_VALUE)
    public ResponseEntity<String> getMetrics() {
        String output = metricService.exportPrometheusFormat();
        return ResponseEntity.ok(output);
    }

    @PostMapping("/{type}")
    public ResponseEntity<Map<String, String>> recordMetric(
            @PathVariable String type,
            @Valid @RequestBody MetricRequest request) {
        
        Map<String, String> labels = request.getLabels() != null ? request.getLabels() : new HashMap<>();
        double value = request.getValue();
        String typeLower = type.toLowerCase();

        switch (typeLower) {
            case "counter":
                metricService.recordCounter(request.getMetricName(), labels, value);
                break;
            case "gauge":
                metricService.recordGauge(request.getMetricName(), labels, value);
                break;
            case "histogram":
                metricService.recordHistogram(request.getMetricName(), labels, value);
                break;
            default:
                throw new IllegalArgumentException("Invalid metric type: " + type + ". Supported types: counter, gauge, histogram");
        }

        Map<String, String> response = new HashMap<>();
        response.put("status", "success");
        response.put("type", typeLower);
        response.put("metricName", request.getMetricName());
        return ResponseEntity.ok(response);
    }

    @GetMapping("/query")
    public ResponseEntity<List<QueryResponse>> queryMetrics(
            @RequestParam String metricName,
            @RequestParam String type,
            @RequestParam(required = false) String startTime,
            @RequestParam(required = false) String endTime,
            @RequestParam(required = false) Map<String, String> allParams) {
        
        Map<String, String> labels = new HashMap<>();
        for (Map.Entry<String, String> entry : allParams.entrySet()) {
            String key = entry.getKey();
            if (!"metricName".equals(key) && !"type".equals(key) && 
                !"startTime".equals(key) && !"endTime".equals(key) &&
                key.startsWith("label_")) {
                labels.put(key.substring(6), entry.getValue());
            }
        }

        Instant start = startTime != null ? Instant.parse(startTime) : null;
        Instant end = endTime != null ? Instant.parse(endTime) : null;

        List<QueryResponse> results = metricService.queryMetrics(metricName, type, 
            labels.isEmpty() ? null : labels, start, end);

        return ResponseEntity.ok(results);
    }
}
