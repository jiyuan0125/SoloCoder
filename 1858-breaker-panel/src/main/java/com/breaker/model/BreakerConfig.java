package com.breaker.model;

import javax.validation.constraints.Min;

public class BreakerConfig {
    @Min(value = 1, message = "failure_threshold must be at least 1")
    private Integer failureThreshold;

    @Min(value = 1, message = "open_duration_seconds must be at least 1")
    private Integer openDurationSeconds;

    public BreakerConfig() {}

    public BreakerConfig(Integer failureThreshold, Integer openDurationSeconds) {
        this.failureThreshold = failureThreshold;
        this.openDurationSeconds = openDurationSeconds;
    }

    public Integer getFailureThreshold() {
        return failureThreshold;
    }

    public void setFailureThreshold(Integer failureThreshold) {
        this.failureThreshold = failureThreshold;
    }

    public Integer getOpenDurationSeconds() {
        return openDurationSeconds;
    }

    public void setOpenDurationSeconds(Integer openDurationSeconds) {
        this.openDurationSeconds = openDurationSeconds;
    }
}
