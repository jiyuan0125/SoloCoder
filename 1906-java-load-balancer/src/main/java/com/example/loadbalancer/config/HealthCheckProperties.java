package com.example.loadbalancer.config;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

@Data
@Component
@ConfigurationProperties(prefix = "load-balancer.health-check")
public class HealthCheckProperties {
    private int intervalMs = 5000;
    private int timeoutMs = 3000;
    private String path = "/health";
}
