package com.poolguard.model;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class PoolDetail {
    private String id;
    private String name;
    private int maxSize;
    private int minIdle;
    private int healthCheckInterval;
    private PoolState state;
    private int activeConnections;
    private int idleConnections;
    private LocalDateTime lastHealthCheckTime;
}
