package com.poolguard.controller;

import com.poolguard.model.PoolConfig;
import com.poolguard.model.PoolDetail;
import com.poolguard.model.PoolOverview;
import com.poolguard.service.PoolManagerService;
import jakarta.validation.Valid;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/pools")
public class PoolController {
    private static final Logger logger = LoggerFactory.getLogger(PoolController.class);
    private final PoolManagerService poolManagerService;

    public PoolController(PoolManagerService poolManagerService) {
        this.poolManagerService = poolManagerService;
    }

    @PostMapping
    public ResponseEntity<?> registerPool(@Valid @RequestBody PoolConfig config) {
        try {
            String id = poolManagerService.registerPool(config);
            return ResponseEntity.status(HttpStatus.CREATED).body(Map.of("id", id));
        } catch (IllegalArgumentException e) {
            logger.warn("Failed to register pool: {}", e.getMessage());
            return ResponseEntity.status(HttpStatus.CONFLICT).body(Map.of("error", e.getMessage()));
        }
    }

    @GetMapping
    public ResponseEntity<List<PoolOverview>> getAllPools() {
        List<PoolOverview> overviews = poolManagerService.getAllPoolOverviews();
        return ResponseEntity.ok(overviews);
    }

    @GetMapping("/{id}")
    public ResponseEntity<?> getPool(@PathVariable String id) {
        PoolDetail detail = poolManagerService.getPoolDetail(id);
        if (detail == null) {
            return ResponseEntity.status(HttpStatus.NOT_FOUND).body(Map.of("error", "Pool not found"));
        }
        return ResponseEntity.ok(detail);
    }

    @PostMapping("/{id}/reset")
    public ResponseEntity<?> resetPool(@PathVariable String id) {
        boolean success = poolManagerService.resetPool(id);
        if (!success) {
            return ResponseEntity.status(HttpStatus.NOT_FOUND).body(Map.of("error", "Pool not found"));
        }
        return ResponseEntity.ok(Map.of("message", "Pool reset successfully"));
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<?> deletePool(@PathVariable String id) {
        try {
            boolean success = poolManagerService.removePool(id);
            if (!success) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND).body(Map.of("error", "Pool not found"));
            }
            return ResponseEntity.ok(Map.of("message", "Pool removed successfully"));
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body(Map.of("error", "Pool removal interrupted"));
        }
    }
}
