package com.example.circuitbreaker.api;

import com.example.circuitbreaker.breaker.CircuitBreaker;
import com.example.circuitbreaker.breaker.CircuitBreakerRegistry;
import com.example.circuitbreaker.config.CircuitBreakerConfig;
import com.example.circuitbreaker.config.ConfigService;
import com.example.circuitbreaker.stats.StatsService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.UUID;

@Slf4j
@RestController
@RequestMapping("/proxy")
@RequiredArgsConstructor
public class ProxyController {

    private final CircuitBreakerRegistry circuitBreakerRegistry;
    private final ConfigService configService;
    private final StatsService statsService;

    @PostMapping("/{service}/call")
    public ResponseEntity<String> callService(
        @PathVariable String service,
        @RequestBody(required = false) CallRequest request) {

        String requestId = (request != null && request.getRequestId() != null)
            ? request.getRequestId()
            : UUID.randomUUID().toString();

        CircuitBreakerConfig config = configService.getConfig(service)
            .orElseGet(() -> {
                CircuitBreakerConfig newConfig = new CircuitBreakerConfig();
                configService.registerService(service, newConfig);
                return newConfig;
            });

        CircuitBreaker breaker = circuitBreakerRegistry.getOrCreate(service, config);

        if (!breaker.allowRequest(requestId)) {
            statsService.recordCall(service, requestId, false);
            log.warn("Request {} to service {} blocked - circuit breaker is OPEN", requestId, service);
            return ResponseEntity.status(HttpStatus.SERVICE_UNAVAILABLE)
                .body("Service unavailable: circuit breaker is OPEN");
        }

        boolean callSuccess = executeCall(service, request);

        if (callSuccess) {
            breaker.recordSuccess(requestId);
            statsService.recordCall(service, requestId, true);
            return ResponseEntity.ok("Request " + requestId + " processed successfully");
        } else {
            breaker.recordFailure(requestId);
            statsService.recordCall(service, requestId, false);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body("Request " + requestId + " failed");
        }
    }

    private boolean executeCall(String service, CallRequest request) {
        if (request != null) {
            return request.isSuccess();
        }
        return true;
    }
}
