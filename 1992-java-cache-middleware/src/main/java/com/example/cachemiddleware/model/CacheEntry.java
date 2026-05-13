package com.example.cachemiddleware.model;

import lombok.AllArgsConstructor;
import lombok.Data;

@Data
@AllArgsConstructor
public class CacheEntry {
    private String key;
    private Object value;
    private long createdAt;
    private long ttlSeconds;
    private int accessCount;

    public CacheEntry(String key, Object value, long ttlSeconds) {
        this.key = key;
        this.value = value;
        this.createdAt = System.currentTimeMillis();
        this.ttlSeconds = ttlSeconds;
        this.accessCount = 0;
    }

    public boolean isExpired() {
        if (ttlSeconds <= 0) {
            return false;
        }
        return System.currentTimeMillis() - createdAt > ttlSeconds * 1000;
    }

    public void incrementAccessCount() {
        this.accessCount++;
    }
}
