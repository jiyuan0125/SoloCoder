package com.example.gateway.stats;

import java.util.Collection;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicLong;

public class StatsStore {
    private final Map<String, StatsSnapshot> statsMap = new ConcurrentHashMap<>();

    public void recordRequest(String version, long responseTimeMs, int statusCode) {
        StatsSnapshot snapshot = statsMap.computeIfAbsent(version, k -> new StatsSnapshot(version));
        snapshot.totalCalls.incrementAndGet();
        snapshot.totalResponseTimeMs.addAndGet(responseTimeMs);
        if (statusCode >= 400 && statusCode < 500) {
            snapshot.clientErrorCount.incrementAndGet();
        } else if (statusCode >= 500) {
            snapshot.serverErrorCount.incrementAndGet();
        }
    }

    public VersionStats getStats(String version) {
        StatsSnapshot snapshot = statsMap.get(version);
        if (snapshot == null) {
            return new VersionStats(version);
        }
        return toVersionStats(snapshot);
    }

    public Collection<VersionStats> getAllStats() {
        return statsMap.values().stream()
                .map(this::toVersionStats)
                .toList();
    }

    private VersionStats toVersionStats(StatsSnapshot snapshot) {
        VersionStats vs = new VersionStats(snapshot.version);
        vs.setTotalCalls(snapshot.totalCalls.get());
        vs.setTotalResponseTimeMs(snapshot.totalResponseTimeMs.get());
        vs.setClientErrorCount(snapshot.clientErrorCount.get());
        vs.setServerErrorCount(snapshot.serverErrorCount.get());
        return vs;
    }

    private static class StatsSnapshot {
        final String version;
        final AtomicLong totalCalls = new AtomicLong(0);
        final AtomicLong totalResponseTimeMs = new AtomicLong(0);
        final AtomicLong clientErrorCount = new AtomicLong(0);
        final AtomicLong serverErrorCount = new AtomicLong(0);

        StatsSnapshot(String version) {
            this.version = version;
        }
    }
}
