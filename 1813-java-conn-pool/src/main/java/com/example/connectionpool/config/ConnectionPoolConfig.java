package com.example.connectionpool.config;

import lombok.Data;

@Data
public class ConnectionPoolConfig {
    private String poolName;
    private int maxConnections = 20;
    private long waitTimeoutMs = 5000;
    private int leakThresholdSeconds = 60;
    private int maxQueueSize = 50;
}
