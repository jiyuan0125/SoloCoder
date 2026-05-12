package com.example.fusesvc.dto;

import jakarta.validation.constraints.Min;

public class CircuitBreakerConfig {
    @Min(value = 1, message = "failure_threshold must be at least 1")
    private int failureThreshold;
    
    @Min(value = 1, message = "open_duration_seconds must be at least 1")
    private int openDurationSeconds;

    public int getFailureThreshold() {
        return failureThreshold;
    }

    public void setFailureThreshold(int failureThreshold) {
        this.failureThreshold = failureThreshold;
    }

    public int getOpenDurationSeconds() {
        return openDurationSeconds;
    }

    public void setOpenDurationSeconds(int openDurationSeconds) {
        this.openDurationSeconds = openDurationSeconds;
    }
}
