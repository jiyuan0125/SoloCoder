package com.loadbalancer.model;

import java.time.LocalDateTime;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.concurrent.atomic.AtomicLong;

public class Instance {
    private final String id;
    private final String host;
    private final int port;
    private volatile LocalDateTime registeredAt;
    private volatile boolean healthy;
    private volatile boolean online;
    private final AtomicInteger activeRequests;
    private final AtomicLong totalRequests;
    private final AtomicLong successfulRequests;
    private final AtomicLong failedRequests;

    public Instance(String id, String host, int port) {
        this.id = id;
        this.host = host;
        this.port = port;
        this.registeredAt = LocalDateTime.now();
        this.healthy = false;
        this.online = true;
        this.activeRequests = new AtomicInteger(0);
        this.totalRequests = new AtomicLong(0);
        this.successfulRequests = new AtomicLong(0);
        this.failedRequests = new AtomicLong(0);
    }

    public String getId() {
        return id;
    }

    public String getHost() {
        return host;
    }

    public int getPort() {
        return port;
    }

    public String getUrl() {
        return "http://" + host + ":" + port;
    }

    public LocalDateTime getRegisteredAt() {
        return registeredAt;
    }

    public void setRegisteredAt(LocalDateTime registeredAt) {
        this.registeredAt = registeredAt;
    }

    public boolean isHealthy() {
        return healthy;
    }

    public void setHealthy(boolean healthy) {
        this.healthy = healthy;
    }

    public boolean isOnline() {
        return online;
    }

    public void setOnline(boolean online) {
        this.online = online;
    }

    public int getActiveRequests() {
        return activeRequests.get();
    }

    public int incrementActiveRequests() {
        return activeRequests.incrementAndGet();
    }

    public int decrementActiveRequests() {
        return activeRequests.decrementAndGet();
    }

    public long getTotalRequests() {
        return totalRequests.get();
    }

    public long incrementTotalRequests() {
        return totalRequests.incrementAndGet();
    }

    public long getSuccessfulRequests() {
        return successfulRequests.get();
    }

    public long incrementSuccessfulRequests() {
        return successfulRequests.incrementAndGet();
    }

    public long getFailedRequests() {
        return failedRequests.get();
    }

    public long incrementFailedRequests() {
        return failedRequests.incrementAndGet();
    }

    public double getSuccessRate() {
        if (totalRequests.get() == 0) {
            return 100.0;
        }
        return (successfulRequests.get() * 100.0) / totalRequests.get();
    }
}
