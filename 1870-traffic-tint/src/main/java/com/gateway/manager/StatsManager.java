package com.gateway.manager;

import com.gateway.model.Stats;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

public class StatsManager {
    private final Map<String, Stats> statsMap = new ConcurrentHashMap<>();

    public void incrementRequestCount(String versionName) {
        getOrCreateStats(versionName).incrementRequestCount();
    }

    public void incrementErrorCount(String versionName) {
        getOrCreateStats(versionName).incrementErrorCount();
    }

    private Stats getOrCreateStats(String versionName) {
        return statsMap.computeIfAbsent(versionName, Stats::new);
    }

    public Stats getStats(String versionName) {
        return statsMap.get(versionName);
    }

    public Map<String, Stats> getAllStats() {
        return statsMap;
    }
}
