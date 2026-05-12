package com.example.ratelimiter.controller;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.time.Instant;
import java.util.Map;

@RestController
@RequestMapping("/api")
public class TestController {
    private static final Logger log = LoggerFactory.getLogger(TestController.class);

    @GetMapping("/resource")
    public ResponseEntity<Map<String, Object>> getResource() {
        log.info("Processing GET /api/resource request");
        return ResponseEntity.ok(Map.of(
                "status", "success",
                "message", "Resource retrieved",
                "timestamp", Instant.now().toString()
        ));
    }

    @PostMapping("/resource")
    public ResponseEntity<Map<String, Object>> createResource(@RequestBody Map<String, Object> body) {
        log.info("Processing POST /api/resource request with body: {}", body);
        return ResponseEntity.ok(Map.of(
                "status", "success",
                "message", "Resource created",
                "received", body,
                "timestamp", Instant.now().toString()
        ));
    }

    @GetMapping("/users")
    public ResponseEntity<Map<String, Object>> getUsers() {
        log.info("Processing GET /api/users request");
        return ResponseEntity.ok(Map.of(
                "status", "success",
                "users", java.util.List.of("user1", "user2", "user3"),
                "timestamp", Instant.now().toString()
        ));
    }

    @GetMapping("/orders")
    public ResponseEntity<Map<String, Object>> getOrders() {
        log.info("Processing GET /api/orders request");
        return ResponseEntity.ok(Map.of(
                "status", "success",
                "orders", java.util.List.of("order1", "order2"),
                "timestamp", Instant.now().toString()
        ));
    }
}
