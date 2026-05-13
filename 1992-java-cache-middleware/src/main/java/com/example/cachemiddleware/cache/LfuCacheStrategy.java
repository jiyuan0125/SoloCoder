package com.example.cachemiddleware.cache;

import com.example.cachemiddleware.model.CacheEntry;
import com.example.cachemiddleware.model.EvictionPolicy;

import java.util.Collection;
import java.util.Comparator;
import java.util.HashMap;
import java.util.Map;
import java.util.Optional;

public class LfuCacheStrategy implements CacheStrategy {

    private final Map<String, CacheEntry> cacheMap;

    public LfuCacheStrategy() {
        this.cacheMap = new HashMap<>();
    }

    @Override
    public EvictionPolicy getPolicy() {
        return EvictionPolicy.LFU;
    }

    @Override
    public void put(CacheEntry entry) {
        cacheMap.put(entry.getKey(), entry);
    }

    @Override
    public CacheEntry get(String key) {
        CacheEntry entry = cacheMap.get(key);
        if (entry != null) {
            entry.incrementAccessCount();
        }
        return entry;
    }

    @Override
    public CacheEntry remove(String key) {
        return cacheMap.remove(key);
    }

    @Override
    public void onAccess(CacheEntry entry) {
        entry.incrementAccessCount();
    }

    @Override
    public CacheEntry evict() {
        if (cacheMap.isEmpty()) {
            return null;
        }
        Optional<Map.Entry<String, CacheEntry>> minEntry = cacheMap.entrySet().stream()
                .min(Comparator.comparingInt(e -> e.getValue().getAccessCount()));
        if (minEntry.isPresent()) {
            String key = minEntry.get().getKey();
            cacheMap.remove(key);
            return minEntry.get().getValue();
        }
        return null;
    }

    @Override
    public Collection<CacheEntry> getAll() {
        return cacheMap.values();
    }

    @Override
    public int size() {
        return cacheMap.size();
    }

    @Override
    public void clear() {
        cacheMap.clear();
    }
}
