package com.poolguard.model;

import lombok.AllArgsConstructor;
import lombok.Data;

@Data
@AllArgsConstructor
public class PoolOverview {
    private String id;
    private String name;
    private PoolState state;
    private int activeConnections;
    private int idleConnections;
}
