package com.example.circuitbreaker.config;

import jakarta.validation.constraints.Min;
import lombok.Data;

@Data
public class CircuitBreakerConfig {
    public static final int DEFAULT_FAILURE_THRESHOLD = 5;
    public static final int DEFAULT_OPEN_DURATION_SECONDS = 30;

    @Min(1)
    private int failureThreshold = DEFAULT_FAILURE_THRESHOLD;

    @Min(1)
    private int openDurationSeconds = DEFAULT_OPEN_DURATION_SECONDS;

    public CircuitBreakerConfig() {
    }

    public CircuitBreakerConfig(int failureThreshold, int openDurationSeconds) {
        this.failureThreshold = failureThreshold;
        this.openDurationSeconds = openDurationSeconds;
    }
}
