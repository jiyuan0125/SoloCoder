package com.example.circuitbreaker.dto;

import lombok.Data;

import javax.validation.constraints.Min;

@Data
public class CircuitBreakerConfigRequest {
    @Min(1)
    private int failureThreshold;

    @Min(1000)
    private long openDuration;

    @Min(1)
    private int halfOpenPermittedCalls;
}
