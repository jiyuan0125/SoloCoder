package com.loadbalancer.model;

import lombok.Data;
import lombok.AllArgsConstructor;
import lombok.NoArgsConstructor;

import java.time.Instant;
import java.util.Deque;
import java.util.ArrayDeque;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.concurrent.locks.ReentrantReadWriteLock;

@Data
public class Node {
    private String id;
    private String address;
    private int weight;
    private int currentWeight;
    private NodeStatus status;
    private int healthCheckIntervalSeconds;
    private int healthCheckTimeoutSeconds;
    private int consecutiveSuccesses;
    private int consecutiveFailures;
    private AtomicInteger activeConnections;
    private Deque<HealthCheckResult> recentChecks;
    private Instant lastHealthCheckTime;
    private ReentrantReadWriteLock lock;

    public Node(String id, String address, int weight) {
        this.id = id;
        this.address = address;
        this.weight = weight;
        this.currentWeight = 0;
        this.status = NodeStatus.NEW_REGISTERED;
        this.consecutiveSuccesses = 0;
        this.consecutiveFailures = 0;
        this.activeConnections = new AtomicInteger(0);
        this.recentChecks = new ArrayDeque<>(3);
        this.lock = new ReentrantReadWriteLock();
    }

    public void addHealthCheckResult(HealthCheckResult result) {
        if (recentChecks.size() >= 3) {
            recentChecks.removeFirst();
        }
        recentChecks.addLast(result);
        this.lastHealthCheckTime = result.getTimestamp();
    }

    public void incrementActiveConnections() {
        activeConnections.incrementAndGet();
    }

    public void decrementActiveConnections() {
        activeConnections.decrementAndGet();
    }

    public int getActiveConnectionsCount() {
        return activeConnections.get();
    }
}