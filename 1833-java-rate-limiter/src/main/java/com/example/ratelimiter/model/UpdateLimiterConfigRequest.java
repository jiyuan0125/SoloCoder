package com.example.ratelimiter.model;

import jakarta.validation.constraints.Min;

public class UpdateLimiterConfigRequest {

    @Min(value = 1, message = "capacity must be at least 1")
    private int capacity;

    @Min(value = 1, message = "rate must be at least 1")
    private double rate;

    public UpdateLimiterConfigRequest() {}

    public UpdateLimiterConfigRequest(int capacity, double rate) {
        this.capacity = capacity;
        this.rate = rate;
    }

    public int getCapacity() {
        return capacity;
    }

    public void setCapacity(int capacity) {
        this.capacity = capacity;
    }

    public double getRate() {
        return rate;
    }

    public void setRate(double rate) {
        this.rate = rate;
    }
}
