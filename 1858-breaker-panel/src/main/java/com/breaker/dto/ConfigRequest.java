package com.breaker.dto;

import com.fasterxml.jackson.annotation.JsonProperty;

import javax.validation.constraints.Min;

public class ConfigRequest {
    @JsonProperty("failure_threshold")
    @Min(value = 1, message = "failure_threshold must be at least 1")
    private Integer failureThreshold;

    @JsonProperty("open_duration_seconds")
    @Min(value = 1, message = "open_duration_seconds must be at least 1")
    private Integer openDurationSeconds;

    public ConfigRequest() {}

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
