package com.example.cachemiddleware.cache;

import com.example.cachemiddleware.model.CacheEntry;
import com.example.cachemiddleware.model.EvictionPolicy;

import java.util.Collection;
import java.util.LinkedHashMap;
import java.util.Map;

public class LruCacheStrategy implements CacheStrategy {

    private final Map<String, CacheEntry> cacheMap;

    public LruCacheStrategy() {
        this.cacheMap = new LinkedHashMap<>(16, 0.75f, true);
    }

    @Override
    public EvictionPolicy getPolicy() {
        return EvictionPolicy.LRU;
    }

    @Override
    public void put(CacheEntry entry) {
        cacheMap.put(entry.getKey(), entry);
    }

    @Override
    public CacheEntry get(String key) {
        return cacheMap.get(key);
    }

    @Override
    public CacheEntry remove(String key) {
        return cacheMap.remove(key);
    }

    @Override
    public void onAccess(CacheEntry entry) {
        cacheMap.get(entry.getKey());
    }

    @Override
    public CacheEntry evict() {
        if (cacheMap.isEmpty()) {
            return null;
        }
        Map.Entry<String, CacheEntry> firstEntry = cacheMap.entrySet().iterator().next();
        cacheMap.remove(firstEntry.getKey());
        return firstEntry.getValue();
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
