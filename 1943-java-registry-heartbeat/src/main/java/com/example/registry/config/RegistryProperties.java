package com.example.registry.config;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

@Data
@Component
@ConfigurationProperties(prefix = "registry")
public class RegistryProperties {
    private long heartbeatCheckIntervalMs = 5000;
    private int heartbeatMissedThreshold = 3;
    private int notificationRetryMax = 2;
    private long notificationRetryIntervalMs = 5000;
    private long notificationTimeoutMs = 3000;
}
