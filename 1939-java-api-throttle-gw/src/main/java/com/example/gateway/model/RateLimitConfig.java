package com.example.gateway.model;

public class RateLimitConfig {
    private final int capacity;
    private final int refillRatePerSecond;

    public RateLimitConfig(int capacity, int refillRatePerSecond) {
        this.capacity = capacity;
        this.refillRatePerSecond = refillRatePerSecond;
    }

    public int getCapacity() {
        return capacity;
    }

    public int getRefillRatePerSecond() {
        return refillRatePerSecond;
    }
}
