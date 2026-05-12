package com.breaker.model;

import java.time.LocalDateTime;

public class StateChangeEvent {
    private BreakerState fromState;
    private BreakerState toState;
    private String reason;
    private LocalDateTime timestamp;

    public StateChangeEvent(BreakerState fromState, BreakerState toState, String reason, LocalDateTime timestamp) {
        this.fromState = fromState;
        this.toState = toState;
        this.reason = reason;
        this.timestamp = timestamp;
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
