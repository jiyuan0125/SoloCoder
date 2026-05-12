package com.breaker.model;

import java.time.LocalDateTime;

public class NotificationPayload {
    private String service;
    private BreakerState fromState;
    private BreakerState toState;
    private String reason;
    private LocalDateTime timestamp;

    public NotificationPayload() {}

    public NotificationPayload(String service, StateChangeEvent event) {
        this.service = service;
        this.fromState = event.getFromState();
        this.toState = event.getToState();
        this.reason = event.getReason();
        this.timestamp = event.getTimestamp();
    }

    public String getService() {
        return service;
    }

    public void setService(String service) {
        this.service = service;
    }

    public BreakerState getFromState() {
        return fromState;
    }

    public void setFromState(BreakerState fromState) {
        this.fromState = fromState;
    }

    public BreakerState getToState() {
        return toState;
    }

    public void setToState(BreakerState toState) {
        this.toState = toState;
    }

    public String getReason() {
        return reason;
    }

    public void setReason(String reason) {
        this.reason = reason;
    }

    public LocalDateTime getTimestamp() {
        return timestamp;
    }

    public void setTimestamp(LocalDateTime timestamp) {
        this.timestamp = timestamp;
    }
}
