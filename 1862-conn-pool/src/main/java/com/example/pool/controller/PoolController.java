package com.example.pool.controller;

import com.example.pool.model.PoolConfig;
import com.example.pool.model.PoolDetail;
import com.example.pool.model.PoolInfo;
import com.example.pool.service.PoolManagerService;
import jakarta.validation.Valid;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;

@RestController
@RequestMapping("/pools")
public class PoolController {

    private final PoolManagerService poolManagerService;
    private final Map<String, Map.Entry<String, Object>> acquiredConnections = new ConcurrentHashMap<>();

    @Autowired
    public PoolController(PoolManagerService poolManagerService) {
        this.poolManagerService = poolManagerService;
    }

    @PostMapping
    public ResponseEntity<?> registerPool(@Valid @RequestBody PoolConfig config) {
        try {
            poolManagerService.registerPool(config);
            Map<String, Object> response = new HashMap<>();
            response.put("name", config.getName());
            response.put("message", "Pool registered and initialized successfully");
            return ResponseEntity.status(HttpStatus.CREATED).body(response);
        } catch (IllegalArgumentException e) {
            return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to initialize pool: " + e.getMessage()));
        }
    }

    @GetMapping
    public ResponseEntity<List<PoolInfo>> listPools() {
        return ResponseEntity.ok(poolManagerService.listPools());
    }

    @GetMapping("/{name}")
    public ResponseEntity<?> getPoolDetail(@PathVariable String name) {
        try {
            PoolDetail detail = poolManagerService.getPoolDetail(name);
            return ResponseEntity.ok(detail);
        } catch (IllegalArgumentException e) {
            return ResponseEntity.status(HttpStatus.NOT_FOUND).body(Map.of("error", e.getMessage()));
        }
    }

    @DeleteMapping("/{name}")
    public ResponseEntity<?> deletePool(@PathVariable String name) {
        try {
            poolManagerService.deletePool(name);
            return ResponseEntity.ok(Map.of("message", "Pool deleted successfully"));
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to delete pool: " + e.getMessage()));
        }
    }

    @GetMapping("/{name}/acquire")
    public ResponseEntity<?> acquireConnection(@PathVariable String name) {
        try {
            Object connection = poolManagerService.acquireConnection(name);
            String connectionId = UUID.randomUUID().toString();
            acquiredConnections.put(connectionId, Map.entry(name, connection));
            
            Map<String, Object> response = new HashMap<>();
            response.put("connection_id", connectionId);
            response.put("message", "Connection acquired successfully");
            return ResponseEntity.ok(response);
        } catch (IllegalArgumentException e) {
            return ResponseEntity.status(HttpStatus.NOT_FOUND).body(Map.of("error", e.getMessage()));
        } catch (IllegalStateException e) {
            return ResponseEntity.status(HttpStatus.SERVICE_UNAVAILABLE)
                .body(Map.of("error", e.getMessage()));
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.SERVICE_UNAVAILABLE)
                .body(Map.of("error", "Failed to acquire connection: " + e.getMessage()));
        }
    }

    @PostMapping("/{name}/release")
    public ResponseEntity<?> releaseConnection(
        @PathVariable String name,
        @RequestBody Map<String, String> body
    ) {
        String connectionId = body.get("connection_id");
        if (connectionId == null || connectionId.isEmpty()) {
            return ResponseEntity.badRequest().body(Map.of("error", "connection_id is required"));
        }

        Map.Entry<String, Object> entry = acquiredConnections.remove(connectionId);
        if (entry == null) {
            return ResponseEntity.status(HttpStatus.NOT_FOUND)
                .body(Map.of("error", "Connection not found: " + connectionId));
        }

        if (!entry.getKey().equals(name)) {
            acquiredConnections.put(connectionId, entry);
            return ResponseEntity.badRequest()
                .body(Map.of("error", "Connection does not belong to pool: " + name));
        }

        try {
            poolManagerService.releaseConnection(name, entry.getValue());
            return ResponseEntity.ok(Map.of("message", "Connection released successfully"));
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to release connection: " + e.getMessage()));
        }
    }
}
