package com.loadbalancer.controller;

import com.loadbalancer.model.BackendInstance;
import com.loadbalancer.service.InstanceRegistry;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/instances")
public class InstanceController {

    private final InstanceRegistry instanceRegistry;

    public InstanceController(InstanceRegistry instanceRegistry) {
        this.instanceRegistry = instanceRegistry;
    }

    @PostMapping("/register")
    public ResponseEntity<?> register(@RequestBody Map<String, Object> request) {
        String host = (String) request.get("host");
        Integer port = (Integer) request.get("port");
        Integer weight = request.containsKey("weight") ? (Integer) request.get("weight") : 1;

        if (host == null || port == null) {
            return ResponseEntity.badRequest()
                    .body(Map.of("error", "host and port are required"));
        }

        BackendInstance instance = new BackendInstance(host, port, weight);
        instanceRegistry.register(instance);

        return ResponseEntity.ok(Map.of(
                "id", instance.getId(),
                "host", host,
                "port", port,
                "weight", weight,
                "registeredAt", instance.getRegisteredAt().toString()
        ));
    }

    @PostMapping("/deregister")
    public ResponseEntity<?> deregister(@RequestBody Map<String, Object> request) {
        String instanceId = (String) request.get("id");

        if (instanceId == null) {
            return ResponseEntity.badRequest()
                    .body(Map.of("error", "instance id is required"));
        }

        instanceRegistry.deregister(instanceId);
        return ResponseEntity.ok(Map.of("status", "deregistered", "id", instanceId));
    }

    @GetMapping
    public ResponseEntity<List<BackendInstance>> listInstances() {
        return ResponseEntity.ok(instanceRegistry.getAllInstances());
    }

    @GetMapping("/{instanceId}")
    public ResponseEntity<?> getInstance(@PathVariable String instanceId) {
        BackendInstance instance = instanceRegistry.getInstance(instanceId);
        if (instance == null) {
            return ResponseEntity.notFound().build();
        }

        int effectiveWeight = instanceRegistry.getEffectiveWeight(instance);
        return ResponseEntity.ok(Map.of(
                "instance", instance,
                "effectiveWeight", effectiveWeight
        ));
    }
}
