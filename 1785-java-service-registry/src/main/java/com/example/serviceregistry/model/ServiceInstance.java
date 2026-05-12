package com.example.serviceregistry.model;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.Instant;
import java.util.Map;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class ServiceInstance {
    private String instanceId;
    private String serviceName;
    private String ip;
    private int port;
    private Map<String, String> tags;
    private int weight;
    private int heartbeatInterval;
    private Instant registeredAt;
    private Instant lastHeartbeatAt;
    private boolean healthy;
}
