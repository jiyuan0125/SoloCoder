package com.depgraph.model;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class ServiceNode {
    private String name;
    private HealthStatus status = HealthStatus.HEALTHY;
    private LocalDateTime lastHeartbeat;
    private LocalDateTime registeredAt;

    public ServiceNode(String name) {
        this.name = name;
        this.registeredAt = LocalDateTime.now();
        this.lastHeartbeat = LocalDateTime.now();
    }

    public enum HealthStatus {
        HEALTHY, UNHEALTHY, UNKNOWN
    }
}
