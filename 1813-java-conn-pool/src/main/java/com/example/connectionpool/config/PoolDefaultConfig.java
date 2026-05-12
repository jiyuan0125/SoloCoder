package com.example.connectionpool.config;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

@Component
@ConfigurationProperties(prefix = "pool.defaults")
@Data
public class PoolDefaultConfig {
    private int maxConnections = 20;
    private long waitTimeoutMs = 5000;
    private int leakThresholdSeconds = 60;
    private int maxQueueSize = 50;

    public ConnectionPoolConfig toConfig(String poolName) {
        ConnectionPoolConfig config = new ConnectionPoolConfig();
        config.setPoolName(poolName);
        config.setMaxConnections(maxConnections);
        config.setWaitTimeoutMs(waitTimeoutMs);
        config.setLeakThresholdSeconds(leakThresholdSeconds);
        config.setMaxQueueSize(maxQueueSize);
        return config;
    }
}
