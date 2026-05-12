package com.example.circuitbreaker.config;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

import java.util.HashMap;
import java.util.Map;

@Data
@Component
@ConfigurationProperties(prefix = "circuit-breaker")
public class CircuitBreakerProperties {
    private ServiceConfig defaultConfig = new ServiceConfig();
    private Map<String, ServiceConfig> services = new HashMap<>();

    @Data
    public static class ServiceConfig {
        private int failureThreshold = 5;
        private long openDuration = 30000;
        private int halfOpenPermittedCalls = 3;
    }

    public ServiceConfig getConfigForService(String serviceName) {
        return services.getOrDefault(serviceName, defaultConfig);
    }
}
