package com.example.registry.model;

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
    private Map<String, String> metadata;
    private InstanceStatus status;
    private long leaseDuration;
    private Instant lastHeartbeat;
    private int heartbeatFailures;
    private Instant deregisterTime;
    
    public boolean isRecentlyDeregistered(long recoveryWindow) {
        if (deregisterTime == null || status != InstanceStatus.DEREGISTERED) {
            return false;
        }
        return Instant.now().minusMillis(recoveryWindow).isBefore(deregisterTime);
    }
}
