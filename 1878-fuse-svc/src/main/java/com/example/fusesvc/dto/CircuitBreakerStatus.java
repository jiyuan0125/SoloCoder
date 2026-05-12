package com.example.fusesvc.dto;

import com.example.fusesvc.model.CallResult;
import com.example.fusesvc.model.CircuitBreakerState;
import java.util.List;

public class CircuitBreakerStatus {
    private String serviceName;
    private CircuitBreakerState state;
    private int failureThreshold;
    private int openDurationSeconds;
    private int currentFailureCount;
    private List<CallResult> recentCalls;

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

    public int getCurrentFailureCount() {
        return currentFailureCount;
    }

    public void setCurrentFailureCount(int currentFailureCount) {
        this.currentFailureCount = currentFailureCount;
    }

    public List<CallResult> getRecentCalls() {
        return recentCalls;
    }

    public void setRecentCalls(List<CallResult> recentCalls) {
        this.recentCalls = recentCalls;
    }
}
