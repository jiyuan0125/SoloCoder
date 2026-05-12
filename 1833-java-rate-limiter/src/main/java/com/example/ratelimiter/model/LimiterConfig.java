package com.example.ratelimiter.model;

import jakarta.validation.constraints.Min;
import jakarta.validation.constraints.NotBlank;

public class LimiterConfig {

    @NotBlank(message = "path is required")
    private String path;

    @Min(value = 1, message = "capacity must be at least 1")
    private int capacity;

    @Min(value = 1, message = "rate must be at least 1")
    private double rate;

    public LimiterConfig() {}

    public LimiterConfig(String path, int capacity, double rate) {
        this.path = path;
        this.capacity = capacity;
        this.rate = rate;
    }

    public String getPath() {
        return path;
    }

    public void setPath(String path) {
        this.path = path;
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
