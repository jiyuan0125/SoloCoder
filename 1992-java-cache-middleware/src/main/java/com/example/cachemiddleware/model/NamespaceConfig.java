package com.example.cachemiddleware.model;

import lombok.Builder;
import lombok.Data;

@Data
@Builder
public class NamespaceConfig {
    private String namespace;
    private long ttlSeconds;
    private int maxCapacity;
    private EvictionPolicy evictionPolicy;
}
