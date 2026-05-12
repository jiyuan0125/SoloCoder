package com.loadbalancer.model;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class BackendInstance {
    private String id;
    private String host;
    private int port;
    private int weight;
    private int activeConnections;
    private LocalDateTime registeredAt;
    private boolean healthy;

    public BackendInstance(String host, int port, int weight) {
        this.id = host + ":" + port;
        this.host = host;
        this.port = port;
        this.weight = weight;
        this.activeConnections = 0;
        this.registeredAt = LocalDateTime.now();
        this.healthy = true;
    }
}
