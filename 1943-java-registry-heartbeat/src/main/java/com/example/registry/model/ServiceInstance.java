package com.example.registry.model;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.Instant;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class ServiceInstance {
    private String id;
    private String serviceName;
    private String ip;
    private int port;
    private int weight;
    private int heartbeatTtlSeconds;
    private HealthStatus healthStatus;
    private Instant lastHeartbeat;
    private Instant registerTime;
    private int missedHeartbeats;

    public enum HealthStatus {
        HEALTHY,
        UNHEALTHY
    }
}
