package com.depgraph.controller;

import com.depgraph.model.CycleAlert;
import com.depgraph.model.GraphData;
import com.depgraph.model.ImpactResult;
import com.depgraph.model.ServiceNode;
import com.depgraph.model.ServiceRegistration;
import com.depgraph.service.DependencyGraphService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/")
@RequiredArgsConstructor
public class DependencyGraphController {

    private final DependencyGraphService graphService;

    @PostMapping("/services")
    public ResponseEntity<Map<String, Object>> registerService(@Valid @RequestBody ServiceRegistration registration) {
        CycleAlert cycleAlert = graphService.registerService(registration.getServiceName(), registration.getDependencies());

        return ResponseEntity.ok(Map.of(
                "message", "Service registered successfully",
                "serviceName", registration.getServiceName(),
                "dependencies", registration.getDependencies(),
                "cycleCheck", cycleAlert
        ));
    }

    @GetMapping("/services")
    public ResponseEntity<List<ServiceNode>> listServices() {
        return ResponseEntity.ok(graphService.getAllServices());
    }

    @GetMapping("/services/{serviceName}")
    public ResponseEntity<ServiceNode> getService(@PathVariable String serviceName) {
        ServiceNode node = graphService.getService(serviceName);
        if (node == null) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(node);
    }

    @DeleteMapping("/services/{serviceName}")
    public ResponseEntity<Map<String, Object>> unregisterService(@PathVariable String serviceName) {
        boolean removed = graphService.unregisterService(serviceName);
        if (!removed) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(Map.of(
                "message", "Service unregistered successfully",
                "serviceName", serviceName
        ));
    }

    @GetMapping("/impact/{serviceName}")
    public ResponseEntity<ImpactResult> analyzeImpact(@PathVariable String serviceName) {
        ImpactResult result = graphService.analyzeImpact(serviceName);
        return ResponseEntity.ok(result);
    }

    @GetMapping("/graph")
    public ResponseEntity<GraphData> getGraph() {
        return ResponseEntity.ok(graphService.getGraphData());
    }

    @PostMapping("/services/{serviceName}/heartbeat")
    public ResponseEntity<Map<String, Object>> heartbeat(@PathVariable String serviceName) {
        boolean updated = graphService.heartbeat(serviceName);
        if (!updated) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(Map.of(
                "message", "Heartbeat recorded",
                "serviceName", serviceName,
                "status", ServiceNode.HealthStatus.HEALTHY
        ));
    }

    @PutMapping("/services/{serviceName}/health")
    public ResponseEntity<Map<String, Object>> updateHealth(@PathVariable String serviceName,
                                                             @RequestBody Map<String, String> body) {
        String statusStr = body.get("status");
        if (statusStr == null) {
            return ResponseEntity.badRequest().body(Map.of("error", "status is required"));
        }

        ServiceNode.HealthStatus status;
        try {
            status = ServiceNode.HealthStatus.valueOf(statusStr.toUpperCase());
        } catch (IllegalArgumentException e) {
            return ResponseEntity.badRequest().body(Map.of(
                    "error", "Invalid status. Valid values: HEALTHY, UNHEALTHY, UNKNOWN"
            ));
        }

        boolean updated = graphService.updateHealth(serviceName, status);
        if (!updated) {
            return ResponseEntity.notFound().build();
        }

        return ResponseEntity.ok(Map.of(
                "message", "Health status updated",
                "serviceName", serviceName,
                "status", status
        ));
    }
}
