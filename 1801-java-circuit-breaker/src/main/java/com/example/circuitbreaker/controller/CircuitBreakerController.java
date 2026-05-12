package com.example.circuitbreaker.controller;

import com.example.circuitbreaker.core.CircuitBreaker;
import com.example.circuitbreaker.core.CircuitBreakerManager;
import com.example.circuitbreaker.dto.CircuitBreakerConfigRequest;
import com.example.circuitbreaker.dto.CircuitBreakerStatusResponse;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.validation.annotation.Validated;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

@RestController
@RequestMapping("/circuit-breakers")
@RequiredArgsConstructor
public class CircuitBreakerController {
    private final CircuitBreakerManager manager;

    @GetMapping
    public ResponseEntity<List<CircuitBreakerStatusResponse>> getAllStatus() {
        List<CircuitBreakerStatusResponse> result = manager.getAll().entrySet().stream()
                .map(entry -> CircuitBreakerStatusResponse.from(entry.getKey(), entry.getValue()))
                .collect(Collectors.toList());
        return ResponseEntity.ok(result);
    }

    @GetMapping("/{service}")
    public ResponseEntity<CircuitBreakerStatusResponse> getStatus(@PathVariable String service) {
        CircuitBreaker cb = manager.get(service);
        if (cb == null) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(CircuitBreakerStatusResponse.from(service, cb));
    }

    @PutMapping("/{service}/config")
    public ResponseEntity<CircuitBreakerStatusResponse> updateConfig(
            @PathVariable String service,
            @Validated @RequestBody CircuitBreakerConfigRequest request) {
        CircuitBreaker cb = manager.getOrCreate(service);
        cb.updateConfig(
                request.getFailureThreshold(),
                request.getOpenDuration(),
                request.getHalfOpenPermittedCalls()
        );
        return ResponseEntity.ok(CircuitBreakerStatusResponse.from(service, cb));
    }

    @PostMapping("/{service}/force-open")
    public ResponseEntity<CircuitBreakerStatusResponse> forceOpen(@PathVariable String service) {
        CircuitBreaker cb = manager.getOrCreate(service);
        cb.forceOpen();
        return ResponseEntity.ok(CircuitBreakerStatusResponse.from(service, cb));
    }

    @PostMapping("/{service}/force-close")
    public ResponseEntity<CircuitBreakerStatusResponse> forceClose(@PathVariable String service) {
        CircuitBreaker cb = manager.getOrCreate(service);
        cb.forceClosed();
        return ResponseEntity.ok(CircuitBreakerStatusResponse.from(service, cb));
    }

    @ExceptionHandler(Exception.class)
    public ResponseEntity<Map<String, Object>> handleException(Exception e) {
        Map<String, Object> body = new HashMap<>();
        body.put("error", e.getMessage());
        return ResponseEntity.badRequest().body(body);
    }
}
