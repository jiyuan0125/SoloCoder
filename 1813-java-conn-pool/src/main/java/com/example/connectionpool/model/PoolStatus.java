package com.example.connectionpool.model;

import lombok.Builder;
import lombok.Data;

@Data
@Builder
public class PoolStatus {
    private String poolName;
    private int activeCount;
    private int idleCount;
    private int waitingQueueSize;
    private long totalBorrowCount;
    private long totalTimeoutRejectCount;
    private long totalLeakDetectCount;
    private int maxConnections;
    private long waitTimeoutMs;
    private int leakThresholdSeconds;
    private int maxQueueSize;
}
