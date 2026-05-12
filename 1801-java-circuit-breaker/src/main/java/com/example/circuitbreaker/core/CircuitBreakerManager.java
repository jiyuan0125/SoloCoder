package com.example.circuitbreaker.core;

import com.example.circuitbreaker.config.CircuitBreakerProperties;
import lombok.RequiredArgsConstructor;
import org.springframework.context.ApplicationEventPublisher;
import org.springframework.stereotype.Component;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

@Component
@RequiredArgsConstructor
public class CircuitBreakerManager {
    private final CircuitBreakerProperties properties;
    private final ApplicationEventPublisher eventPublisher;
    private final Map<String, CircuitBreaker> circuitBreakers = new ConcurrentHashMap<>();

    public CircuitBreaker getOrCreate(String serviceName) {
        return circuitBreakers.computeIfAbsent(serviceName, this::createCircuitBreaker);
    }

    public CircuitBreaker get(String serviceName) {
        return circuitBreakers.get(serviceName);
    }

    public Map<String, CircuitBreaker> getAll() {
        return new ConcurrentHashMap<>(circuitBreakers);
    }

    private CircuitBreaker createCircuitBreaker(String serviceName) {
        CircuitBreakerProperties.ServiceConfig config = properties.getConfigForService(serviceName);
        return new CircuitBreaker(
                serviceName,
                eventPublisher,
                config.getFailureThreshold(),
                config.getOpenDuration(),
                config.getHalfOpenPermittedCalls()
        );
    }
}
