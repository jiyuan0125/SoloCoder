package com.example.cachemiddleware.cache;

import com.example.cachemiddleware.model.CacheEntry;
import com.example.cachemiddleware.model.EvictionPolicy;

import java.util.Collection;

public interface CacheStrategy {
    EvictionPolicy getPolicy();

    void put(CacheEntry entry);

    CacheEntry get(String key);

    CacheEntry remove(String key);

    void onAccess(CacheEntry entry);

    CacheEntry evict();

    Collection<CacheEntry> getAll();

    int size();

    void clear();
}
