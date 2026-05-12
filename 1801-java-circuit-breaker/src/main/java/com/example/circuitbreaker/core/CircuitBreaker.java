package com.example.circuitbreaker.core;

import lombok.Getter;
import lombok.extern.slf4j.Slf4j;
import org.springframework.context.ApplicationEventPublisher;

import java.util.concurrent.atomic.AtomicInteger;
import java.util.concurrent.atomic.AtomicLong;
import java.util.concurrent.atomic.AtomicReference;
import java.util.concurrent.locks.ReentrantLock;

@Slf4j
public class CircuitBreaker {
    private final String serviceName;
    private final ApplicationEventPublisher eventPublisher;

    private final AtomicReference<CircuitState> state = new AtomicReference<>(CircuitState.CLOSED);
    private final AtomicInteger failureCount = new AtomicInteger(0);
    private final AtomicInteger halfOpenSuccessCount = new AtomicInteger(0);
    private final AtomicLong openTimestamp = new AtomicLong(0);
    private final ReentrantLock transitionLock = new ReentrantLock();

    @Getter
    private volatile int failureThreshold;
    @Getter
    private volatile long openDuration;
    @Getter
    private volatile int halfOpenPermittedCalls;

    public CircuitBreaker(String serviceName, ApplicationEventPublisher eventPublisher,
                          int failureThreshold, long openDuration, int halfOpenPermittedCalls) {
        this.serviceName = serviceName;
        this.eventPublisher = eventPublisher;
        this.failureThreshold = failureThreshold;
        this.openDuration = openDuration;
        this.halfOpenPermittedCalls = halfOpenPermittedCalls;
    }

    public synchronized void updateConfig(int failureThreshold, long openDuration, int halfOpenPermittedCalls) {
        this.failureThreshold = failureThreshold;
        this.openDuration = openDuration;
        this.halfOpenPermittedCalls = halfOpenPermittedCalls;
    }

    public boolean isCallPermitted() {
        CircuitState currentState = state.get();

        if (currentState == CircuitState.CLOSED) {
            return true;
        }

        if (currentState == CircuitState.OPEN) {
            if (shouldAttemptHalfOpen()) {
                transitionToHalfOpen();
                return true;
            }
            return false;
        }

        if (currentState == CircuitState.HALF_OPEN) {
            int current = halfOpenSuccessCount.incrementAndGet();
            return current <= halfOpenPermittedCalls;
        }

        return false;
    }

    public void recordSuccess() {
        failureCount.set(0);

        if (state.get() == CircuitState.HALF_OPEN) {
            if (halfOpenSuccessCount.get() >= halfOpenPermittedCalls) {
                transitionToClosed("Half-open probe succeeded");
            }
        }
    }

    public void recordFailure() {
        int currentFailures = failureCount.incrementAndGet();
        CircuitState currentState = state.get();

        if (currentState == CircuitState.CLOSED && currentFailures >= failureThreshold) {
            transitionToOpen("Failure threshold exceeded: " + currentFailures);
        } else if (currentState == CircuitState.HALF_OPEN) {
            transitionToOpen("Half-open probe failed");
        }
    }

    public void forceOpen() {
        transitionToOpen("Manually forced open");
    }

    public void forceClosed() {
        transitionToClosed("Manually forced closed");
    }

    public CircuitState getState() {
        return state.get();
    }

    public int getFailureCount() {
        return failureCount.get();
    }

    public long getOpenRemainingMs() {
        if (state.get() != CircuitState.OPEN) {
            return 0;
        }
        long elapsed = System.currentTimeMillis() - openTimestamp.get();
        return Math.max(0, openDuration - elapsed);
    }

    private boolean shouldAttemptHalfOpen() {
        return System.currentTimeMillis() - openTimestamp.get() >= openDuration;
    }

    private void transitionToOpen(String reason) {
        transitionLock.lock();
        try {
            CircuitState from = state.getAndSet(CircuitState.OPEN);
            if (from != CircuitState.OPEN) {
                openTimestamp.set(System.currentTimeMillis());
                publishEvent(from, CircuitState.OPEN, reason);
            }
        } finally {
            transitionLock.unlock();
        }
    }

    private void transitionToHalfOpen() {
        transitionLock.lock();
        try {
            CircuitState from = state.getAndSet(CircuitState.HALF_OPEN);
            if (from != CircuitState.HALF_OPEN) {
                halfOpenSuccessCount.set(0);
                publishEvent(from, CircuitState.HALF_OPEN, "Open duration elapsed");
            }
        } finally {
            transitionLock.unlock();
        }
    }

    private void transitionToClosed(String reason) {
        transitionLock.lock();
        try {
            CircuitState from = state.getAndSet(CircuitState.CLOSED);
            if (from != CircuitState.CLOSED) {
                failureCount.set(0);
                halfOpenSuccessCount.set(0);
                publishEvent(from, CircuitState.CLOSED, reason);
            }
        } finally {
            transitionLock.unlock();
        }
    }

    private void publishEvent(CircuitState from, CircuitState to, String reason) {
        try {
            eventPublisher.publishEvent(
                    new CircuitStateChangedEvent(this, serviceName, from, to, reason)
            );
            log.info("Circuit breaker [{}] state changed: {} -> {}, reason: {}",
                    serviceName, from, to, reason);
        } catch (Exception e) {
            log.error("Failed to publish circuit breaker event for service [{}]", serviceName, e);
        }
    }
}
