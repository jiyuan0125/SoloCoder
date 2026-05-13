package com.example.cachemiddleware.stats;

import org.springframework.stereotype.Component;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicLong;

@Component
public class StatsCollector {

    private final Map<String, AtomicLong> hitCounts = new ConcurrentHashMap<>();
    private final Map<String, AtomicLong> missCounts = new ConcurrentHashMap<>();
    private final Map<String, AtomicLong> evictionCounts = new ConcurrentHashMap<>();

    private AtomicLong getOrCreate(Map<String, AtomicLong> map, String namespace) {
        return map.computeIfAbsent(namespace, k -> new AtomicLong(0));
    }

    public void recordHit(String namespace) {
        getOrCreate(hitCounts, namespace).incrementAndGet();
    }

    public void recordMiss(String namespace) {
        getOrCreate(missCounts, namespace).incrementAndGet();
    }

    public void recordEviction(String namespace) {
        getOrCreate(evictionCounts, namespace).incrementAndGet();
    }

    public long getHitCount(String namespace) {
        return getOrCreate(hitCounts, namespace).get();
    }

    public long getMissCount(String namespace) {
        return getOrCreate(missCounts, namespace).get();
    }

    public long getEvictionCount(String namespace) {
        return getOrCreate(evictionCounts, namespace).get();
    }

    public void reset(String namespace) {
        hitCounts.remove(namespace);
        missCounts.remove(namespace);
        evictionCounts.remove(namespace);
    }
}
