package com.example.circuitbreaker.dto;

import com.example.circuitbreaker.core.CircuitBreaker;
import com.example.circuitbreaker.core.CircuitState;
import lombok.Builder;
import lombok.Data;

@Data
@Builder
public class CircuitBreakerStatusResponse {
    private String serviceName;
    private CircuitState state;
    private int failureCount;
    private long openRemainingMs;
    private int failureThreshold;
    private long openDuration;
    private int halfOpenPermittedCalls;

    public static CircuitBreakerStatusResponse from(String serviceName, CircuitBreaker cb) {
        return CircuitBreakerStatusResponse.builder()
                .serviceName(serviceName)
                .state(cb.getState())
                .failureCount(cb.getFailureCount())
                .openRemainingMs(cb.getOpenRemainingMs())
                .failureThreshold(cb.getFailureThreshold())
                .openDuration(cb.getOpenDuration())
                .halfOpenPermittedCalls(cb.getHalfOpenPermittedCalls())
                .build();
    }
}
