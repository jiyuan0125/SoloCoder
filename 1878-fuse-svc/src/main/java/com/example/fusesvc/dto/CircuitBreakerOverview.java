package com.example.fusesvc.dto;

import com.example.fusesvc.model.CircuitBreakerState;

public class CircuitBreakerOverview {
    private String serviceName;
    private CircuitBreakerState state;
    private int failureThreshold;
    private int openDurationSeconds;

    public CircuitBreakerOverview(String serviceName, CircuitBreakerState state,
                                   int failureThreshold, int openDurationSeconds) {
        this.serviceName = serviceName;
        this.state = state;
        this.failureThreshold = failureThreshold;
        this.openDurationSeconds = openDurationSeconds;
    }

    public String getServiceName() {
        return serviceName;
    }

    public void setServiceName(String serviceName) {
        this.serviceName = serviceName;
    }

    public CircuitBreakerState getState() {
        return state;
    }

    public void setState(CircuitBreakerState state) {
        this.state = state;
    }

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
