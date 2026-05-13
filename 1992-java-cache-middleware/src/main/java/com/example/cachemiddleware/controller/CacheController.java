package com.example.cachemiddleware.controller;

import com.example.cachemiddleware.cache.CacheManager;
import com.example.cachemiddleware.model.*;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.util.HashMap;
import java.util.Map;

@RestController
@RequestMapping("/api")
public class CacheController {

    private final CacheManager cacheManager;

    public CacheController(CacheManager cacheManager) {
        this.cacheManager = cacheManager;
    }

    @GetMapping("/namespaces/{namespace}/keys/{key}")
    public ResponseEntity<Object> get(@PathVariable String namespace, @PathVariable String key) {
        Object value = cacheManager.get(namespace, key);
        if (value == null) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(value);
    }

    @PostMapping("/namespaces/{namespace}/keys/{key}")
    public ResponseEntity<Map<String, Object>> put(
            @PathVariable String namespace,
            @PathVariable String key,
            @RequestBody(required = false) Object value,
            @RequestParam(required = false) Long ttlSeconds) {
        boolean success = cacheManager.put(namespace, key, value, ttlSeconds);
        Map<String, Object> result = new HashMap<>();
        result.put("success", success);
        if (!success) {
            result.put("error", "Key already exists or value size exceeds limit");
            return ResponseEntity.badRequest().body(result);
        }
        return ResponseEntity.ok(result);
    }

    @PutMapping("/namespaces/{namespace}/keys/{key}")
    public ResponseEntity<Map<String, Object>> update(
            @PathVariable String namespace,
            @PathVariable String key,
            @RequestBody(required = false) Object value,
            @RequestParam(required = false) Long ttlSeconds) {
        boolean success = cacheManager.update(namespace, key, value, ttlSeconds);
        Map<String, Object> result = new HashMap<>();
        result.put("success", success);
        if (!success) {
            result.put("error", "Key not found or value size exceeds limit");
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(result);
    }

    @DeleteMapping("/namespaces/{namespace}/keys/{key}")
    public ResponseEntity<Map<String, Object>> delete(
            @PathVariable String namespace,
            @PathVariable String key) {
        boolean success = cacheManager.delete(namespace, key);
        Map<String, Object> result = new HashMap<>();
        result.put("success", success);
        if (!success) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(result);
    }

    @DeleteMapping("/namespaces/{namespace}/keys")
    public ResponseEntity<Map<String, Object>> clearNamespace(@PathVariable String namespace) {
        cacheManager.clearNamespace(namespace);
        Map<String, Object> result = new HashMap<>();
        result.put("success", true);
        return ResponseEntity.ok(result);
    }

    @PostMapping("/namespaces/{namespace}/batch")
    public ResponseEntity<Map<String, Object>> batchPut(
            @PathVariable String namespace,
            @Valid @RequestBody BatchWriteRequest request) {
        CacheManager.BatchWriteResult result = cacheManager.batchPut(namespace, request.getEntries());
        Map<String, Object> response = new HashMap<>();
        response.put("success", result.success);
        if (!result.success) {
            response.put("error", result.errorMessage);
            return ResponseEntity.badRequest().body(response);
        }
        return ResponseEntity.ok(response);
    }

    @GetMapping("/namespaces/{namespace}/config")
    public ResponseEntity<NamespaceConfig> getConfig(@PathVariable String namespace) {
        NamespaceConfig config = cacheManager.getConfig(namespace);
        return ResponseEntity.ok(config);
    }

    @PutMapping("/namespaces/{namespace}/config")
    public ResponseEntity<NamespaceConfig> updateConfig(
            @PathVariable String namespace,
            @RequestBody NamespaceConfig config) {
        cacheManager.updateConfig(namespace, config);
        NamespaceConfig updated = cacheManager.getConfig(namespace);
        return ResponseEntity.ok(updated);
    }

    @GetMapping("/namespaces/{namespace}/stats")
    public ResponseEntity<NamespaceStats> getStats(@PathVariable String namespace) {
        NamespaceStats stats = cacheManager.getStats(namespace);
        return ResponseEntity.ok(stats);
    }
}
