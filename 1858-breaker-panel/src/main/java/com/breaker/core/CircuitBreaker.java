package com.breaker.core;

import com.breaker.model.BreakerConfig;
import com.breaker.model.BreakerState;
import com.breaker.model.CallRecord;
import com.breaker.model.StateChangeEvent;

import java.time.LocalDateTime;
import java.util.Deque;
import java.util.LinkedList;
import java.util.concurrent.locks.ReentrantLock;
import java.util.function.Consumer;

public class CircuitBreaker {
    private final String serviceName;
    private volatile BreakerConfig config;
    private volatile BreakerState state;
    private final Deque<CallRecord> recentCalls;
    private volatile StateChangeEvent lastStateChange;
    private volatile LocalDateTime openStartTime;
    private volatile LocalDateTime scheduledHalfOpenTime;
    private int consecutiveFailures;
    private final ReentrantLock lock;
    private Consumer<CircuitBreaker> stateChangeListener;

    private static final int MAX_RECENT_CALLS = 10;

    public CircuitBreaker(String serviceName, BreakerConfig config) {
        this.serviceName = serviceName;
        this.config = config;
        this.state = BreakerState.CLOSED;
        this.recentCalls = new LinkedList<>();
        this.consecutiveFailures = 0;
        this.lock = new ReentrantLock();
    }

    public void setStateChangeListener(Consumer<CircuitBreaker> listener) {
        this.stateChangeListener = listener;
    }

    public String getServiceName() {
        return serviceName;
    }

    public BreakerConfig getConfig() {
        return config;
    }

    public BreakerState getState() {
        checkAndTransitionFromOpen();
        return state;
    }

    public Deque<CallRecord> getRecentCalls() {
        synchronized (recentCalls) {
            return new LinkedList<>(recentCalls);
        }
    }

    public StateChangeEvent getLastStateChange() {
        return lastStateChange;
    }

    public void updateConfig(BreakerConfig newConfig) {
        lock.lock();
        try {
            this.config = newConfig;
            if (state == BreakerState.OPEN && openStartTime != null) {
                this.scheduledHalfOpenTime = openStartTime.plusSeconds(newConfig.getOpenDurationSeconds());
            }
        } finally {
            lock.unlock();
        }
    }

    private void checkAndTransitionFromOpen() {
        if (state == BreakerState.OPEN && scheduledHalfOpenTime != null) {
            if (LocalDateTime.now().isAfter(scheduledHalfOpenTime)) {
                lock.lock();
                try {
                    if (state == BreakerState.OPEN && LocalDateTime.now().isAfter(scheduledHalfOpenTime)) {
                        transitionTo(BreakerState.HALF_OPEN, "Cooling down period expired, entering half-open state");
                    }
                } finally {
                    lock.unlock();
                }
            }
        }
    }

    public boolean allowRequest() {
        checkAndTransitionFromOpen();
        BreakerState current = state;
        if (current == BreakerState.OPEN) {
            return false;
        }
        if (current == BreakerState.CLOSED) {
            return true;
        }
        return true;
    }

    public void recordSuccess() {
        lock.lock();
        try {
            addCallRecord(true);
            consecutiveFailures = 0;

            if (state == BreakerState.HALF_OPEN) {
                transitionTo(BreakerState.CLOSED, "Probe request succeeded, closing circuit");
            }
        } finally {
            lock.unlock();
        }
    }

    public void recordFailure() {
        lock.lock();
        try {
            addCallRecord(false);
            consecutiveFailures++;

            if (state == BreakerState.CLOSED) {
                if (consecutiveFailures >= config.getFailureThreshold()) {
                    transitionTo(BreakerState.OPEN, "Consecutive failures reached threshold: " + consecutiveFailures);
                }
            } else if (state == BreakerState.HALF_OPEN) {
                transitionTo(BreakerState.OPEN, "Probe request failed in half-open state");
            }
        } finally {
            lock.unlock();
        }
    }

    private void transitionTo(BreakerState newState, String reason) {
        BreakerState oldState = this.state;
        this.state = newState;

        LocalDateTime now = LocalDateTime.now();
        this.lastStateChange = new StateChangeEvent(oldState, newState, reason, now);

        if (newState == BreakerState.OPEN) {
            this.openStartTime = now;
            this.scheduledHalfOpenTime = now.plusSeconds(config.getOpenDurationSeconds());
        }

        if (stateChangeListener != null) {
            stateChangeListener.accept(this);
        }
    }

    private void addCallRecord(boolean success) {
        synchronized (recentCalls) {
            recentCalls.addLast(new CallRecord(success, LocalDateTime.now()));
            if (recentCalls.size() > MAX_RECENT_CALLS) {
                recentCalls.removeFirst();
            }
        }
    }
}
