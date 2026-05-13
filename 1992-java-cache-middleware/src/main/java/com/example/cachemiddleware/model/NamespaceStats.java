package com.example.cachemiddleware.model;

import lombok.Builder;
import lombok.Data;

@Data
@Builder
public class NamespaceStats {
    private String namespace;
    private long hitCount;
    private long missCount;
    private long evictionCount;
    private int currentCapacity;
    private long ttlSeconds;
    private int maxCapacity;
    private EvictionPolicy evictionPolicy;
}
