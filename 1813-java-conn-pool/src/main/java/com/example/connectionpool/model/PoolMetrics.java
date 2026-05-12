package com.example.connectionpool.model;

import java.util.concurrent.atomic.AtomicLong;

public class PoolMetrics {
    private final AtomicLong totalBorrowCount = new AtomicLong(0);
    private final AtomicLong totalTimeoutRejectCount = new AtomicLong(0);
    private final AtomicLong totalLeakDetectCount = new AtomicLong(0);

    public void incrementBorrowCount() {
        totalBorrowCount.incrementAndGet();
    }

    public void incrementTimeoutRejectCount() {
        totalTimeoutRejectCount.incrementAndGet();
    }

    public void incrementLeakDetectCount() {
        totalLeakDetectCount.incrementAndGet();
    }

    public long getTotalBorrowCount() {
        return totalBorrowCount.get();
    }

    public long getTotalTimeoutRejectCount() {
        return totalTimeoutRejectCount.get();
    }

    public long getTotalLeakDetectCount() {
        return totalLeakDetectCount.get();
    }
}
