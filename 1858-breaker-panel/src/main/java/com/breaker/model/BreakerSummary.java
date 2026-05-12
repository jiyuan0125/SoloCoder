package com.breaker.model;

public class BreakerSummary {
    private String service;
    private BreakerState state;
    private BreakerConfig config;

    public BreakerSummary(String service, BreakerState state, BreakerConfig config) {
        this.service = service;
        this.state = state;
        this.config = config;
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
}
