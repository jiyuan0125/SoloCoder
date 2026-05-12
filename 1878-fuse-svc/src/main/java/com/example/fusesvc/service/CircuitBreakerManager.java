package com.example.fusesvc.service;

import com.example.fusesvc.dto.CircuitBreakerOverview;
import com.example.fusesvc.dto.CircuitBreakerStatus;
import java.util.Collection;
import java.util.List;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;
import org.springframework.stereotype.Service;

@Service
public class CircuitBreakerManager {
    private static final int DEFAULT_FAILURE_THRESHOLD = 5;
    private static final int DEFAULT_OPEN_DURATION_SECONDS = 30;
    
    private final ConcurrentHashMap<String, CircuitBreaker> circuitBreakers;

    public CircuitBreakerManager() {
        this.circuitBreakers = new ConcurrentHashMap<>();
    }

    public CircuitBreaker getOrCreate(String serviceName) {
        return circuitBreakers.computeIfAbsent(serviceName, name -> 
            new CircuitBreaker(name, DEFAULT_FAILURE_THRESHOLD, DEFAULT_OPEN_DURATION_SECONDS)
        );
    }

    public CircuitBreaker get(String serviceName) {
        return circuitBreakers.get(serviceName);
    }

    public boolean exists(String serviceName) {
        return circuitBreakers.containsKey(serviceName);
    }

    public void updateConfig(String serviceName, int failureThreshold, int openDurationSeconds) {
        CircuitBreaker breaker = getOrCreate(serviceName);
        breaker.updateConfig(failureThreshold, openDurationSeconds);
    }

    public void report(String serviceName, boolean success) {
        CircuitBreaker breaker = getOrCreate(serviceName);
        breaker.report(success);
    }

    public CircuitBreakerStatus getStatus(String serviceName) {
        CircuitBreaker breaker = getOrCreate(serviceName);
        return toStatus(breaker);
    }

    public List<CircuitBreakerOverview> getAllOverviews() {
        Collection<CircuitBreaker> breakers = circuitBreakers.values();
        return breakers.stream()
            .map(this::toOverview)
            .collect(Collectors.toList());
    }

    private CircuitBreakerStatus toStatus(CircuitBreaker breaker) {
        CircuitBreakerStatus status = new CircuitBreakerStatus();
        status.setServiceName(breaker.getServiceName());
        status.setState(breaker.getState());
        status.setFailureThreshold(breaker.getFailureThreshold());
        status.setOpenDurationSeconds(breaker.getOpenDurationSeconds());
        status.setCurrentFailureCount(breaker.getCurrentFailureCount());
        status.setRecentCalls(breaker.getRecentCalls());
        return status;
    }

    private CircuitBreakerOverview toOverview(CircuitBreaker breaker) {
        return new CircuitBreakerOverview(
            breaker.getServiceName(),
            breaker.getState(),
            breaker.getFailureThreshold(),
            breaker.getOpenDurationSeconds()
        );
    }
}
