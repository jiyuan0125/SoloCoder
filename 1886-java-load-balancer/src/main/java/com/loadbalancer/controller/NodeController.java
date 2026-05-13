package com.loadbalancer.controller;

import com.loadbalancer.dto.HealthConfigRequest;
import com.loadbalancer.dto.RegisterNodeRequest;
import com.loadbalancer.dto.UpdateWeightRequest;
import com.loadbalancer.model.Node;
import com.loadbalancer.service.NodeRegistryService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

@RestController
@RequestMapping("/nodes")
@RequiredArgsConstructor
public class NodeController {
    private final NodeRegistryService nodeRegistryService;

    @PostMapping
    public ResponseEntity<?> registerNode(@Valid @RequestBody RegisterNodeRequest request) {
        Node node = nodeRegistryService.registerNode(request.getAddress(), request.getWeight());
        if (node == null) {
            return ResponseEntity.status(HttpStatus.CONFLICT)
                    .body(Map.of("error", "Address is already registered"));
        }
        return ResponseEntity.status(HttpStatus.CREATED).body(node);
    }

    @PutMapping("/{id}/weight")
    public ResponseEntity<?> updateWeight(
            @PathVariable String id,
            @Valid @RequestBody UpdateWeightRequest request) {
        boolean updated = nodeRegistryService.updateWeight(id, request.getWeight());
        if (!updated) {
            return ResponseEntity.notFound().build();
        }
        Node node = nodeRegistryService.getNode(id);
        return ResponseEntity.ok(Map.of(
                "id", node.getId(),
                "weight", node.getWeight()
        ));
    }

    @PostMapping("/{id}/health-config")
    public ResponseEntity<?> updateHealthConfig(
            @PathVariable String id,
            @Valid @RequestBody HealthConfigRequest request) {
        if (request.getIntervalSeconds() == null && request.getTimeoutSeconds() == null) {
            return ResponseEntity.badRequest().body(
                Map.of("error", "At least one of intervalSeconds or timeoutSeconds must be provided")
            );
        }
        boolean updated = nodeRegistryService.updateHealthConfig(
                id, request.getIntervalSeconds(), request.getTimeoutSeconds());
        if (!updated) {
            return ResponseEntity.notFound().build();
        }
        Node node = nodeRegistryService.getNode(id);
        return ResponseEntity.ok(Map.of(
                "id", node.getId(),
                "healthCheckIntervalSeconds", node.getHealthCheckIntervalSeconds(),
                "healthCheckTimeoutSeconds", node.getHealthCheckTimeoutSeconds()
        ));
    }

    @PutMapping("/{id}/offline")
    public ResponseEntity<?> takeOffline(@PathVariable String id) {
        boolean updated = nodeRegistryService.takeNodeOffline(id);
        if (!updated) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok().build();
    }

    @PutMapping("/{id}/online")
    public ResponseEntity<?> bringOnline(@PathVariable String id) {
        boolean updated = nodeRegistryService.bringNodeOnline(id);
        if (!updated) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok().build();
    }
}
