package com.example.fusesvc.service;

import com.example.fusesvc.model.CallResult;
import com.example.fusesvc.model.CircuitBreakerState;
import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.Collections;
import java.util.LinkedList;
import java.util.List;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.concurrent.atomic.AtomicLong;
import java.util.concurrent.atomic.AtomicReference;
import java.util.concurrent.locks.Lock;
import java.util.concurrent.locks.ReentrantLock;

public class CircuitBreaker {
    private static final int MAX_RECENT_CALLS = 10;
    
    private final String serviceName;
    private final AtomicReference<CircuitBreakerState> state;
    private final AtomicInteger failureThreshold;
    private final AtomicInteger openDurationSeconds;
    private final AtomicInteger currentFailureCount;
    private final AtomicLong openUntilTime;
    private final List<CallResult> recentCalls;
    private final Lock stateLock;

    public CircuitBreaker(String serviceName, int failureThreshold, int openDurationSeconds) {
        this.serviceName = serviceName;
        this.state = new AtomicReference<>(CircuitBreakerState.CLOSED);
        this.failureThreshold = new AtomicInteger(failureThreshold);
        this.openDurationSeconds = new AtomicInteger(openDurationSeconds);
        this.currentFailureCount = new AtomicInteger(0);
        this.openUntilTime = new AtomicLong(0);
        this.recentCalls = Collections.synchronizedList(new LinkedList<>());
        this.stateLock = new ReentrantLock();
    }

    public String getServiceName() {
        return serviceName;
    }

    public CircuitBreakerState getState() {
        checkAndTransitionFromOpen();
        return state.get();
    }

    public int getFailureThreshold() {
        return failureThreshold.get();
    }

    public int getOpenDurationSeconds() {
        return openDurationSeconds.get();
    }

    public int getCurrentFailureCount() {
        return currentFailureCount.get();
    }

    public List<CallResult> getRecentCalls() {
        synchronized (recentCalls) {
            return new ArrayList<>(recentCalls);
        }
    }

    public void updateConfig(int newFailureThreshold, int newOpenDurationSeconds) {
        this.failureThreshold.set(newFailureThreshold);
        this.openDurationSeconds.set(newOpenDurationSeconds);
    }

    public boolean allowRequest() {
        CircuitBreakerState currentState = getState();
        
        if (currentState == CircuitBreakerState.OPEN) {
            return false;
        }
        
        if (currentState == CircuitBreakerState.HALF_OPEN) {
            stateLock.lock();
            try {
                if (state.get() == CircuitBreakerState.HALF_OPEN) {
                    return true;
                }
                return state.get() == CircuitBreakerState.CLOSED;
            } finally {
                stateLock.unlock();
            }
        }
        
        return true;
    }

    public void report(boolean success) {
        addRecentCall(new CallResult(LocalDateTime.now(), success));
        
        stateLock.lock();
        try {
            CircuitBreakerState currentState = state.get();
            
            switch (currentState) {
                case CLOSED:
                    handleClosedState(success);
                    break;
                case OPEN:
                    break;
                case HALF_OPEN:
                    handleHalfOpenState(success);
                    break;
            }
        } finally {
            stateLock.unlock();
        }
    }

    private void handleClosedState(boolean success) {
        if (success) {
            currentFailureCount.set(0);
        } else {
            int newCount = currentFailureCount.incrementAndGet();
            if (newCount >= failureThreshold.get()) {
                transitionToOpen();
            }
        }
    }

    private void handleHalfOpenState(boolean success) {
        if (success) {
            transitionToClosed();
        } else {
            transitionToOpen();
        }
    }

    private void transitionToOpen() {
        state.set(CircuitBreakerState.OPEN);
        openUntilTime.set(System.currentTimeMillis() + openDurationSeconds.get() * 1000L);
        currentFailureCount.set(0);
    }

    private void transitionToClosed() {
        state.set(CircuitBreakerState.CLOSED);
        currentFailureCount.set(0);
        openUntilTime.set(0);
    }

    private void checkAndTransitionFromOpen() {
        if (state.get() == CircuitBreakerState.OPEN) {
            stateLock.lock();
            try {
                if (state.get() == CircuitBreakerState.OPEN && 
                    System.currentTimeMillis() >= openUntilTime.get()) {
                    state.set(CircuitBreakerState.HALF_OPEN);
                }
            } finally {
                stateLock.unlock();
            }
        }
    }

    private void addRecentCall(CallResult callResult) {
        synchronized (recentCalls) {
            recentCalls.add(callResult);
            while (recentCalls.size() > MAX_RECENT_CALLS) {
                recentCalls.remove(0);
            }
        }
    }
}
