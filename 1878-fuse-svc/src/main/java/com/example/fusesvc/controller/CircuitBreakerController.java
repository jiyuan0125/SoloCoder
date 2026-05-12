package com.example.fusesvc.controller;

import com.example.fusesvc.dto.CircuitBreakerConfig;
import com.example.fusesvc.dto.CircuitBreakerOverview;
import com.example.fusesvc.dto.CircuitBreakerStatus;
import com.example.fusesvc.dto.ReportRequest;
import com.example.fusesvc.service.CircuitBreaker;
import com.example.fusesvc.service.CircuitBreakerManager;
import jakarta.validation.Valid;
import java.util.List;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/breakers")
public class CircuitBreakerController {
    
    private final CircuitBreakerManager breakerManager;

    public CircuitBreakerController(CircuitBreakerManager breakerManager) {
        this.breakerManager = breakerManager;
    }

    @GetMapping
    public List<CircuitBreakerOverview> getAllBreakers() {
        return breakerManager.getAllOverviews();
    }

    @GetMapping("/{service}")
    public CircuitBreakerStatus getBreakerStatus(@PathVariable String service) {
        return breakerManager.getStatus(service);
    }

    @GetMapping("/{service}/allow")
    public ResponseEntity<Void> allowRequest(@PathVariable String service) {
        CircuitBreaker breaker = breakerManager.getOrCreate(service);
        if (breaker.allowRequest()) {
            return ResponseEntity.ok().build();
        } else {
            return ResponseEntity.status(HttpStatus.SERVICE_UNAVAILABLE).build();
        }
    }

    @PostMapping("/{service}/report")
    public ResponseEntity<Void> report(
            @PathVariable String service,
            @RequestBody ReportRequest request) {
        breakerManager.report(service, request.isSuccess());
        return ResponseEntity.ok().build();
    }

    @PutMapping("/{service}/config")
    public ResponseEntity<Void> updateConfig(
            @PathVariable String service,
            @Valid @RequestBody CircuitBreakerConfig config) {
        breakerManager.updateConfig(
            service, 
            config.getFailureThreshold(), 
            config.getOpenDurationSeconds()
        );
        return ResponseEntity.ok().build();
    }
}
