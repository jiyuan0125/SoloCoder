package com.example.pool.model;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class PoolInfo {
    
    private String name;
    private DataSourceType type;
    private PoolStatus status;
    private int maxConnections;
    private int minIdle;
    private int activeConnections;
    private int idleConnections;
    private int waitingRequests;
}
