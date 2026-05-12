package com.example.circuitbreaker.api;

import com.example.circuitbreaker.breaker.CircuitBreaker;
import com.example.circuitbreaker.breaker.CircuitBreakerRegistry;
import com.example.circuitbreaker.breaker.CircuitBreakerState;
import com.example.circuitbreaker.config.CircuitBreakerConfig;
import com.example.circuitbreaker.config.ConfigService;
import com.example.circuitbreaker.stats.CallRecord;
import com.example.circuitbreaker.stats.StatsService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.*;
import java.util.stream.Collectors;

@RestController
@RequestMapping("/breakers")
@RequiredArgsConstructor
public class CircuitBreakerController {

    private final CircuitBreakerRegistry circuitBreakerRegistry;
    private final ConfigService configService;
    private final StatsService statsService;

    @GetMapping
    public ResponseEntity<List<BreakerStatus>> getAllBreakers() {
        Map<String, CircuitBreakerConfig> allConfigs = configService.getAllConfigs();
        Set<String> allServices = new HashSet<>();
        allServices.addAll(allConfigs.keySet());
        allServices.addAll(circuitBreakerRegistry.getAll().keySet());

        List<BreakerStatus> statuses = allServices.stream()
            .map(this::buildBreakerStatus)
            .sorted(Comparator.comparing(BreakerStatus::getServiceName))
            .collect(Collectors.toList());

        return ResponseEntity.ok(statuses);
    }

    @GetMapping("/{service}")
    public ResponseEntity<BreakerStatus> getBreaker(@PathVariable String service) {
        BreakerStatus status = buildBreakerStatus(service);
        return ResponseEntity.ok(status);
    }

    @PutMapping("/{service}/config")
    public ResponseEntity<CircuitBreakerConfig> updateConfig(
        @PathVariable String service,
        @Valid @RequestBody CircuitBreakerConfig config) {

        CircuitBreakerConfig savedConfig = configService.updateConfig(service, config);
        circuitBreakerRegistry.getOrCreate(service, savedConfig);

        return ResponseEntity.ok(savedConfig);
    }

    @DeleteMapping("/{service}/config")
    public ResponseEntity<Void> deleteConfig(@PathVariable String service) {
        if (!configService.exists(service)) {
            return ResponseEntity.notFound().build();
        }

        configService.removeConfig(service);
        statsService.removeServiceRecords(service);

        return ResponseEntity.noContent().build();
    }

    private BreakerStatus buildBreakerStatus(String serviceName) {
        CircuitBreakerConfig config = configService.getConfig(serviceName).orElse(null);

        Optional<CircuitBreaker> breakerOpt = circuitBreakerRegistry.get(serviceName);
        String state;
        Integer threshold = null;
        Integer duration = null;

        if (breakerOpt.isPresent()) {
            CircuitBreaker breaker = breakerOpt.get();
            state = breaker.getState().name();
            threshold = breaker.getFailureThreshold();
            duration = breaker.getOpenDurationSeconds();
        } else if (config != null) {
            state = CircuitBreakerState.CLOSED.name();
            threshold = config.getFailureThreshold();
            duration = config.getOpenDurationSeconds();
        } else {
            state = CircuitBreakerState.CLOSED.name();
            threshold = CircuitBreakerConfig.DEFAULT_FAILURE_THRESHOLD;
            duration = CircuitBreakerConfig.DEFAULT_OPEN_DURATION_SECONDS;
        }

        List<CallRecord> recentCalls = statsService.getRecentCalls(serviceName);
        List<CallRecordView> callViews = recentCalls.stream()
            .map(r -> new CallRecordView(r.getRequestId(), r.isSuccess(), r.getTimestamp()))
            .collect(Collectors.toList());

        return BreakerStatus.builder()
            .serviceName(serviceName)
            .state(state)
            .failureThreshold(threshold)
            .openDurationSeconds(duration)
            .recentCalls(callViews)
            .build();
    }
}
