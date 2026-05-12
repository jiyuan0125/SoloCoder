package com.example.connectionpool.controller;

import com.example.connectionpool.dto.UpdatePoolConfigRequest;
import com.example.connectionpool.model.PoolStatus;
import com.example.connectionpool.pool.ConnectionPoolManager;
import com.example.connectionpool.pool.HttpConnectionPool;
import com.example.connectionpool.service.ConnectionService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

@RestController
@RequestMapping("/api")
@RequiredArgsConstructor
public class ConnectionController {
    private final ConnectionService connectionService;
    private final ConnectionPoolManager poolManager;

    @PostMapping("/pools/{poolName}/borrow")
    public ResponseEntity<Map<String, String>> borrowConnection(
            @PathVariable String poolName) throws InterruptedException {
        String connectionId = connectionService.borrowConnection(poolName);
        return ResponseEntity.ok(Map.of("connectionId", connectionId, "poolName", poolName));
    }

    @PostMapping("/pools/{poolName}/release/{connectionId}")
    public ResponseEntity<Map<String, String>> releaseConnection(
            @PathVariable String poolName,
            @PathVariable String connectionId) {
        connectionService.releaseConnection(connectionId);
        return ResponseEntity.ok(Map.of("status", "released", "connectionId", connectionId));
    }

    @PutMapping("/pools/{poolName}/config")
    public ResponseEntity<PoolStatus> updatePoolConfig(
            @PathVariable String poolName,
            @RequestBody UpdatePoolConfigRequest request) {
        HttpConnectionPool pool = poolManager.getOrCreatePool(poolName);

        if (request.getMaxConnections() != null) {
            pool.updateMaxConnections(request.getMaxConnections());
        }
        if (request.getWaitTimeoutMs() != null) {
            pool.updateWaitTimeout(request.getWaitTimeoutMs());
        }
        if (request.getLeakThresholdSeconds() != null) {
            pool.updateLeakThresholdSeconds(request.getLeakThresholdSeconds());
        }
        if (request.getMaxQueueSize() != null) {
            pool.updateMaxQueueSize(request.getMaxQueueSize());
        }

        return ResponseEntity.ok(pool.getStatus());
    }

    @GetMapping("/pools")
    public ResponseEntity<Map<String, PoolStatus>> getAllPoolStatuses() {
        return ResponseEntity.ok(poolManager.getAllPoolStatuses());
    }

    @GetMapping("/pools/{poolName}")
    public ResponseEntity<PoolStatus> getPoolStatus(@PathVariable String poolName) {
        if (!poolManager.poolExists(poolName)) {
            return ResponseEntity.notFound().build();
        }
        HttpConnectionPool pool = poolManager.getPool(poolName);
        return ResponseEntity.ok(pool.getStatus());
    }
}
