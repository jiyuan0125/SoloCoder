package com.example.circuitbreaker.stats;

public enum ChangeReason {
    CONSECUTIVE_FAILURES,
    HEALTH_CHECK,
    SUCCESS_AFTER_HALF_OPEN,
    FAILURE_AFTER_HALF_OPEN,
    MANUAL_TRANSITION
}
