package com.servicemesh.controlplane.controller;

import com.servicemesh.common.model.TopologyData;
import com.servicemesh.controlplane.service.TopologyService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

@RestController
@RequestMapping("/api/topology")
public class TopologyController {

    private final TopologyService topologyService;

    public TopologyController(TopologyService topologyService) {
        this.topologyService = topologyService;
    }

    @GetMapping
    public ResponseEntity<TopologyData> getTopology() {
        return ResponseEntity.ok(topologyService.getTopology());
    }

    @PostMapping("/telemetry")
    public ResponseEntity<Map<String, String>> recordTelemetry(@RequestBody Map<String, Object> payload) {
        try {
            String source = (String) payload.get("source");
            String target = (String) payload.get("target");
            double qps = ((Number) payload.getOrDefault("qps", 0)).doubleValue();
            long totalRequests = ((Number) payload.getOrDefault("totalRequests", 0)).longValue();
            long errorCount = ((Number) payload.getOrDefault("errorCount", 0)).longValue();
            double errorRate = ((Number) payload.getOrDefault("errorRate", 0)).doubleValue();
            double avgLatencyMs = ((Number) payload.getOrDefault("avgLatencyMs", 0)).doubleValue();

            topologyService.recordTelemetry(source, target, qps, totalRequests, errorCount, errorRate, avgLatencyMs);
            return ResponseEntity.ok(Map.of("status", "ok"));
        } catch (Exception e) {
            return ResponseEntity.status(400).body(Map.of("status", "error", "message", e.getMessage()));
        }
    }
}
