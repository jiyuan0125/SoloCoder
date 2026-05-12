package com.poolmgr.controller;

import com.poolmgr.model.AcquiredConnection;
import com.poolmgr.model.PoolConfig;
import com.poolmgr.model.PoolDetail;
import com.poolmgr.model.PoolStatus;
import com.poolmgr.model.ReleaseRequest;
import com.poolmgr.pool.PoolManager;
import jakarta.validation.Valid;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@RestController
@RequestMapping("/pools")
public class PoolController {

    private final PoolManager poolManager;

    public PoolController(PoolManager poolManager) {
        this.poolManager = poolManager;
    }

    @PostMapping
    public ResponseEntity<String> registerPool(@Valid @RequestBody PoolConfig config) {
        boolean success = poolManager.registerPool(config);
        if (success) {
            return ResponseEntity.status(HttpStatus.CREATED).body("Pool registered: " + config.getName());
        }
        return ResponseEntity.status(HttpStatus.CONFLICT).body("Pool already exists or initialization failed: " + config.getName());
    }

    @GetMapping("/{name}/acquire")
    public ResponseEntity<?> acquireConnection(@PathVariable String name) {
        try {
            AcquiredConnection conn = poolManager.acquireConnection(name);
            return ResponseEntity.ok(conn);
        } catch (PoolManager.PoolTimeoutException e) {
            return ResponseEntity.status(HttpStatus.SERVICE_UNAVAILABLE).body(e.getMessage());
        } catch (PoolManager.PoolDeletedException e) {
            return ResponseEntity.status(HttpStatus.GONE).body(e.getMessage());
        } catch (PoolManager.PoolNotFoundException e) {
            return ResponseEntity.status(HttpStatus.NOT_FOUND).body(e.getMessage());
        }
    }

    @PostMapping("/{name}/release")
    public ResponseEntity<String> releaseConnection(@PathVariable String name, @Valid @RequestBody ReleaseRequest request) {
        try {
            poolManager.releaseConnection(name, request.getConnectionId());
            return ResponseEntity.ok("Connection released");
        } catch (PoolManager.PoolNotFoundException e) {
            return ResponseEntity.status(HttpStatus.NOT_FOUND).body(e.getMessage());
        }
    }

    @GetMapping
    public ResponseEntity<List<PoolStatus>> getAllPools() {
        List<PoolStatus> statuses = poolManager.getAllStatuses();
        return ResponseEntity.ok(statuses);
    }

    @GetMapping("/{name}")
    public ResponseEntity<?> getPoolDetail(@PathVariable String name) {
        try {
            PoolDetail detail = poolManager.getPoolDetail(name);
            return ResponseEntity.ok(detail);
        } catch (PoolManager.PoolNotFoundException e) {
            return ResponseEntity.status(HttpStatus.NOT_FOUND).body(e.getMessage());
        }
    }

    @DeleteMapping("/{name}")
    public ResponseEntity<String> deletePool(@PathVariable String name) {
        try {
            boolean success = poolManager.deletePool(name);
            if (success) {
                return ResponseEntity.ok("Pool deleted: " + name);
            }
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body("Failed to delete pool: " + name);
        } catch (PoolManager.PoolNotFoundException e) {
            return ResponseEntity.status(HttpStatus.NOT_FOUND).body(e.getMessage());
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body("Interrupted while deleting pool: " + name);
        }
    }
}
