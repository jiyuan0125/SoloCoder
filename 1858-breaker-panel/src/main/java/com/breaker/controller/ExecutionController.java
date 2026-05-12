package com.breaker.controller;

import com.breaker.core.BreakerManager;
import com.breaker.core.CircuitBreaker;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Optional;

@RestController
@RequestMapping("/breakers/{service}/execute")
public class ExecutionController {

    private final BreakerManager breakerManager;

    public ExecutionController(BreakerManager breakerManager) {
        this.breakerManager = breakerManager;
    }

    @PostMapping
    public ResponseEntity<?> execute(@PathVariable String service, @RequestParam(required = false) Boolean simulateFailure) {
        Optional<CircuitBreaker> breakerOpt = breakerManager.getBreaker(service);
        if (!breakerOpt.isPresent()) {
            return ResponseEntity.notFound().build();
        }

        CircuitBreaker breaker = breakerOpt.get();
        if (!breaker.allowRequest()) {
            return ResponseEntity.status(HttpStatus.SERVICE_UNAVAILABLE).body("Circuit breaker is OPEN");
        }

        boolean success = simulateFailure == null || !simulateFailure;
        if (success) {
            breaker.recordSuccess();
            return ResponseEntity.ok("Request succeeded");
        } else {
            breaker.recordFailure();
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body("Request failed");
        }
    }

    @PostMapping("/success")
    public ResponseEntity<?> recordSuccess(@PathVariable String service) {
        Optional<CircuitBreaker> breakerOpt = breakerManager.getBreaker(service);
        if (!breakerOpt.isPresent()) {
            return ResponseEntity.notFound().build();
        }
        breakerOpt.get().recordSuccess();
        return ResponseEntity.ok().build();
    }

    @PostMapping("/failure")
    public ResponseEntity<?> recordFailure(@PathVariable String service) {
        Optional<CircuitBreaker> breakerOpt = breakerManager.getBreaker(service);
        if (!breakerOpt.isPresent()) {
            return ResponseEntity.notFound().build();
        }
        breakerOpt.get().recordFailure();
        return ResponseEntity.ok().build();
    }
}
