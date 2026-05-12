package com.health.config;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

@Data
@Component
@ConfigurationProperties(prefix = "health.monitor")
public class HealthMonitorProperties {
    private Integer defaultIntervalSeconds = 30;
    private Integer defaultTimeoutMilliseconds = 5000;
    private Integer unhealthyThreshold = 3;
    private Integer historyRetentionMinutes = 1440;
}