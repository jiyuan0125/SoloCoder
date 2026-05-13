package com.servicemesh.registry.config;

import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

@Component
@ConfigurationProperties(prefix = "registry")
public class RegistryProperties {
    private long heartbeatInterval = 5000;
    private int maxMissedHeartbeats = 3;
    private int healthCheckTimeout = 3000;

    public long getHeartbeatInterval() { return heartbeatInterval; }
    public void setHeartbeatInterval(long heartbeatInterval) { this.heartbeatInterval = heartbeatInterval; }
    public int getMaxMissedHeartbeats() { return maxMissedHeartbeats; }
    public void setMaxMissedHeartbeats(int maxMissedHeartbeats) { this.maxMissedHeartbeats = maxMissedHeartbeats; }
    public int getHealthCheckTimeout() { return healthCheckTimeout; }
    public void setHealthCheckTimeout(int healthCheckTimeout) { this.healthCheckTimeout = healthCheckTimeout; }
}
