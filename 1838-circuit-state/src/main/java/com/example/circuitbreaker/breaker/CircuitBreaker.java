package com.example.circuitbreaker.breaker;

import com.example.circuitbreaker.stats.ChangeReason;
import com.example.circuitbreaker.stats.StatsService;
import lombok.Getter;
import lombok.extern.slf4j.Slf4j;

import java.time.Instant;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.concurrent.atomic.AtomicReference;

@Slf4j
public class CircuitBreaker {

    @Getter
    private final String serviceName;
    private final int failureThreshold;
    private final int openDurationSeconds;

    private final AtomicReference<CircuitBreakerState> state = new AtomicReference<>(CircuitBreakerState.CLOSED);
    private final AtomicInteger consecutiveFailures = new AtomicInteger(0);
    private volatile Instant openTime;

    private final StatsService statsService;

    public CircuitBreaker(String serviceName, int failureThreshold, int openDurationSeconds, StatsService statsService) {
        this.serviceName = serviceName;
        this.failureThreshold = failureThreshold;
        this.openDurationSeconds = openDurationSeconds;
        this.statsService = statsService;
    }

    public CircuitBreakerState getState() {
        return state.get();
    }

    public synchronized boolean allowRequest(String requestId) {
        if (state.get() == CircuitBreakerState.OPEN) {
            if (isOpenDurationExceeded()) {
                transitionTo(CircuitBreakerState.HALF_OPEN, requestId, ChangeReason.HEALTH_CHECK);
            }
        }
        return state.get() != CircuitBreakerState.OPEN;
    }

    private boolean isOpenDurationExceeded() {
        if (openTime == null) return true;
        return Instant.now().isAfter(openTime.plusSeconds(openDurationSeconds));
    }

    public synchronized void recordSuccess(String requestId) {
        consecutiveFailures.set(0);
        if (state.get() == CircuitBreakerState.HALF_OPEN) {
            transitionTo(CircuitBreakerState.CLOSED, requestId, ChangeReason.SUCCESS_AFTER_HALF_OPEN);
        }
    }

    public synchronized void recordFailure(String requestId) {
        int failures = consecutiveFailures.incrementAndGet();
        log.debug("Service {}: consecutive failures = {}", serviceName, failures);

        if (state.get() == CircuitBreakerState.HALF_OPEN) {
            transitionTo(CircuitBreakerState.OPEN, requestId, ChangeReason.FAILURE_AFTER_HALF_OPEN);
            return;
        }

        if (state.get() == CircuitBreakerState.CLOSED && failures >= failureThreshold) {
            transitionTo(CircuitBreakerState.OPEN, requestId, ChangeReason.CONSECUTIVE_FAILURES);
        }
    }

    public synchronized void transitionTo(CircuitBreakerState newState, String requestId, ChangeReason reason) {
        CircuitBreakerState oldState = state.get();
        if (oldState == newState) return;

        log.info("Service {} circuit breaker: {} -> {} (reason: {})", serviceName, oldState, newState, reason);
        state.set(newState);

        if (newState == CircuitBreakerState.OPEN) {
            openTime = Instant.now();
            consecutiveFailures.set(0);
        }

        statsService.recordStateChange(serviceName, oldState, newState, reason, requestId);
    }

    public void transitionTo(CircuitBreakerState newState, ChangeReason reason) {
        transitionTo(newState, null, reason);
    }

    public int getFailureThreshold() {
        return failureThreshold;
    }

    public int getOpenDurationSeconds() {
        return openDurationSeconds;
    }

    public Instant getOpenTime() {
        return openTime;
    }

    public int getConsecutiveFailures() {
        return consecutiveFailures.get();
    }
}
