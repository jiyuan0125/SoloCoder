package com.loadbalancer.config;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Configuration;

@Data
@Configuration
@ConfigurationProperties(prefix = "load-balancer.health-check")
public class LoadBalancerConfig {
    private int defaultIntervalSeconds = 10;
    private int defaultTimeoutSeconds = 3;
    private String path = "/health";
}