package com.example.circuitbreaker.breaker;

import com.example.circuitbreaker.config.CircuitBreakerConfig;
import com.example.circuitbreaker.stats.StatsService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;

@Slf4j
@Service
@RequiredArgsConstructor
public class CircuitBreakerRegistry {

    private final Map<String, CircuitBreaker> breakers = new ConcurrentHashMap<>();
    private final StatsService statsService;

    public CircuitBreaker getOrCreate(String serviceName, int failureThreshold, int openDurationSeconds) {
        return breakers.computeIfAbsent(serviceName, name ->
            new CircuitBreaker(name, failureThreshold, openDurationSeconds, statsService));
    }

    public CircuitBreaker getOrCreate(String serviceName, CircuitBreakerConfig config) {
        return getOrCreate(serviceName, config.getFailureThreshold(), config.getOpenDurationSeconds());
    }

    public Optional<CircuitBreaker> get(String serviceName) {
        return Optional.ofNullable(breakers.get(serviceName));
    }

    public void remove(String serviceName) {
        breakers.remove(serviceName);
        log.info("Removed circuit breaker for service: {}", serviceName);
    }

    public Map<String, CircuitBreaker> getAll() {
        return new ConcurrentHashMap<>(breakers);
    }

    public boolean exists(String serviceName) {
        return breakers.containsKey(serviceName);
    }
}
