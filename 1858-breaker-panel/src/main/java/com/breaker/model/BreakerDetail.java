package com.breaker.model;

import java.util.List;

public class BreakerDetail {
    private String service;
    private BreakerState state;
    private BreakerConfig config;
    private List<CallRecord> recentCalls;
    private StateChangeEvent lastStateChange;

    public BreakerDetail() {}

    public BreakerDetail(String service, BreakerState state, BreakerConfig config,
                         List<CallRecord> recentCalls, StateChangeEvent lastStateChange) {
        this.service = service;
        this.state = state;
        this.config = config;
        this.recentCalls = recentCalls;
        this.lastStateChange = lastStateChange;
    }

    public String getService() {
        return service;
    }

    public void setService(String service) {
        this.service = service;
    }

    public BreakerState getState() {
        return state;
    }

    public void setState(BreakerState state) {
        this.state = state;
    }

    public BreakerConfig getConfig() {
        return config;
    }

    public void setConfig(BreakerConfig config) {
        this.config = config;
    }

    public List<CallRecord> getRecentCalls() {
        return recentCalls;
    }

    public void setRecentCalls(List<CallRecord> recentCalls) {
        this.recentCalls = recentCalls;
    }

    public StateChangeEvent getLastStateChange() {
        return lastStateChange;
    }

    public void setLastStateChange(StateChangeEvent lastStateChange) {
        this.lastStateChange = lastStateChange;
    }
}
