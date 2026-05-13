package com.graphql.gateway.model;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicLong;

public class QueryStats {
    private final AtomicLong totalQueries = new AtomicLong(0);
    private final AtomicLong totalParseTimeNanos = new AtomicLong(0);
    private final AtomicLong totalExecutionTimeNanos = new AtomicLong(0);
    private final Map<String, BackendCallStats> backendStats = new ConcurrentHashMap<>();

    public void recordQuery(long parseTimeNanos, long executionTimeNanos) {
        totalQueries.incrementAndGet();
        totalParseTimeNanos.addAndGet(parseTimeNanos);
        totalExecutionTimeNanos.addAndGet(executionTimeNanos);
    }

    public void recordBackendCall(String backendName, long durationNanos, boolean success) {
        backendStats.computeIfAbsent(backendName, k -> new BackendCallStats())
                .recordCall(durationNanos, success);
    }

    public long getTotalQueries() {
        return totalQueries.get();
    }

    public double getAverageParseTimeMillis() {
        long count = totalQueries.get();
        if (count == 0) return 0.0;
        return (totalParseTimeNanos.get() / (double) count) / 1_000_000.0;
    }

    public double getAverageExecutionTimeMillis() {
        long count = totalQueries.get();
        if (count == 0) return 0.0;
        return (totalExecutionTimeNanos.get() / (double) count) / 1_000_000.0;
    }

    public Map<String, BackendCallStats> getBackendStats() {
        return backendStats;
    }

    public static class BackendCallStats {
        private final AtomicLong callCount = new AtomicLong(0);
        private final AtomicLong successCount = new AtomicLong(0);
        private final AtomicLong totalTimeNanos = new AtomicLong(0);

        public void recordCall(long durationNanos, boolean success) {
            callCount.incrementAndGet();
            if (success) successCount.incrementAndGet();
            totalTimeNanos.addAndGet(durationNanos);
        }

        public long getCallCount() {
            return callCount.get();
        }

        public long getSuccessCount() {
            return successCount.get();
        }

        public double getAverageTimeMillis() {
            long count = callCount.get();
            if (count == 0) return 0.0;
            return (totalTimeNanos.get() / (double) count) / 1_000_000.0;
        }
    }
}
